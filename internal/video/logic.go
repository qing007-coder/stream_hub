package video

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"

	"stream_hub/internal/infra"
	"stream_hub/internal/proto/video"
	"stream_hub/pkg/constant"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
)

type Video struct {
	*infra.Base
}

func NewVideo(base *infra.Base) *Video {
	return &Video{base}
}

func (v *Video) CreateVideo(ctx context.Context, req *video.CreateVideoRequest, resp *video.AuthorVideoInfo) error {

	uid := ctx.Value("user_id").(string)

	model := storage.VideoModel{
		Title:           req.Title,
		Description:     req.Description,
		AuthorID:        uid,
		SourceObjectKey: req.SourceObjectKey,
		CoverUrl:        req.CoverUrl,
	}

	if err := v.DB.Create(&model).Error; err != nil {
		return err
	}

	v.fillAuthorVideoInfo(resp, &model)

	return v.TaskSender.SendTask(infra_.TaskMessage{
		Type:  constant.TaskVideoToES,
		BizID: model.ID,
	})
}

func (v *Video) GetVideo(ctx context.Context, req *video.GetVideoRequest, resp *video.GetVideoResponse) error {

	uid := ctx.Value("user_id").(string)

	var model storage.VideoModel
	if err := v.DB.Where("id = ?", req.VideoId).First(&model).Error; err != nil {
		return err
	}

	// 作者视角
	if uid == model.AuthorID {
		info := &video.AuthorVideoInfo{}
		v.fillAuthorVideoInfo(info, &model)
		resp.Data = &video.GetVideoResponse_AuthorVideo{
			AuthorVideo: info,
		}
		return nil
	}

	// 访客视角
	if model.IsPublic == constant.VideoPublic &&
		model.Status == constant.VideoApproved {

		info := &video.PublicVideoInfo{}
		v.fillPublicVideoInfo(info, &model)
		resp.Data = &video.GetVideoResponse_PublicVideo{
			PublicVideo: info,
		}
		return nil
	}

	return errors.New("video is private or not approved")
}

func (v *Video) UpdateVideo(ctx context.Context, req *video.UpdateVideoRequest, resp *video.AuthorVideoInfo) error {

	uid := ctx.Value("user_id").(string)

	if err := v.DB.Model(&storage.VideoModel{}).
		Where("id = ? and author_id = ?", req.VideoId, uid).
		Updates(map[string]interface{}{
			"title":       req.Title,
			"description": req.Description,
			"cover_url":   req.CoverUrl,
			"is_public":   req.IsPublic,
		}).Error; err != nil {
		return err
	}

	var model storage.VideoModel
	if err := v.DB.Where("id = ?", req.VideoId).First(&model).Error; err != nil {
		return err
	}

	v.fillAuthorVideoInfo(resp, &model)

	return v.TaskSender.SendTask(infra_.TaskMessage{
		Type:  constant.TaskVideoToES,
		BizID: model.ID,
	})
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

	return v.TaskSender.SendTask(infra_.TaskMessage{
		Type:  constant.TaskVideoToES,
		BizID: req.VideoId,
	})
}

func (v *Video) ListUserPublishedVideos(ctx context.Context, req *video.ListUserPublishedVideosRequest, resp *video.ListUserPublishedVideosResponse) error {

	var (
		list  []storage.VideoModel
		total int64
	)

	db := v.DB.Model(&storage.VideoModel{}).
		Where("author_id = ?", req.UserId).
		Where("is_public = ?", constant.VideoPublic).
		Where("status = ?", constant.VideoApproved)

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

	resp.Total = total
	resp.Videos = make([]*video.PublicVideoInfo, 0, len(list))

	for i := range list {
		info := &video.PublicVideoInfo{}
		v.fillPublicVideoInfo(info, &list[i])
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

	resp.Total = total
	resp.Videos = make([]*video.AuthorVideoInfo, 0, len(list))

	for i := range list {
		info := &video.AuthorVideoInfo{}
		v.fillAuthorVideoInfo(info, &list[i])
		resp.Videos = append(resp.Videos, info)
	}

	return nil
}

func (v *Video) fillPublicVideoInfo(resp *video.PublicVideoInfo, m *storage.VideoModel) {
	resp.Id = m.ID
	resp.Title = m.Title
	resp.CoverUrl = m.CoverUrl
	resp.AuthorId = m.AuthorID
	resp.Duration = m.Duration
	resp.CreatedAt = timestamppb.New(m.CreatedAt)
}

func (v *Video) fillAuthorVideoInfo(resp *video.AuthorVideoInfo, m *storage.VideoModel) {
	resp.Id = m.ID
	resp.Title = m.Title
	resp.Description = m.Description
	resp.CoverUrl = m.CoverUrl
	resp.Status = int32(m.Status)
	resp.IsPublic = int32(m.IsPublic)
	resp.Duration = m.Duration
	resp.CreatedAt = timestamppb.New(m.CreatedAt)
	resp.UpdatedAt = timestamppb.New(m.UpdatedAt)
}
