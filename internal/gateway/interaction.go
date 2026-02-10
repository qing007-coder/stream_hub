package gateway

import (
	"context"
	"stream_hub/internal/proto/interaction"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4/metadata"
)

// @Summary 点赞
// @Description 为视频点赞
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/like/{video_id} [post]
// CreateLike 点赞
func (g *Gateway) CreateLike(ctx *gin.Context) {
	var req api.LikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.LikeRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.CreateLike(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Like created successfully")
}

// @Summary 取消点赞
// @Description 取消对视频的点赞
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/like/{video_id} [delete]
// DeleteLike 取消点赞
func (g *Gateway) DeleteLike(ctx *gin.Context) {
	var req api.LikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.LikeRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.DeleteLike(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Like deleted successfully")
}

// @Summary 是否点赞
// @Description 检查用户是否为视频点赞
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/like/{video_id} [get]
// IsLike 是否点赞
func (g *Gateway) IsLike(ctx *gin.Context) {
	var req api.IsLikeRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.IsLikeRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.IsLike(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.IsLikeResponse{
		IsLike: resp.IsLike,
	}

	utils.StatusOK(ctx, apiResp, "Like status retrieved successfully")
}

// @Summary 获取点赞列表
// @Description 获取视频的点赞列表
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/likes/{video_id} [get]
// ListLikes 获取点赞列表
func (g *Gateway) ListLikes(ctx *gin.Context) {
	var req api.ListLikesRequest
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

	grpcReq := &interaction.ListLikesRequest{
		VideoId: req.VideoID,
		Page:    req.Page,
		Size:    req.Size,
	}

	resp, err := g.interactionClient.ListLikes(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var users []api.UserView
	for _, u := range resp.Users {
		users = append(users, api.UserView{
			UserID:    u.UserId,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	apiResp := api.ListLikesResponse{
		Users: users,
		Total: resp.Total,
	}

	utils.StatusOK(ctx, apiResp, "Likes retrieved successfully")
}

// @Summary 收藏
// @Description 收藏视频
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/favorite/{video_id} [post]
// CreateFavorite 收藏
func (g *Gateway) CreateFavorite(ctx *gin.Context) {
	var req api.FavoriteRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.FavoriteRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.CreateFavorite(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Favorite created successfully")
}

// @Summary 取消收藏
// @Description 取消收藏视频
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/favorite/{video_id} [delete]
// DeleteFavorite 取消收藏
func (g *Gateway) DeleteFavorite(ctx *gin.Context) {
	var req api.FavoriteRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.FavoriteRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.DeleteFavorite(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Favorite deleted successfully")
}

// @Summary 是否收藏
// @Description 检查用户是否收藏视频
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/favorite/{video_id} [get]
// IsFavorite 是否收藏
func (g *Gateway) IsFavorite(ctx *gin.Context) {
	var req api.IsFavoriteRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.IsFavoriteRequest{
		VideoId: req.VideoID,
	}

	resp, err := g.interactionClient.IsFavorite(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.IsFavoriteResponse{
		IsFavorite: resp.IsFavorite,
	}

	utils.StatusOK(ctx, apiResp, "Favorite status retrieved successfully")
}

// @Summary 关注
// @Description 关注用户
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_user_id path string true "目标用户 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/follow/{target_user_id} [post]
// CreateFollow 关注
func (g *Gateway) CreateFollow(ctx *gin.Context) {
	var req api.FollowRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.FollowRequest{
		TargetUserId: req.TargetUserID,
	}

	resp, err := g.interactionClient.CreateFollow(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Follow created successfully")
}

// @Summary 取消关注
// @Description 取消关注用户
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_user_id path string true "目标用户 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/follow/{target_user_id} [delete]
// DeleteFollow 取消关注
func (g *Gateway) DeleteFollow(ctx *gin.Context) {
	var req api.FollowRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.FollowRequest{
		TargetUserId: req.TargetUserID,
	}

	resp, err := g.interactionClient.DeleteFollow(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Follow deleted successfully")
}

// @Summary 是否关注
// @Description 检查用户是否关注目标用户
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param target_user_id path string true "目标用户 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/follow/{target_user_id} [get]
// IsFollow 是否关注
func (g *Gateway) IsFollow(ctx *gin.Context) {
	var req api.IsFollowRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.IsFollowRequest{
		TargetUserId: req.TargetUserID,
	}

	resp, err := g.interactionClient.IsFollow(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.IsFollowResponse{
		IsFollow: resp.IsFollow,
	}

	utils.StatusOK(ctx, apiResp, "Follow status retrieved successfully")
}

// @Summary 获取粉丝列表
// @Description 获取用户的粉丝列表
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "用户 ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/followers/{user_id} [get]
// ListFollowers 获取粉丝列表
func (g *Gateway) ListFollowers(ctx *gin.Context) {
	var req api.ListFollowersRequest
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

	grpcReq := &interaction.ListFollowersRequest{
		UserId: req.UserID,
		Page:   req.Page,
		Size:   req.Size,
	}

	resp, err := g.interactionClient.ListFollowers(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var users []api.UserView
	for _, u := range resp.Users {
		users = append(users, api.UserView{
			UserID:    u.UserId,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	apiResp := api.ListFollowersResponse{
		Users: users,
		Total: resp.Total,
	}

	utils.StatusOK(ctx, apiResp, "Followers retrieved successfully")
}

// @Summary 获取关注列表
// @Description 获取用户的关注列表
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "用户 ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/followings/{user_id} [get]
// ListFollowings 获取关注列表
func (g *Gateway) ListFollowings(ctx *gin.Context) {
	var req api.ListFollowingsRequest
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

	grpcReq := &interaction.ListFollowingsRequest{
		UserId: req.UserID,
		Page:   req.Page,
		Size:   req.Size,
	}

	resp, err := g.interactionClient.ListFollowings(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var users []api.UserView
	for _, u := range resp.Users {
		users = append(users, api.UserView{
			UserID:    u.UserId,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	apiResp := api.ListFollowingsResponse{
		Users: users,
		Total: resp.Total,
	}

	utils.StatusOK(ctx, apiResp, "Followings retrieved successfully")
}

// @Summary 创建评论
// @Description 为视频创建评论
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body api.CreateCommentRequest true "创建评论请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/comment [post]
// CreateComment 创建评论
func (g *Gateway) CreateComment(ctx *gin.Context) {
	var req api.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.CreateCommentRequest{
		VideoId:       req.VideoID,
		Content:       req.Content,
		ParentId:      req.ParentID,
		ReplyToUserId: req.ReplyToUserID,
	}

	resp, err := g.interactionClient.CreateComment(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.Comment{
		ID:            resp.Id,
		VideoID:       resp.VideoId,
		UserID:        resp.UserId,
		Nickname:      resp.Nickname,
		Avatar:        resp.Avatar,
		Content:       resp.Content,
		ParentID:      resp.ParentId,
		ReplyToUserID: resp.ReplyToUserId,
		LikeCount:     resp.LikeCount,
		ReplyCount:    resp.ReplyCount,
		CreateTime:    resp.CreateTime,
	}

	utils.StatusOK(ctx, apiResp, "Comment created successfully")
}

// @Summary 删除评论
// @Description 删除评论
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param comment_id path string true "评论 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/comment/{comment_id} [delete]
// DeleteComment 删除评论
func (g *Gateway) DeleteComment(ctx *gin.Context) {
	var req api.DeleteCommentRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	userID := ctx.GetString("user_id")
	ctxWithMetadata := metadata.NewContext(context.Background(), map[string]string{
		"user_id": userID,
	})

	grpcReq := &interaction.DeleteCommentRequest{
		CommentId: req.CommentID,
	}

	resp, err := g.interactionClient.DeleteComment(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	apiResp := api.ActionResponse{
		Success: resp.Success,
		Message: resp.Message,
	}

	utils.StatusOK(ctx, apiResp, "Comment deleted successfully")
}

// @Summary 获取评论列表
// @Description 获取视频的评论列表
// @Tags Interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param video_id path string true "视频 ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param parent_id query string false "父评论 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Failure 400 {object} map[string]interface{} "请求错误"
// @Router /api/interaction/comments/{video_id} [get]
// ListComments 获取评论列表
func (g *Gateway) ListComments(ctx *gin.Context) {
	var req api.ListCommentsRequest
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

	grpcReq := &interaction.ListCommentsRequest{
		VideoId:  req.VideoID,
		Page:     req.Page,
		PageSize: req.PageSize,
		ParentId: req.ParentID,
	}

	resp, err := g.interactionClient.ListComments(ctxWithMetadata, grpcReq)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var comments []api.Comment
	for _, c := range resp.Comments {
		comments = append(comments, api.Comment{
			ID:            c.Id,
			VideoID:       c.VideoId,
			UserID:        c.UserId,
			Nickname:      c.Nickname,
			Avatar:        c.Avatar,
			Content:       c.Content,
			ParentID:      c.ParentId,
			ReplyToUserID: c.ReplyToUserId,
			LikeCount:     c.LikeCount,
			ReplyCount:    c.ReplyCount,
			CreateTime:    c.CreateTime,
		})
	}

	apiResp := api.ListCommentsResponse{
		Comments: comments,
		Total:    resp.Total,
		HasMore:  resp.HasMore,
	}

	utils.StatusOK(ctx, apiResp, "Comments retrieved successfully")
}
