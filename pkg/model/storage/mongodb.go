package storage

import "go.mongodb.org/mongo-driver/bson/primitive"

type CommentModel struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"` // MongoDB 自动生成
	VideoID string             `bson:"video_id"`      // 索引字段
	UserID  string             `bson:"user_id"`
	// 冗余用户信息，提升 List 接口性能（压测关键）
	Nickname string `bson:"nickname"`
	Avatar   string `bson:"avatar"`

	Content       string `bson:"content"`
	ParentID      string `bson:"parent_id"`        // 一级评论为 "0"
	ReplyToUserID string `bson:"reply_to_user_id"` // 回复的对象UID

	LikeCount  int64 `bson:"like_count"`
	ReplyCount int64 `bson:"reply_count"`
	CreateTime int64 `bson:"create_time"` // 索引字段
	IsDeleted  bool  `bson:"is_deleted"`
}
