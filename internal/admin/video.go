package admin

import (
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type VideoController struct {
	*infra.Base
}

func NewVideoController(base *infra.Base) *VideoController {
	return &VideoController{
		base,
	}
}

func (v *VideoController) GetVideoList(ctx *gin.Context) {
	var req api.GetVideoListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, "invalid query")
		return
	}

	offset := (req.Page - 1) * req.Size
	var videos []storage.VideoModel
	var total int64

	query := v.DB.Model(&storage.VideoModel{})

	if req.UserID != "" {
		query = query.Where("author_id = ?", req.UserID)
	}

	if req.IsPublic {
		query = query.Where("is_public = ?", constant.VideoPublic)
	}

	if err := query.Count(&total).Error; err != nil {
		utils.BadRequest(ctx, "count error: "+err.Error())
		return
	}

	if err := query.Offset(offset).Limit(req.Size).Order("created_at DESC").Find(&videos).Error; err != nil {
		utils.BadRequest(ctx, "query error: "+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  videos,
	}, "find successfully")
}

func (v *VideoController) UpdateVideo(ctx *gin.Context) {
	id := ctx.Param("id")

	var req api.UpdateVideoReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request: "+err.Error())
		return
	}

	var video storage.VideoModel
	if err := v.DB.Where("id = ?", id).First(&video).Error; err != nil {
		utils.BadRequest(ctx, "video not found: "+err.Error())
		return
	}

	if req.Title != "" {
		video.Title = req.Title
	}
	if req.Description != "" {
		video.Description = req.Description
	}
	if req.CoverURL != "" {
		video.CoverUrl = req.CoverURL
	}
	if req.IsPublic != 0 {
		video.IsPublic = req.IsPublic
	}

	if err := v.DB.Save(&video).Error; err != nil {
		utils.BadRequest(ctx, "update video failed: "+err.Error())
		return
	}

	var media storage.MediaModel
	if err := v.DB.Where("video_id = ?", id).First(&media).Error; err != nil {
		utils.BadRequest(ctx, "media not found: "+err.Error())
		return
	}

	if req.AuditStatus != 0 {
		media.AuditStatus = req.AuditStatus
		if err := v.DB.Save(&media).Error; err != nil {
			utils.BadRequest(ctx, "update media failed: "+err.Error())
			return
		}
	}

	utils.StatusOK(ctx, gin.H{
		"id":                video.ID,
		"title":             video.Title,
		"description":       video.Description,
		"author_id":         video.AuthorID,
		"cover_url":         video.CoverUrl,
		"is_public":         video.IsPublic,
		"duration":          video.Duration,
		"media_type":        media.Type,
		"source_object_key": media.SourceObjectKey,
		"m3u8_url":          media.M3u8Url,
		"transcode_status":  media.TranscodeStatus,
		"audit_status":      media.AuditStatus,
		"metadata":          media.Metadata,
	}, "更新成功")
}

func (v *VideoController) GetVideoDetail(ctx *gin.Context) {
	id := ctx.Param("id")

	var media storage.MediaModel
	var video storage.VideoModel

	if err := v.DB.Where("id = ?", id).First(&video).Error; err != nil {
		utils.BadRequest(ctx, "video is not find, err:"+err.Error())
		return
	}

	if err := v.DB.Where("video_id = ?", id).First(&media).Error; err != nil {
		utils.BadRequest(ctx, "media is not find, err:"+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id":                video.ID,
		"title":             video.Title,
		"description":       video.Description,
		"author_id":         video.AuthorID,
		"cover_url":         video.CoverUrl,
		"is_public":         video.IsPublic,
		"duration":          video.Duration,
		"created_at":        video.CreatedAt,
		"updated_at":        video.UpdatedAt,
		"media_type":        media.Type,
		"source_object_key": media.SourceObjectKey,
		"m3u8_url":          media.M3u8Url,
		"transcode_status":  media.TranscodeStatus,
		"audit_status":      media.AuditStatus,
		"media_metadata":    media.Metadata,
	}, "查询成功")
}

