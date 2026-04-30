package api

type MachineAudit struct {
	TargetURL string `json:"target_url"`
}

type AuditResult struct {
	Status  string        `json:"status"`  // 取值范围: "pass", "reject", "suspicious"
	Tags    []string      `json:"tags"`    // 命中的所有唯一标签集合
	Details []FrameDetail `json:"details"` // 命中规则的具体帧详情
}

type FrameDetail struct {
	FrameIndex int          `json:"frame_index"`
	Results    []RuleResult `json:"results"` // 该帧命中的具体规则结果
}

type RuleResult struct {
	Tag      string  `json:"tag"`               // 标签名称，如 "black", "skin", "ocr"
	Score    float64 `json:"score,omitempty"`   // 匹配分数（根据具体规则实现可选）
	Message  string  `json:"message,omitempty"` // 描述信息
}

type CalculateFeedReq struct {}

type CalculateFeedResp struct {
	Status bool `json:"status"`
	Message string `json:"message"`
}