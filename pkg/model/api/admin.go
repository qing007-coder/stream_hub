package api

type GetUserListReq struct {
	Page     int    `form:"page" binding:"min=1"`
	Size     int    `form:"size" binding:"min=10"`
	UserID   string `form:"user_id"`
	Nickname string `form:"nickname"`
}

type UpdateUserStatusReq struct {
	TargetUserID string `json:"target_user_id"`
	Status       string `json:"status"`
}

type GetVideoListReq struct {
	Page     int    `form:"page" binding:"min=1"`
	Size     int    `form:"size" binding:"min=10"`
	UserID   string `form:"user_id"`
	IsPublic bool   `form:"is_public"`
}

type UpdateVideoReq struct {
	Title       string `json:"title" binding:"max=100"`
	Description string `json:"description" binding:"max=5000"`
	CoverURL    string `json:"cover_url"`
	IsPublic    int    `json:"is_public" binding:"omitempty,oneof=0 1"`
	AuditStatus int    `json:"audit_status"`
}

type GetTaskListReq struct {
	Page     int    `form:"page" binding:"min=1"`
	Size     int    `form:"size" binding:"min=10"`
	TaskType string `form:"task_type"`
	Status   *int8  `form:"status"`
	BizID    string `form:"biz_id"`
}

type UpdateTaskStatusReq struct {
	TaskID string `json:"task_id"`
	Status int8   `json:"status"`
}
