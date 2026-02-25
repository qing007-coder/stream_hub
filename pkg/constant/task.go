package constant

const (
	TaskPending int8 = 0 // 待执行
	TaskSuccess int8 = 1 // 成功
	TaskFailed  int8 = 2 // 失败
)

const (
	TaskSendEmailCode = "send_email_code"

	TaskVideoTranscode = "video_transcode"
	TaskVideoAudit     = "video_audit"

	TaskSendNotify = "send_notify"

	TaskVideoToES = "video_to_es"
)

const (
	ActionCreate = "action_create"
	ActionUpdate = "action_update"
	ActionDelete = "action_delete"
)