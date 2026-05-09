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

	TaskCalculateFeed = "calculate_feed"

	TaskStorePersistency = "storage_persistency"
)

const (
	ActionCreateLike = "action_create_like"
	ActionUpdateLike = "action_update_like"
	ActionDeleteLike = "action_delete_like"

	ActionCreateFavourite = "action_create_favourite"
	ActionUpdateFavourite = "action_update_favourite"
	ActionDeleteFavourite = "action_delete_favourite"
)