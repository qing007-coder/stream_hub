package gateway

import (
	"context"
	"stream_hub/internal/proto/video"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/metadata"
)

// @Summary 创建视频
// @Description 创建新的视频
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.CreateVideoRequest true "创建视频请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/create [post]
// CreateVideo 创建视频
func (g *Gateway) CreateVideo(ctx *gin.Context) {
	var req api.CreateVideoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.CreateVideoRequest{
		Title:           req.Title,
		Description:     req.Description,
		SourceObjectKey: req.SourceObjectKey,
		CoverUrl:        req.CoverURL,
	}

	resp, err := g.videoClient.CreateVideo(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.AuthorVideoInfo{
		ID:              resp.Id,
		Title:           resp.Title,
		Description:     resp.Description,
		CoverURL:        resp.CoverUrl,
		M3U8URL:         resp.M3U8Url,
		IsPublic:        resp.IsPublic,
		Duration:        resp.Duration,
		TranscodeStatus: resp.TranscodeStatus,
		AuditStatus:     resp.AuditStatus,
		CreatedAt:       resp.CreatedAt.AsTime(),
		UpdatedAt:       resp.UpdatedAt.AsTime(),
	}

	utils.StatusOK(ctx, apiResp, "Video created successfully")
}

// @Summary 获取视频详情
// @Description 根据视频 ID 获取视频详情
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/get/{video_id} [get]
// GetVideo 获取视频详情
func (g *Gateway) GetVideo(ctx *gin.Context) {
	var req api.GetVideoRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	grpcReq := &video.GetVideoRequest{
		VideoId: req.VideoID,
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	resp, err := g.videoClient.GetVideo(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.GetVideoResponse{
		ID:            resp.Id,
		Title:         resp.Title,
		CoverURL:      resp.CoverUrl,
		AuthorID:      resp.AuthorId,
		Duration:      resp.Duration,
		LikeCount:     resp.LikeCount,
		CommentCount:  resp.CommentCount,
		FavoriteCount: resp.FavoriteCount,
		ViewCount:     resp.ViewCount,
		M3U8URL:       resp.M3U8Url,
		CreatedAt:     resp.CreatedAt.AsTime(),
	}

	utils.StatusOK(ctx, apiResp, "Video retrieved successfully")
}


// @Summary 更新视频
// @Description 更新视频信息
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Param request body api.UpdateVideoRequest true "更新视频请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/update/{video_id} [put]
// UpdateVideo 更新视频
func (g *Gateway) UpdateVideo(ctx *gin.Context) {	var req api.UpdateVideoRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.UpdateVideoRequest{
		VideoId:     req.VideoID,
		Title:       req.Title,
		Description: req.Description,
		CoverUrl:    req.CoverURL,
		IsPublic:    req.IsPublic,
	}

	resp, err := g.videoClient.UpdateVideo(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.AuthorVideoInfo{
		ID:              resp.Id,
		Title:           resp.Title,
		Description:     resp.Description,
		CoverURL:        resp.CoverUrl,
		M3U8URL:         resp.M3U8Url,
		IsPublic:        resp.IsPublic,
		Duration:        resp.Duration,
		TranscodeStatus: resp.TranscodeStatus,
		AuditStatus:     resp.AuditStatus,
		CreatedAt:       resp.CreatedAt.AsTime(),
		UpdatedAt:       resp.UpdatedAt.AsTime(),
	}

	utils.StatusOK(ctx, apiResp, "Video updated successfully")
}

// @Summary 删除视频
// @Description 根据视频 ID 删除视频
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/delete/{video_id} [delete]
// DeleteVideo 删除视频
func (g *Gateway) DeleteVideo(ctx *gin.Context) {
	var req api.DeleteVideoRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.DeleteVideoRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.videoClient.DeleteVideo(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.DeleteVideoResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Video deleted successfully")
}

// @Summary 获取用户公开视频列表
// @Description 获取指定用户的公开视频列表
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "用户 ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/list/{user_id} [get]
// ListUserPublishedVideos 获取用户公开视频列表
func (g *Gateway) ListUserPublishedVideos(ctx *gin.Context) {
	var req api.ListUserPublishedVideosRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.ListUserPublishedVideosRequest{
		UserId: req.UserID,
		Page:   req.Page,
		Size:   req.Size,
	}

	resp, err := g.videoClient.ListUserPublishedVideos(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var videos []api.PublicVideoInfo
	for _, v := range resp.Videos {
		videos = append(videos, api.PublicVideoInfo{
			ID:        v.Id,
			Title:     v.Title,
			CoverURL:  v.CoverUrl,
			AuthorID:  v.AuthorId,
			Duration:  v.Duration,
			CreatedAt: v.CreatedAt.AsTime(),
		})
	}

	apiResp := api.ListUserPublishedVideosResponse{
		Videos: videos,
		Total:  resp.Total,
	}

	utils.StatusOK(ctx, apiResp, "Videos retrieved successfully")
}

// @Summary 获取我的视频列表
// @Description 获取当前用户的视频列表
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/my/list [get]
// ListMyVideos 获取我的视频列表
func (g *Gateway) ListMyVideos(ctx *gin.Context) {
	var req api.ListMyVideosRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.ListMyVideosRequest{
		Page: req.Page,
		Size: req.Size,
	}

	resp, err := g.videoClient.ListMyVideos(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var videos []api.AuthorVideoInfo
	for _, v := range resp.Videos {
		videos = append(videos, api.AuthorVideoInfo{
			ID:              v.Id,
			Title:           v.Title,
			Description:     v.Description,
			CoverURL:        v.CoverUrl,
			M3U8URL:         v.M3U8Url,
			IsPublic:        v.IsPublic,
			Duration:        v.Duration,
			LikeCount:       v.LikeCount,
			CommentCount:    v.CommentCount,
			FavoriteCount:   v.FavoriteCount,
			ViewCount:       v.ViewCount,
			TranscodeStatus: v.TranscodeStatus,
			AuditStatus:     v.AuditStatus,
			CreatedAt:       v.CreatedAt.AsTime(),
			UpdatedAt:       v.UpdatedAt.AsTime(),
		})
	}

	apiResp := api.ListMyVideosResponse{
		Videos: videos,
		Total:  resp.Total,
	}

	utils.StatusOK(ctx, apiResp, "Videos retrieved successfully")
}

// @Summary 获取视频Feed流
// @Description 获取关注用户的视频Feed流
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param last_id query string false "最后一个视频ID，用于分页"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/feed [get]
// ListFeedVideos 获取视频Feed流
func (g *Gateway) ListFeedVideos(ctx *gin.Context) {
	var req api.ListFeedVideosRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.ListFeedVideosRequest{
		Page:    req.Page,
		Size:    req.Size,
		LastId:  req.LastID,
	}

	resp, err := g.videoClient.ListFeedVideos(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var videos []api.FeedVideoInfo
	for _, v := range resp.Videos {
		videos = append(videos, api.FeedVideoInfo{
			ID:             v.Id,
			Title:          v.Title,
			CoverURL:       v.CoverUrl,
			M3U8URL:        v.M3U8Url,
			AuthorID:       v.AuthorId,
			AuthorNickname: v.AuthorNickname,
			AuthorAvatar:   v.AuthorAvatar,
			Duration:       v.Duration,
			LikeCount:      v.LikeCount,
			CommentCount:   v.CommentCount,
			FavoriteCount:  v.FavoriteCount,
			ViewCount:      v.ViewCount,
			CreatedAt:      v.CreatedAt.AsTime(),
		})
	}

	apiResp := api.ListFeedVideosResponse{
		Videos:  videos,
		Total:   resp.Total,
		HasMore: resp.HasMore,
	}

	utils.StatusOK(ctx, apiResp, "Feed videos retrieved successfully")
}

// @Summary 搜索视频
// @Description 根据关键词搜索视频
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/search [get]
// SearchVideos 搜索视频
func (g *Gateway) SearchVideos(ctx *gin.Context) {
	var req api.SearchVideosRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.SearchVideosRequest{
		Keyword: req.Keyword,
		Page:    req.Page,
		Size:    req.Size,
	}

	resp, err := g.videoClient.SearchVideos(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var videos []api.PublicVideoInfo
	for _, v := range resp.Videos {
		videos = append(videos, api.PublicVideoInfo{
			ID:        v.Id,
			Title:     v.Title,
			CoverURL:  v.CoverUrl,
			AuthorID:  v.AuthorId,
			Duration:  v.Duration,
			CreatedAt: v.CreatedAt.AsTime(),
		})
	}

	apiResp := api.SearchVideosResponse{
		Videos:  videos,
		Total:   resp.Total,
		HasMore: resp.HasMore,
	}

	utils.StatusOK(ctx, apiResp, "Videos searched successfully")
}

// @Summary 获取热度视频
// @Description 获取热度视频列表
// @Tags Video
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/video/hot [get]
// ListHotVideos 获取热度视频
func (g *Gateway) ListHotVideos(ctx *gin.Context) {
	var req api.ListHotVideosRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &video.ListHotVideosRequest{
		Page: req.Page,
		Size: req.Size,
	}

	resp, err := g.videoClient.ListHotVideos(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var videos []api.FeedVideoInfo
	for _, v := range resp.Videos {
		videos = append(videos, api.FeedVideoInfo{
			ID:             v.Id,
			Title:          v.Title,
			CoverURL:       v.CoverUrl,
			M3U8URL:        v.M3U8Url,
			AuthorID:       v.AuthorId,
			AuthorNickname: v.AuthorNickname,
			AuthorAvatar:   v.AuthorAvatar,
			Duration:       v.Duration,
			LikeCount:      v.LikeCount,
			CommentCount:   v.CommentCount,
			FavoriteCount:  v.FavoriteCount,
			ViewCount:      v.ViewCount,
			CreatedAt:      v.CreatedAt.AsTime(),
		})
	}

	apiResp := api.ListHotVideosResponse{
		Videos:  videos,
		Total:   resp.Total,
		HasMore: resp.HasMore,
	}

	utils.StatusOK(ctx, apiResp, "Hot videos retrieved successfully")
}

