package admin

import (
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/model/auth"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	*infra.Base
	auth *security.Auth
}

func NewAuthController(base *infra.Base, auth *security.Auth) *AuthController {
	return &AuthController{
		Base: base,
		auth: auth,
	}
}

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (a *AuthController) Login(ctx *gin.Context) {
	var req LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, "invalid request body")
		return
	}

	var admin storage.Admin
	if err := a.DB.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		utils.BadRequest(ctx, "email or password incorrect")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		utils.BadRequest(ctx, "email or password incorrect")
		return
	}

	if admin.Status != 1 {
		utils.ForbiddenRequest(ctx, "account disabled")
		return
	}

	claims := &auth.Claims{
		UserID:    admin.ID,
		Role:      "admin",
		CreatedAt: time.Now().Unix(),
	}

	tokenMap, err := a.auth.GenerateToken(claims)
	if err != nil {
		utils.InternalServerError(ctx)
		return
	}

	utils.StatusOK(ctx, gin.H{
		"access_token":  tokenMap["access_token"],
		"refresh_token": tokenMap["refresh_token"],
		"admin": gin.H{
			"id":    admin.ID,
			"email": admin.Email,
			"name":  admin.Name,
		},
	}, "login success")
}

func (a *AuthController) RefreshToken(ctx *gin.Context) {
	refreshToken := ctx.GetHeader("Refresh-Token")
	if refreshToken == "" {
		utils.BadRequest(ctx, "need refresh token")
		return
	}

	tokenMap, err := a.auth.RefreshToken(refreshToken)
	if err != nil {
		utils.BadRequest(ctx, fmt.Sprintf("refresh token failed: %s", err.Error()))
		return
	}

	utils.StatusOK(ctx, gin.H{
		"access_token":  tokenMap["access_token"],
		"refresh_token": tokenMap["refresh_token"],
	}, "refresh token success")
}

func (a *AuthController) Logout(ctx *gin.Context) {
	refreshToken := ctx.GetHeader("Refresh-Token")
	if refreshToken != "" {
		a.Redis.Del(ctx, refreshToken)
	}

	utils.StatusOK(ctx, gin.H{}, "logout success")
}
