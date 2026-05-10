package video

import (
	"context"
	"errors"
	"fmt"
	"stream_hub/pkg/utils"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"stream_hub/internal/infra"
	"stream_hub/internal/proto/video"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
)

type Video struct {
	*infra.Base
	sender     *EventSender
	TaskSender *infra.TaskSender
}

func NewVideo(base *infra.Base, sender *EventSender, taskSender *infra.TaskSender) *Video {
	return &Video{base, sender, taskSender}
}

func (v *Video) CreateVideo(ctx context.Context, req *video.CreateVideoRequest, resp *video.AuthorVideoInfo) error {
	uid := ctx.Value("user_id").(string)

	model := storage.VideoModel{
		Title:       req.Title,
		Description: req.Description,
		AuthorID:    uid,
		CoverUrl:    req.CoverUrl,
	}

	if err := v.DB.Create(&model).Error; err != nil {
		return err
	}

	media := storage.MediaModel{
		VideoID:         model.ID,
		Type:            "original",
		SourceObjectKey: req.SourceObjectKey,
		TranscodeStatus: 0,
		AuditStatus:     0,
	}

	if err := v.DB.Create(&media).Error; err != nil {
		return err
	}

	v.fillAuthorVideoInfo(resp, &model, &media)

	eventType := ctx.Value("event_type").(string)
	resourceType := ctx.Value("resource_type").(string)

	v.sender.Send(&storage.Event{
		EventID:      utils.CreateID(),
		EventType:    eventType,
		UserID:       uid,
		ResourceType: resourceType,
		ResourceID:   model.ID,
		Timestamp:    time.Now().Unix(),
	})

	if err := v.TaskSender.SendTask(infra_.TaskMessage{
		Type:       constant.TaskVideoTranscode,
		BizID:      media.ID,
		Priority:   "critical",
		RetryCount: 0,
		Payload: infra_.TaskPayload{
			Operator: "",
			Source:   constant.Media,
			Data:     nil,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (v *Video) GetVideo(ctx context.Context, req *video.GetVideoRequest, resp *video.GetVideoResponse) error {
	uid := ctx.Value("user_id").(string)

	var model storage.VideoModel
	if err := v.DB.Where("id = ?", req.VideoId).First(&model).Error; err != nil {
		return err
	}

	var media storage.MediaModel
	if err := v.DB.Where("video_id = ?", req.VideoId).First(&media).Error; err != nil {
		return err
	}

	if uid == model.AuthorID {
		info := &video.AuthorVideoInfo{}
		v.fillAuthorVideoInfo(info, &model, &media)
		resp.Data = &video.GetVideoResponse_AuthorVideo{
			AuthorVideo: info,
		}
		return nil
	}

	if model.IsPublic == constant.VideoPublic && (media.AuditStatus == constant.VideoStatusMachinePassed || media.AuditStatus == constant.VideoStatusApproved) {
		info := &video.PublicVideoInfo{}
		v.fillPublicVideoInfo(info, &model, &media)
		resp.Data = &video.GetVideoResponse_PublicVideo{
			PublicVideo: info,
		}
		return nil
	}

	return errors.New("video is private or not audited")
}

func (v *Video) UpdateVideo(ctx context.Context, req *video.UpdateVideoRequest, resp *video.AuthorVideoInfo) error {
	uid := ctx.Value("user_id").(string)

	updates := map[string]interface{}{
		"title":       req.Title,
		"description": req.Description,
		"cover_url":   req.CoverUrl,
		"is_public":   req.IsPublic,
	}

	if err := v.DB.Model(&storage.VideoModel{}).
		Where("id = ? and author_id = ?", req.VideoId, uid).
		Updates(updates).Error; err != nil {
		return err
	}

	var model storage.VideoModel
	if err := v.DB.Where("id = ?", req.VideoId).First(&model).Error; err != nil {
		return err
	}

	var media storage.MediaModel
	if err := v.DB.Where("video_id = ?", req.VideoId).First(&media).Error; err != nil {
		return err
	}

	v.fillAuthorVideoInfo(resp, &model, &media)

	return nil
}

func (v *Video) DeleteVideo(ctx context.Context, req *video.DeleteVideoRequest, resp *video.DeleteVideoResponse) error {
	uid := ctx.Value("user_id").(string)

	if err := v.DB.
		Where("id = ? and author_id = ?", req.VideoId, uid).
		Delete(&storage.VideoModel{}).Error; err != nil {
		return err
	}

	v.Redis.Del(ctx)

	resp.Success = true
	resp.Message = "ok"

	return nil
}

func (v *Video) ListUserPublishedVideos(ctx context.Context, req *video.ListUserPublishedVideosRequest, resp *video.ListUserPublishedVideosResponse) error {
	var (
		list  []storage.VideoModel
		total int64
	)

	db := v.DB.Model(&storage.VideoModel{}).
		Where("author_id = ?", req.UserId).
		Where("is_public = ?", constant.VideoPublic)

	if err := db.Count(&total).Error; err != nil {
		return err
	}

	if err := db.
		Order("created_at desc").
		Limit(int(req.Size)).
		Offset(int((req.Page - 1) * req.Size)).
		Find(&list).Error; err != nil {
		return err
	}

	videoIDs := make([]string, 0, len(list))
	for _, item := range list {
		videoIDs = append(videoIDs, item.ID)
	}

	var medias []storage.MediaModel
	if len(videoIDs) > 0 {
		if err := v.DB.Where("video_id IN ?", videoIDs).Find(&medias).Error; err != nil {
			return err
		}
	}

	mediaMap := make(map[string]*storage.MediaModel)
	for i := range medias {
		if medias[i].AuditStatus == constant.VideoStatusMachinePassed || medias[i].AuditStatus == constant.VideoStatusApproved {
			mediaMap[medias[i].VideoID] = &medias[i]
		}
	}

	resp.Total = total
	resp.Videos = make([]*video.PublicVideoInfo, 0, len(list))

	for i := range list {
		media := mediaMap[list[i].ID]
		if media == nil {
			continue
		}
		info := &video.PublicVideoInfo{}
		v.fillPublicVideoInfo(info, &list[i], media)
		resp.Videos = append(resp.Videos, info)
	}

	return nil
}

func (v *Video) ListMyVideos(ctx context.Context, req *video.ListMyVideosRequest, resp *video.ListMyVideosResponse) error {
	uid := ctx.Value("user_id").(string)

	var (
		list  []storage.VideoModel
		total int64
	)

	db := v.DB.Model(&storage.VideoModel{}).
		Where("author_id = ?", uid)

	if err := db.Count(&total).Error; err != nil {
		return err
	}

	if err := db.
		Order("created_at desc").
		Limit(int(req.Size)).
		Offset(int((req.Page - 1) * req.Size)).
		Find(&list).Error; err != nil {
		return err
	}

	videoIDs := make([]string, 0, len(list))
	for _, item := range list {
		videoIDs = append(videoIDs, item.ID)
	}

	var medias []storage.MediaModel
	if len(videoIDs) > 0 {
		if err := v.DB.Where("video_id IN ?", videoIDs).Find(&medias).Error; err != nil {
			return err
		}
	}

	mediaMap := make(map[string]*storage.MediaModel)
	for i := range medias {
		mediaMap[medias[i].VideoID] = &medias[i]
	}

	resp.Total = total
	resp.Videos = make([]*video.AuthorVideoInfo, 0, len(list))

	for i := range list {
		info := &video.AuthorVideoInfo{}
		v.fillAuthorVideoInfo(info, &list[i], mediaMap[list[i].ID])
		resp.Videos = append(resp.Videos, info)
	}

	return nil
}

func (v *Video) fillPublicVideoInfo(resp *video.PublicVideoInfo, m *storage.VideoModel, media *storage.MediaModel) {
	resp.Id = m.ID
	resp.Title = m.Title
	resp.CoverUrl = m.CoverUrl
	resp.AuthorId = m.AuthorID
	resp.Duration = m.Duration
	resp.LikeCount = m.LikeCount
	resp.CommentCount = m.CommentCount
	resp.FavoriteCount = m.FavoriteCount
	resp.ViewCount = m.ViewCount
	resp.CreatedAt = timestamppb.New(m.CreatedAt)
	if media != nil {
		resp.M3U8Url = media.M3u8Url
	}
}

func (v *Video) fillAuthorVideoInfo(resp *video.AuthorVideoInfo, m *storage.VideoModel, media *storage.MediaModel) {
	resp.Id = m.ID
	resp.Title = m.Title
	resp.Description = m.Description
	resp.CoverUrl = m.CoverUrl
	resp.IsPublic = int32(m.IsPublic)
	resp.Duration = m.Duration
	resp.LikeCount = m.LikeCount
	resp.CommentCount = m.CommentCount
	resp.FavoriteCount = m.FavoriteCount
	resp.ViewCount = m.ViewCount
	resp.CreatedAt = timestamppb.New(m.CreatedAt)
	resp.UpdatedAt = timestamppb.New(m.UpdatedAt)
	if media != nil {
		resp.M3U8Url = media.M3u8Url
		resp.TranscodeStatus = int32(media.TranscodeStatus)
		resp.AuditStatus = int32(media.AuditStatus)
	}
}

func (v *Video) ListFeedVideos(ctx context.Context, req *video.ListFeedVideosRequest, resp *video.ListFeedVideosResponse) error {
	uid := ctx.Value("user_id").(string)

	recommendKey := fmt.Sprintf("recommend:users:%s", uid)
	recommendExists, err := v.Redis.Exists(ctx, recommendKey)
	if err != nil {
		return err
	}

	var videoList []storage.VideoModel

	if recommendExists {
		videoIDs, err := v.Redis.SMembers(ctx, recommendKey)
		if err != nil && err.Error() != "redis: nil" {
			return err
		}

		if len(videoIDs) > 0 {
			db := v.DB.Model(&storage.VideoModel{}).
				Where("id IN ?", videoIDs).
				Where("is_public = ?", constant.VideoPublic)

			if req.LastId != "" {
				var lastVideo storage.VideoModel
				if err := v.DB.Where("id = ?", req.LastId).First(&lastVideo).Error; err == nil {
					db = db.Where("created_at < ?", lastVideo.CreatedAt)
				}
			}

			err = db.Order("created_at desc").
				Limit(int(req.Size)).
				Find(&videoList).Error
			if err != nil {
				return err
			}
		}
	}

	if len(videoList) == 0 {
		db := v.DB.Model(&storage.VideoModel{}).
			Where("is_public = ?", constant.VideoPublic)

		if req.LastId != "" {
			var lastVideo storage.VideoModel
			if err := v.DB.Where("id = ?", req.LastId).First(&lastVideo).Error; err == nil {
				db = db.Where("created_at < ?", lastVideo.CreatedAt)
			}
		}

		err = db.Order("created_at desc").
			Limit(int(req.Size)).
			Find(&videoList).Error
		if err != nil {
			return err
		}
	}

	authorIDs := make([]string, 0, len(videoList))
	for _, vd := range videoList {
		authorIDs = append(authorIDs, vd.AuthorID)
	}

	var authors []storage.User
	if len(authorIDs) > 0 {
		if err := v.DB.Where("id IN ?", authorIDs).Find(&authors).Error; err != nil {
			return err
		}
	}

	authorMap := make(map[string]*storage.User)
	for _, author := range authors {
		authorMap[author.ID] = &author
	}

	videoIDs := make([]string, 0, len(videoList))
	for _, vd := range videoList {
		videoIDs = append(videoIDs, vd.ID)
	}

	var medias []storage.MediaModel
	if len(videoIDs) > 0 {
		if err := v.DB.Where("video_id IN ?", videoIDs).Find(&medias).Error; err != nil {
			return err
		}
	}

	mediaMap := make(map[string]*storage.MediaModel)
	for i := range medias {
		mediaMap[medias[i].VideoID] = &medias[i]
	}

	resp.Videos = make([]*video.FeedVideoInfo, 0, len(videoList))
	for _, vd := range videoList {
		author := authorMap[vd.AuthorID]
		media := mediaMap[vd.ID]
		if author == nil || media == nil || (media.AuditStatus != constant.VideoStatusMachinePassed && media.AuditStatus != constant.VideoStatusApproved) {
			continue
		}

		feedVideo := &video.FeedVideoInfo{
			Id:             vd.ID,
			Title:          vd.Title,
			CoverUrl:       vd.CoverUrl,
			AuthorId:       vd.AuthorID,
			AuthorNickname: author.Nickname,
			AuthorAvatar:   author.Avatar,
			Duration:       vd.Duration,
			LikeCount:      vd.LikeCount,
			CommentCount:   vd.CommentCount,
			CreatedAt:      timestamppb.New(vd.CreatedAt),
			M3U8Url:        media.M3u8Url,
		}
		resp.Videos = append(resp.Videos, feedVideo)
	}

	resp.Total = int64(len(videoList))
	resp.HasMore = len(videoList) >= int(req.Size)

	return nil
}

func (v *Video) SearchVideos(ctx context.Context, req *video.SearchVideosRequest, resp *video.SearchVideosResponse) error {
	var videoList []storage.VideoModel
	var total int64

	keyword := "%" + req.Keyword + "%"
	db := v.DB.Model(&storage.VideoModel{}).
		Where("title LIKE ? OR description LIKE ?", keyword, keyword).
		Where("is_public = ?", constant.VideoPublic)

	if err := db.Count(&total).Error; err != nil {
		return err
	}

	if err := db.Order("created_at desc").
		Limit(int(req.Size)).
		Offset(int((req.Page - 1) * req.Size)).
		Find(&videoList).Error; err != nil {
		return err
	}

	videoIDs := make([]string, 0, len(videoList))
	for _, vd := range videoList {
		videoIDs = append(videoIDs, vd.ID)
	}

	var medias []storage.MediaModel
	if len(videoIDs) > 0 {
		if err := v.DB.Where("video_id IN ?", videoIDs).Find(&medias).Error; err != nil {
			return err
		}
	}

	mediaMap := make(map[string]*storage.MediaModel)
	for i := range medias {
		if medias[i].AuditStatus == constant.VideoStatusMachinePassed || medias[i].AuditStatus == constant.VideoStatusApproved {
			mediaMap[medias[i].VideoID] = &medias[i]
		}
	}

	resp.Videos = make([]*video.PublicVideoInfo, 0, len(videoList))
	for _, vd := range videoList {
		media := mediaMap[vd.ID]
		if media == nil {
			continue
		}
		resp.Videos = append(resp.Videos, &video.PublicVideoInfo{
			Id:            vd.ID,
			Title:         vd.Title,
			CoverUrl:      vd.CoverUrl,
			AuthorId:      vd.AuthorID,
			Duration:      vd.Duration,
			LikeCount:     vd.LikeCount,
			CommentCount:  vd.CommentCount,
			FavoriteCount: vd.FavoriteCount,
			ViewCount:     vd.ViewCount,
			CreatedAt:     timestamppb.New(vd.CreatedAt),
			M3U8Url:       media.M3u8Url,
		})
	}

	resp.Total = total
	resp.HasMore = int64((int(req.Page)-1)*int(req.Size)+len(videoList)) < total

	return nil
}

func (v *Video) ListHotVideos(ctx context.Context, req *video.ListHotVideosRequest, resp *video.ListHotVideosResponse) error {
	var videoList []storage.VideoModel
	db := v.DB.Model(&storage.VideoModel{}).
		Where("is_public = ?", constant.VideoPublic)

	if err := db.Order("created_at desc").
		Limit(int(req.Size)).
		Offset(int((req.Page - 1) * req.Size)).
		Find(&videoList).Error; err != nil {
		return err
	}

	authorIDs := make([]string, 0, len(videoList))
	for _, vd := range videoList {
		authorIDs = append(authorIDs, vd.AuthorID)
	}

	var authors []storage.User
	if len(authorIDs) > 0 {
		if err := v.DB.Where("id IN ?", authorIDs).Find(&authors).Error; err != nil {
			return err
		}
	}

	authorMap := make(map[string]*storage.User)
	for _, author := range authors {
		authorMap[author.ID] = &author
	}

	videoIDs := make([]string, 0, len(videoList))
	for _, vd := range videoList {
		videoIDs = append(videoIDs, vd.ID)
	}

	var medias []storage.MediaModel
	if len(videoIDs) > 0 {
		if err := v.DB.Where("video_id IN ?", videoIDs).Find(&medias).Error; err != nil {
			return err
		}
	}

	mediaMap := make(map[string]*storage.MediaModel)
	for i := range medias {
		if medias[i].AuditStatus == constant.VideoStatusMachinePassed || medias[i].AuditStatus == constant.VideoStatusApproved {
			mediaMap[medias[i].VideoID] = &medias[i]
		}
	}

	resp.Videos = make([]*video.FeedVideoInfo, 0, len(videoList))
	for _, vd := range videoList {
		author := authorMap[vd.AuthorID]
		media := mediaMap[vd.ID]
		if author == nil || media == nil {
			continue
		}

		resp.Videos = append(resp.Videos, &video.FeedVideoInfo{
			Id:             vd.ID,
			Title:          vd.Title,
			CoverUrl:       vd.CoverUrl,
			AuthorId:       vd.AuthorID,
			AuthorNickname: author.Nickname,
			AuthorAvatar:   author.Avatar,
			Duration:       vd.Duration,
			LikeCount:      vd.LikeCount,
			CommentCount:   vd.CommentCount,
			CreatedAt:      timestamppb.New(vd.CreatedAt),
			M3U8Url:        media.M3u8Url,
		})
	}

	resp.Total = int64(len(videoList))
	resp.HasMore = len(videoList) >= int(req.Size)

	return nil
}
