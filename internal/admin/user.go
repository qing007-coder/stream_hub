package admin

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
		utils.BadRequest(ctx, err.Error())
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

func (u *UserController) InitRootAdmin() error {
	var count int64
	u.DB.Model(&storage.Admin{}).Count(&count)

	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := storage.Admin{
		Email:    "admin@streamhub.com",
		Name:     "Root Admin",
		Password: string(hashedPassword),
		Status:   1,
	}

	return u.DB.Create(&admin).Error
}

type CreateAdminReq struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=6"`
}

func (u *UserController) CreateAdmin(ctx *gin.Context) {
	var req CreateAdminReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	var existing storage.Admin
	if err := u.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		utils.BadRequest(ctx, "admin email already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	admin := storage.Admin{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
		Status:   1,
	}

	if err := u.DB.Create(&admin).Error; err != nil {
		utils.BadRequest(ctx, "create admin failed: "+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id":    admin.ID,
		"email": admin.Email,
		"name":  admin.Name,
	}, "创建成功")
}

func (u *UserController) GetAdminList(ctx *gin.Context) {
	var admins []storage.Admin
	var total int64

	page := 1
	size := 10

	if err := u.DB.Model(&storage.Admin{}).Count(&total).Error; err != nil {
		utils.BadRequest(ctx, "count error: "+err.Error())
		return
	}

	offset := (page - 1) * size
	if err := u.DB.Offset(offset).Limit(size).Order("created_at DESC").Find(&admins).Error; err != nil {
		utils.BadRequest(ctx, "query error: "+err.Error())
		return
	}

	result := make([]gin.H, 0, len(admins))
	for _, admin := range admins {
		result = append(result, gin.H{
			"id":         admin.ID,
			"email":      admin.Email,
			"name":       admin.Name,
			"status":     admin.Status,
			"created_at": admin.CreatedAt,
			"updated_at": admin.UpdatedAt,
		})
	}

	utils.StatusOK(ctx, gin.H{
		"total": total,
		"list":  result,
	}, "查询成功")
}

func (u *UserController) UpdateAdminStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var req struct {
		Status int8 `json:"status" binding:"required,oneof=0 1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body: "+err.Error())
		return
	}

	var admin storage.Admin
	if err := u.DB.Where("id = ?", id).First(&admin).Error; err != nil {
		utils.BadRequest(ctx, "admin not found: "+err.Error())
		return
	}

	if err := u.DB.Model(&storage.Admin{}).Where("id = ?", id).Update("status", req.Status).Error; err != nil {
		utils.BadRequest(ctx, "update error: "+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id":     id,
		"status": req.Status,
	}, "更新成功")
}

func (u *UserController) DeleteAdmin(ctx *gin.Context) {
	id := ctx.Param("id")

	var admin storage.Admin
	if err := u.DB.Where("id = ?", id).First(&admin).Error; err != nil {
		utils.BadRequest(ctx, "admin not found: "+err.Error())
		return
	}

	if err := u.DB.Delete(&admin).Error; err != nil {
		utils.BadRequest(ctx, "delete error: "+err.Error())
		return
	}

	utils.StatusOK(ctx, gin.H{
		"id": id,
	}, "删除成功")
}
