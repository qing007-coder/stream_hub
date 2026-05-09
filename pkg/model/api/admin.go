package api

type GetUserListReq struct {
	Page int `form:"page" binding:"min=1"`
	Size int `form:"size" binding:"min=10"`
	UserID string `form:"user_id"`
	Nickname string `form:"nickname"` // 支持模糊搜索
}

type UpdateUserStatusReq struct {
	TargetUserID string `json:"target_user_id"`
	Status string `json:"status"`
}

type GetVideoListReq struct {
	Page int `form:"page" binding:"min=1"`
	Size int `form:"size" binding:"min=10"`
	UserID string `form:"user_id"`
	IsPublic bool `form:"is_public"`
}

type UpdateVideoReq struct {
	Title       string `json:"title" binding:"max=100"`
	Description string `json:"description" binding:"max=5000"`
	CoverURL    string `json:"cover_url"`
	IsPublic    int    `json:"is_public" binding:"omitempty,oneof=0 1"`
	AuditStatus int    `json:"audit_status"`
}