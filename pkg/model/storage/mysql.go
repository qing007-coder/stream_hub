package storage

import (
	"encoding/json"
	"gorm.io/gorm"
	"stream_hub/pkg/utils"
	"time"
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
	Executor string `gorm:"type:varchar(64);comment:执行节点"`
	Remark   string `gorm:"type:varchar(255);comment:人工备注"`
}

// FileModel 物理文件表：记录 MinIO 中的实际文件信息
// 只要文件内容一致（Hash相同），该表就只有一条记录
type FileModel struct {
	BaseModel
	FileHash string `gorm:"type:varchar(64);uniqueIndex;not null;comment:文件唯一哈希(MD5或SHA256)"`
	FilePath string `gorm:"type:varchar(255);not null;comment:MinIO中的存储路径"`
	Size     int64  `gorm:"comment:文件大小(字节)"`
	FileType string `gorm:"type:varchar(20);comment:文件后缀名(如.mp4)"`
	Status   int    `gorm:"default:0;comment:文件状态: 1-上传中, 2-已落地, 3-已转码"`
}

// TableName 指定表名
func (FileModel) TableName() string {
	return "media_files"
}

// VideoModel 视频业务表：记录用户上传的视频信息
// 多个用户上传同一个视频，会有多条记录，但指向同一个 FileHash
type VideoModel struct {
	BaseModel
	Title           string          `gorm:"type:varchar(100);not null;comment:视频标题"`
	Description     string          `gorm:"type:text;comment:视频简介"`
	AuthorID        string          `gorm:"index;comment:上传者用户ID"`
	SourceObjectKey string          `gorm:"type:varchar(64);index;comment:原视频文件引用"`
	CoverUrl        string          `gorm:"type:varchar(255);comment:封面图地址"`
	Status          int             `gorm:"default:0;comment:0-待审核 1-审核通过 2-审核未通过"`
	IsPublic        int             `gorm:"default:0;comment:0-私密 1-开放"`
	Duration        int64           `gorm:"comment:视频时长(秒)"`
	VideoMeta       json.RawMessage `gorm:"type:json;not null;comment:视频原始元数据"`
}

func (v *VideoModel) BeforeCreate(tx *gorm.DB) error {
	if err := v.BaseModel.BeforeCreate(tx); err != nil {
		return err
	}

	if len(v.VideoMeta) == 0 {
		v.VideoMeta = json.RawMessage(`{}`)
	}
	return nil
}

func (VideoModel) TableName() string {
	return "user_videos"
}

type VideoLikeModel struct {
	BaseModel
	UserID  string `gorm:"type:varchar(32);index:idx_user_video,unique;comment:点赞用户ID"`
	VideoID string `gorm:"type:varchar(32);index:idx_user_video,unique;index;comment:视频ID"`
}

type VideoFavoriteModel struct {
	BaseModel

	UserID  string `gorm:"type:varchar(32);index:idx_user_video,unique;comment:收藏用户ID"`
	VideoID string `gorm:"type:varchar(32);index:idx_user_video,unique;index;comment:视频ID"`
}

type UserFollowModel struct {
	BaseModel

	UserID       string `gorm:"type:varchar(32);index:idx_user_target,unique;comment:关注者"`
	TargetUserID string `gorm:"type:varchar(32);index:idx_user_target,unique;index;comment:被关注者"`
}

type VideoCommentModel struct {
	BaseModel

	VideoID  string `gorm:"type:varchar(32);index;comment:视频ID"`
	UserID   string `gorm:"type:varchar(32);index;comment:评论用户ID"`
	Content  string `gorm:"type:text;comment:评论内容"`
	ParentID string `gorm:"type:varchar(32);index;comment:父评论ID，一级评论为空"`
}
