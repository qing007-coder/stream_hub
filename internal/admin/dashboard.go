package admin

import (
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/storage"
	"stream_hub/pkg/utils"

	"github.com/gin-gonic/gin"
)

type ApiController struct {
	*infra.Base
}

func NewApiController(base *infra.Base) *ApiController {
	return &ApiController{Base: base}
}

func (a *ApiController) GetDashboardStats(ctx *gin.Context) {
	var userCount int64
	var videoCount int64
	var pendingAuditCount int64
	var todayUserCount int64
	var todayVideoCount int64

	a.DB.Model(&storage.User{}).Count(&userCount)
	a.DB.Model(&storage.VideoModel{}).Count(&videoCount)
	a.DB.Model(&storage.MediaModel{}).Where("audit_status = 0").Count(&pendingAuditCount)

	a.DB.Model(&storage.User{}).Where("DATE(created_at) = CURDATE()").Count(&todayUserCount)
	a.DB.Model(&storage.VideoModel{}).Where("DATE(created_at) = CURDATE()").Count(&todayVideoCount)

	utils.StatusOK(ctx, gin.H{
		"user_count":          userCount,
		"video_count":         videoCount,
		"pending_audit_count": pendingAuditCount,
		"today_user_count":    todayUserCount,
		"today_video_count":   todayVideoCount,
	}, "success")
}

func (a *ApiController) HealthCheck(ctx *gin.Context) {
	utils.StatusOK(ctx, gin.H{
		"status": "ok",
	}, "admin service is running")
}
