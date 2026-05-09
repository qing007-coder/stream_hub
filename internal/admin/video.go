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

	query := v.DB.Offset(offset).Limit(req.Size)

	if req.UserID != "" {
		query = query.Where("author_id = ?", req.UserID)
	}

	if req.IsPublic {
		query = query.Where("is_public = ?", constant.VideoPublic)
	}

	v.DB.Count(&total).Find(&videos)

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
		utils.BadRequest(ctx, "video  is not find, err:"+err.Error())
		return
	}

	if err := v.DB.Where("video_id = ?", id).First(&media).Error; err != nil {
		utils.BadRequest(ctx, "video  is not find, err:"+err.Error())
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
	
}