func (v *VideoController) GetVideoInteractionDetail(ctx *gin.Context) {
	id := ctx.Param("id")

	var likeCount int64
	var commentCount int64
	var favoriteCount int64

	v.DB.Model(&storage.VideoLikeModel{}).Where("video_id = ?", id).Count(&likeCount)
	v.DB.Model(&storage.VideoCommentModel{}).Where("video_id = ?", id).Count(&commentCount)
	v.DB.Model(&storage.VideoFavoriteModel{}).Where("video_id = ?", id).Count(&favoriteCount)

	var comments []storage.VideoCommentModel
	if err := v.DB.Where("video_id = ?", id).Order("created_at DESC").Limit(10).Find(&comments).Error; err != nil {
		utils.BadRequest(ctx, "query comments error: "+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"video_id":       id,
		"like_count":     likeCount,
		"comment_count":  commentCount,
		"favorite_count": favoriteCount,
		"recent_comments": comments,
	}, "查询成功")
}

func (v *VideoController) GetPendingAuditList(ctx *gin.Context) {
	var req api.GetVideoListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, "invalid query")
		return
	}

	offset := (req.Page - 1) * req.Size
	var videos []storage.VideoModel
	var mediaList []storage.MediaModel
	var total int64

	subQuery := v.DB.Model(&storage.MediaModel{}).Select("video_id").Where("audit_status = 0")

	query := v.DB.Model(&storage.VideoModel{}).Where("id IN (?)", subQuery)

	if err := query.Count(&total).Error; err != nil {
		utils.BadRequest(ctx, "count error: "+err.Error())
		return
	}

	if err := query.Offset(offset).Limit(req.Size).Order("created_at DESC").Find(&videos).Error; err != nil {
		utils.BadRequest(ctx, "query error: "+err.Error())
		return
	}

	videoIDs := make([]string, 0, len(videos))
	for _, video := range videos {
		videoIDs = append(videoIDs, video.ID)
	}

	if len(videoIDs) > 0 {
		v.DB.Where("video_id IN ?", videoIDs).Find(&mediaList)
	}

	mediaMap := make(map[string]storage.MediaModel)
	for _, media := range mediaList {
		mediaMap[media.VideoID] = media
	}

	result := make([]gin.H, 0, len(videos))
	for _, video := range videos {
		media := mediaMap[video.ID]
		result = append(result, gin.H{
			"id":                video.ID,
			"title":             video.Title,
			"author_id":         video.AuthorID,
			"cover_url":         video.CoverUrl,
			"duration":          video.Duration,
			"created_at":        video.CreatedAt,
			"transcode_status":  media.TranscodeStatus,
			"audit_status":      media.AuditStatus,
			"source_object_key": media.SourceObjectKey,
		})
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  result,
	}, "find successfully")
}

func (v *VideoController) AuditVideo(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		AuditStatus int    `json:"audit_status" binding:"required,oneof=5 6 7"`
		Remark      string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request: "+err.Error())
		return
	}

	var media storage.MediaModel
	if err := v.DB.Where("video_id = ?", id).First(&media).Error; err != nil {
		utils.BadRequest(ctx, "media not found: "+err.Error())
		return
	}

	media.AuditStatus = req.AuditStatus
	if err := v.DB.Save(&media).Error; err != nil {
		utils.BadRequest(ctx, "update audit status failed: "+err.Error())
		return
	}

	if req.AuditStatus == 5 {
		v.DB.Model(&storage.VideoModel{}).Where("id = ?", id).Update("is_public", constant.VideoPublic)
	} else if req.AuditStatus == 6 || req.AuditStatus == 7 {
		v.DB.Model(&storage.VideoModel{}).Where("id = ?", id).Update("is_public", constant.VideoPrivate)
	}

	utils.StatusOK(ctx, gin.H{
		"video_id":     id,
		"audit_status": req.AuditStatus,
		"remark":       req.Remark,
	}, "审核完成")
}
