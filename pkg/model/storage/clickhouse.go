package storage

import "time"

// UserLogEntry 对应 ClickHouse 中的日志表结构
type UserLogEntry struct {
	EventTime float64 `json:"event_time" ck:"event_time,million2time"` // 事件发生时间
	Level     string  `json:"level"      ck:"level"`                   // 日志级别: info, warn, error
	UID       string  `json:"uid"        ck:"uid"`                     // 用户唯一标识
	IP        string  `json:"ip"         ck:"ip"`                      // 客户端IP
	Method    string  `json:"method"     ck:"method"`                  // GET, POST
	Path      string  `json:"path"       ck:"path"`                    // 接口路径
	Status    int16   `json:"status"     ck:"status"`                  // 状态码: 200, 400, 500
	Latency   int64   `json:"latency"    ck:"latency"`                 // 耗时(ms)
	Message   string  `json:"message"    ck:"message"`                 // 日志描述
	TraceID   string  `json:"trace_id"   ck:"trace_id"`                // 用于链路追踪的唯一ID
	Module    string  `json:"module"     ck:"module"`                  // 所属模块: user, video, stream
}

type SystemLogEntry struct {
	EventTime float64 `json:"event_time" ck:"event_time,million2time"`
	Level     string  `json:"level" ck:"level"`     // 日志级别: info, warn, error
	Type      string  `json:"type" ck:"type"`       // task, metric, bot, sys
	NodeID    string  `json:"node_id" ck:"node_id"` // 哪台机器
	Module    string  `json:"module" ck:"module"`   // 哪个模块
	Payload   string  `json:"payload" ck:"payload"` // 根据 Type 的不同，这里存放上述不同的结构体
	Msg       string  `json:"msg,omitempty" ck:"msg"`
}

type Event struct {
	// 基础元信息
	EventID   string `ck:"event_id"`
	EventType string `ck:"event_type"`

	// 用户相关
	UserID string `ck:"user_id"`

	// 资源定位
	ResourceType string `ck:"resource_type"` // video / comment / user
	ResourceID   string `ck:"resource_id"`

	// 时间
	Timestamp int64     `ck:"timestamp"`  // 秒级时间戳
	EventTime time.Time `ck:"event_time"` // 冗余，CK 用 DateTime

	// 扩展字段（拍平，避免 Map）
	Source string `ck:"source"` // feed / profile / search
	Client string `ck:"client"` // web / ios / android
}
