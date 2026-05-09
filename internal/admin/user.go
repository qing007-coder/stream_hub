package admin

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	*infra.Base
}

func NewUserController(base *infra.Base) *UserController {
	return &UserController{
		base,
	}
}

func (u *UserController) GetUserList(ctx *gin.Context) {
	var req api.GetUserListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.BadRequest(ctx, "invalid param")
		return
	}

	var users []storage.User
	var total int64

	offset := (req.Page - 1) * req.Size
	query := u.DB.Model(&storage.User{}).Offset(offset).Limit(req.Size)

	if req.UserID != "" {
		query = query.Where("id = ?", req.UserID)
	}

	if req.Nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("count error: %s", err.Error()))
		return
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("query error: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  users,
	}, "查询成功")
}

func (u *UserController) GetUserDetail(ctx *gin.Context) {
	id := ctx.Param("id")

	var user storage.User
	if err := u.DB.Where("id = ?", id).First(&user).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("user not found: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"nickname":       user.Nickname,
		"avatar":         user.Avatar,
		"background_url": user.BackgroundURL,
		"signature":      user.Signature,
		"gender":         user.Gender,
		"tags":           user.Tags,
		"like_count":     user.LikeCount,
		"follow_count":   user.FollowCount,
		"follower_count": user.FollowerCount,
		"work_count":     user.WorkCount,
		"favorite_count": user.FavoriteCount,
		"status":         user.Status,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
	}, "查询成功")
}

func (u *UserController) UpdateUserStatus(ctx *gin.Context) {
	var req api.UpdateUserStatusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	var user storage.User
	if err := u.DB.Where("id = ?", req.TargetUserID).First(&user).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("user not found: %s", err.Error()))
		return
	}

	if err := u.DB.Model(&storage.User{}).Where("id = ?", req.TargetUserID).Update("status", req.Status).Error; err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("update error: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"user_id": req.TargetUserID,
		"status":  req.Status,
	}, "更新成功")
}
