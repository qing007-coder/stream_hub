package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"
	"stream_hub/internal/infra"
	"stream_hub/internal/security"
	"stream_hub/pkg/constant"
	errors_ "stream_hub/pkg/errors"
	"stream_hub/pkg/model/api"
	"stream_hub/pkg/model/auth"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"
	"strings"
	"time"
)

type UserApi struct {
	*infra.Base
	auth *security.Auth
}

func NewUserApi(base *infra.Base, auth *security.Auth) *UserApi {

	return &UserApi{
		Base: base,
		auth: auth,
	}
}

func (u *UserApi) Login(ctx *gin.Context) {
	var req api.LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	var user storage.User
	if err := u.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	if !utils.ComparePassword(user.Password, req.Password) {
		utils.BadRequest(ctx, "account or password is wrong")
		return
	}

	switch user.Status {
	case 2:
		utils.BadRequest(ctx, "your account has been banned")
		return
	case 3:
		utils.BadRequest(ctx, "your account has been cancellation")
		return
	}

	token, err := u.auth.GenerateToken(&auth.Claims{
		UserID:    user.ID,
		Role:      constant.RoleUser,
		CreatedAt: time.Now().Unix(),
	})

	if err != nil {
		utils.BadRequest(ctx, "token generate failed")
		return
	}

	utils.StatusOK(ctx, token, "log in successfully")
}

func (u *UserApi) Logout(ctx *gin.Context) {
	uid := ctx.GetString("user_id")
	data, err := u.Redis.Get(context.Background(), uid)
	if err != nil {
		utils.BadRequest(ctx, "redis get failed")
		return
	}

	if err := u.Redis.Del(context.Background(), string(data), uid); err != nil {
		utils.BadRequest(ctx, "redis del failed")
		return
	}

	utils.StatusOK(ctx, nil, "log out successfully")
}

func (u *UserApi) Register(ctx *gin.Context) {
	var req api.RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	var count int64
	u.DB.Model(&storage.User{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		utils.BadRequest(ctx, "邮箱已注册")
		return
	}

	data, err := u.Redis.Get(context.Background(), req.Email)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}
	if string(data) != req.VerificationCode {
		utils.BadRequest(ctx, "验证码错误")
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	user := storage.User{
		Password: string(password),
		Email:    req.Email,
		Nickname: "匿名用户" + utils.CreateID(),
	}
	u.DB.Create(&user)

	// 给上用户权限 ..................
	fmt.Println(user)

	token, err := u.auth.GenerateToken(&auth.Claims{
		UserID:    user.ID,
		Role:      constant.RoleUser, // 这里的权限后面会改
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	utils.StatusOK(ctx, token, "注册成功")
}

func (u *UserApi) RefreshToken(ctx *gin.Context) {
	var req api.RefreshTokenReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	token, err := u.auth.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, errors_.RefreshTokenExpiredError) {
			utils.BadRequest(ctx, "refresh token is expired")
			return
		} else {
			utils.BadRequest(ctx, "token refresh failed")
			return
		}
	}

	utils.StatusOK(ctx, token, "refresh token successfully")
}

func (u *UserApi) SendVerificationCode(ctx *gin.Context) {
	var req api.SendVerifyCodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}
	if err := u.TaskSender.SendTask(infra_.TaskMessage{
		Type: constant.TaskSendEmailCode,
		BizID: req.Email,
		Priority: "critical",
		RetryCount: 0,
		Payload: infra_.TaskPayload{
			Operator: "",
			Source: constant.User,
			Data: nil,
		},
	}); err != nil {
		utils.BadRequest(ctx, "send task failed")
		return
	}

	utils.StatusOK(ctx, nil, "send verification code successfully")
}

func (u *UserApi) UpdateProfile(ctx *gin.Context) {
	var req api.UpdateUserProfileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	uid := ctx.GetString("user_id")
	var user storage.User
	u.DB.Where("id = ?", uid).First(&user)

	if user.BackgroundURL != req.BackgroundUrl && user.BackgroundURL != "" {
		path := strings.SplitN(strings.TrimPrefix(user.BackgroundURL, "/"), "/", 2)
		bucket, object := path[0], path[1]
		if err := u.Minio.Client.RemoveObject(context.Background(), bucket, object, minio.RemoveObjectOptions{}); err != nil {
			utils.BadRequest(ctx, "remove bucket object failed")
			return
		}
	}

	if user.Avatar != req.AvatarUrl && user.Avatar != "" {
		path := strings.SplitN(strings.TrimPrefix(user.Avatar, "/"), "/", 2)
		bucket, object := path[0], path[1]
		if err := u.Minio.Client.RemoveObject(context.Background(), bucket, object, minio.RemoveObjectOptions{}); err != nil {
			utils.BadRequest(ctx, "remove bucket object failed")
			return
		}
	}

	if err := u.DB.Model(&storage.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"nickname":       req.Nickname,
		"avatar":         req.AvatarUrl,
		"background_url": req.BackgroundUrl,
		"gender":         req.Gender,
		"signature":      req.Signature,
	}).Error; err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}

	utils.StatusOK(ctx, nil, "update user successfully")
}

func (u *UserApi) UpdatePassword(ctx *gin.Context) {
	var req api.UpdatePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(ctx, utils.MessageInvalidBody)
		return
	}

	uid := ctx.GetString("user_id")
	var user storage.User
	u.DB.Where("id = ?", uid).First(&user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		utils.BadRequest(ctx, "旧密码错误")
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.BadRequest(ctx, err.Error())
		return
	}
	u.DB.Model(&storage.User{}).Where("id = ?", uid).Update("password", password)
	utils.StatusOK(ctx, nil, "修改成功")
}

func (u *UserApi) GetUserProfile(ctx *gin.Context) {
	uid := ctx.GetString("user_id")
	var user storage.User
	u.DB.Where("id = ?", uid).First(&user)

	utils.StatusOK(ctx, api.GetUserProfileResp{
		Email:          user.Email,
		Nickname:       user.Nickname,
		BackgroundUrl:  user.BackgroundURL,
		AvatarUrl:      user.Avatar,
		Gender:         user.Gender,
		Signature:      user.Signature,
		FollowCount:    user.FollowCount,
		FollowerCount:  user.FollowerCount,
		WorkCount:      user.WorkCount,
		FavouriteCount: user.FavoriteCount,
	}, "get user successfully")
}
