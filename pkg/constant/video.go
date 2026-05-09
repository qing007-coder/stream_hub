package constant

const (
	VideoPrivate = iota
	VideoPublic
)

const (
	VideoStatusPending = iota // 待审核（上传完成，等待审核）
	VideoStatusMachineChecking // 机审中
	VideoStatusMachinePassed // 机审通过
	VideoStatusMachineFailed // 机审失败
	VideoStatusManualChecking // 人工审核中
	VideoStatusApproved // 审核通过
	VideoStatusRejected // 审核拒绝
	VideoStatusBanned // 封禁
)