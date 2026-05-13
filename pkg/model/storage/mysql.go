package storage

import (
	"encoding/json"
	"stream_hub/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.ID = utils.CreateID() // 统一调用你的工具类
	return nil
}

// User 基础信息表
type User struct {
	BaseModel
	Password      string `gorm:"type:varchar(255);not null" json:"-"`        // 密码
	Email         string `gorm:"type:varchar(128);uniqueIndex" json:"email"` // 邮箱
	Nickname      string `gorm:"type:varchar(64)" json:"nickname"`           // 昵称
	BackgroundURL string `gorm:"type:varchar(255)" json:"background_url"`
	Avatar        string `gorm:"type:varchar(255)" json:"avatar"`      // 头像URL
	Signature     string `gorm:"type:varchar(255)" json:"signature"`   // 个人签名
	Gender        int8   `gorm:"type:tinyint;default:0" json:"gender"` // 0:未知, 1:男, 2:女

	// 推荐系统核心画像特征 (冗余常用标签，提升读取速度)
	Tags string `gorm:"type:varchar(512)" json:"tags"` // 兴趣标签，如 "科技,美食,二次元"

	// 统计数据 (高频变动建议后期抽离到Redis存储)
	LikeCount     int64 `gorm:"default:0" json:"like_count"`     // 点赞数
	FollowCount   int64 `gorm:"default:0" json:"follow_count"`   // 关注数
	FollowerCount int64 `gorm:"default:0" json:"follower_count"` // 粉丝数
	WorkCount     int64 `gorm:"default:0" json:"work_count"`     // 作品数
	FavoriteCount int64 `gorm:"default:0" json:"favorite_count"` // 点赞作品数

	// 账号状态
	Status int8 `gorm:"type:tinyint;default:1" json:"status"` // 1:正常, 2:封禁, 3:注销
}

// Task 统一任务表
type Task struct {
	BaseModel
	// 任务类型
	Type string `gorm:"type:varchar(64);not null;index" json:"type"`
	// 示例：send_email_code / video_transcode / video_audit

	// 业务ID（关联具体业务）
	BizID string `gorm:"type:varchar(128);index" json:"biz_id"`
	// 示例：user_id / video_id / order_id

	// 任务状态
	Status int8 `gorm:"not null;index" json:"status"`
	// 0-待执行 1-成功 2-失败

	// 执行次数
	RetryCount int `gorm:"not null;default:0" json:"retry_count"`

	// 失败原因
	ErrorMsg string `gorm:"type:varchar(512)" json:"error_msg"`

	// 任务负载（JSON）
	Payload string `gorm:"type:text" json:"payload"`

	// 下次执行时间（支持延迟任务）
	NextRunAt int64 `gorm:"index" json:"next_run_at"`

	// ===== 运维 & 扩展 =====
	Executor string `gorm:"type:varchar(64);comment:执行节点" json:"executor"`
	Remark   string `gorm:"type:varchar(255);comment:人工备注" json:"remark"`
}

// FileModel 物理文件表：记录 MinIO 中的实际文件信息
// 只要文件内容一致（Hash相同），该表就只有一条记录
type FileModel struct {
	BaseModel
	FileHash string `gorm:"type:varchar(64);uniqueIndex;not null;comment:文件唯一哈希(MD5或SHA256)" json:"file_hash"`
	FilePath string `gorm:"type:varchar(255);not null;comment:MinIO中的存储路径" json:"file_path"`
	Size     int64  `gorm:"comment:文件大小(字节)" json:"size"`
	FileType string `gorm:"type:varchar(20);comment:文件后缀名(如.mp4)" json:"file_type"`
	Status   int    `gorm:"default:0;comment:文件状态: 1-上传中, 2-已落地, 3-已转码" json:"status"`
}

// TableName 指定表名
func (FileModel) TableName() string {
	return "media_files"
}

// VideoModel 视频业务表：记录用户上传的视频信息
type VideoModel struct {
	BaseModel
	Title         string `gorm:"type:varchar(100);not null;comment:视频标题" json:"title"`
	Description   string `gorm:"type:text;comment:视频简介" json:"description"`
	AuthorID      string `gorm:"index;comment:上传者用户ID" json:"author_id"`
	CoverUrl      string `gorm:"type:varchar(255);comment:封面图地址" json:"cover_url"`
	IsPublic      int    `gorm:"default:0;comment:0-私密 1-开放" json:"is_public"`
	Duration      int64  `gorm:"comment:视频时长(秒)" json:"duration"`
	LikeCount     int64  `gorm:"default:0;comment:点赞数" json:"like_count"`
	CommentCount  int64  `gorm:"default:0;comment:评论数" json:"comment_count"`
	FavoriteCount int64  `gorm:"default:0;comment:收藏数" json:"favorite_count"`
	ViewCount     int64  `gorm:"default:0;comment:观看数" json:"view_count"`
}

func (VideoModel) TableName() string {
	return "user_videos"
}

// MediaModel 媒体资产表：管理视频文件、转码和审核状态
type MediaModel struct {
	BaseModel
	VideoID         string          `gorm:"index;comment:关联视频ID" json:"video_id"`
	Type            string          `gorm:"type:varchar(20);not null;comment:媒体类型(original/transcoded/cover)" json:"type"`
	SourceObjectKey string          `gorm:"type:varchar(255);index;comment:原视频文件引用" json:"source_object_key"`
	M3u8Url         string          `gorm:"type:varchar(255);comment:m3u8播放路径" json:"m3u8_url"`
	TranscodeStatus int             `gorm:"default:0;comment:0-待转码 1-转码中 2-转码完成 3-转码失败" json:"transcode_status"`
	AuditStatus     int             `gorm:"default:0;comment:0-待审核 1-机审中 2-机审通过 3-机审失败 4-人工审核中 5-审核通过 6-审核拒绝 7-封禁" json:"audit_status"`
	Metadata        json.RawMessage `gorm:"type:json;not null;comment:媒体原始元数据" json:"metadata"`
}

func (m *MediaModel) BeforeCreate(tx *gorm.DB) error {
	if err := m.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	if len(m.Metadata) == 0 {
		m.Metadata = json.RawMessage(`{}`)
	}
	return nil
}

func (MediaModel) TableName() string {
	return "media_assets"
}

type VideoLikeModel struct {
	BaseModel
	UserID  string `gorm:"type:varchar(32);index:idx_user_video,unique;comment:点赞用户ID" json:"user_id"`
	VideoID string `gorm:"type:varchar(32);index:idx_user_video,unique;index;comment:视频ID" json:"video_id"`
}

type VideoFavoriteModel struct {
	BaseModel

	UserID  string `gorm:"type:varchar(32);index:idx_user_video,unique;comment:收藏用户ID" json:"user_id"`
	VideoID string `gorm:"type:varchar(32);index:idx_user_video,unique;index;comment:视频ID" json:"video_id"`
}

type UserFollowModel struct {
	BaseModel

	UserID       string `gorm:"type:varchar(32);index:idx_user_target,unique;comment:关注者" json:"user_id"`
	TargetUserID string `gorm:"type:varchar(32);index:idx_user_target,unique;index;comment:被关注者" json:"target_user_id"`
}

type VideoCommentModel struct {
	BaseModel

	VideoID  string `gorm:"type:varchar(32);index;comment:视频ID" json:"video_id"`
	UserID   string `gorm:"type:varchar(32);index;comment:评论用户ID" json:"user_id"`
	Content  string `gorm:"type:text;comment:评论内容" json:"content"`
	ParentID string `gorm:"type:varchar(32);index;comment:父评论ID，一级评论为空" json:"parent_id"`
}

type MessageModel struct {
	BaseModel
	SenderID    string `gorm:"type:varchar(32);index;comment:发送者ID" json:"sender_id"`
	ReceiverID  string `gorm:"type:varchar(32);index;comment:接收者ID" json:"receiver_id"`
	Content     string `gorm:"type:text;comment:消息内容" json:"content"`
	ContentType int8   `gorm:"type:tinyint;default:1;comment:消息类型 1-文本 2-图片 3-语音" json:"content_type"`
	Status      int8   `gorm:"type:tinyint;default:0;comment:状态 0-未读 1-已读" json:"status"`
}

func (MessageModel) TableName() string {
	return "im_messages"
}

type ConversationModel struct {
	BaseModel
	UserID        string `gorm:"type:varchar(32);index;comment:用户ID" json:"user_id"`
	TargetUserID  string `gorm:"type:varchar(32);index;comment:目标用户ID" json:"target_user_id"`
	LastMessageID string `gorm:"type:varchar(32);comment:最后一条消息ID" json:"last_message_id"`
	UnreadCount   int    `gorm:"default:0;comment:未读消息数" json:"unread_count"`
}

func (ConversationModel) TableName() string {
	return "im_conversations"
}

type Admin struct {
	BaseModel
	Email    string `gorm:"type:varchar(128);uniqueIndex" json:"email"`
	Name     string `gorm:"type:varchar(64)" json:"name"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`
	Status   int8   `gorm:"type:tinyint;default:1" json:"status"`
}

func (Admin) TableName() string {
	return "admins"
}

type Rule struct {
	BaseModel
	Type string `gorm:"column:type;type:varchar(64);index" json:"type"`
	V0   string `gorm:"column:V0;type:varchar(128);index" json:"v0"`
	V1   string `gorm:"column:V1;type:varchar(128);index" json:"v1"`
	V2   string `gorm:"column:V2;type:varchar(128);index" json:"v2"`
	V3   string `gorm:"column:V3;type:varchar(128)" json:"v3"`
	V4   string `gorm:"column:V4;type:varchar(128)" json:"v4"`
}

func (Rule) TableName() string {
	return "permissions"
}
