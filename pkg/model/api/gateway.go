package api

import "time"

// CreateVideoRequest 创建视频请求
type CreateVideoRequest struct {
	Title           string `json:"title" binding:"required,max=100"`
	Description     string `json:"description" binding:"max=5000"`
	SourceObjectKey string `json:"source_object_key" binding:"required"`
	CoverURL        string `json:"cover_url" binding:"required"`
}

// AuthorVideoInfo 作者视频信息
type AuthorVideoInfo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CoverURL    string    `json:"cover_url"`
	Status      int32     `json:"status"`
	IsPublic    int32     `json:"is_public"`
	Duration    int64     `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PublicVideoInfo 公开视频信息
type PublicVideoInfo struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CoverURL  string    `json:"cover_url"`
	AuthorID  string    `json:"author_id"`
	Duration  int64     `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}

// GetVideoRequest 获取视频请求
type GetVideoRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// GetVideoResponse 获取视频响应
type GetVideoResponse struct {
	Data interface{} `json:"data"` // 可能是 AuthorVideoInfo 或 PublicVideoInfo
}

// UpdateVideoRequest 更新视频请求
type UpdateVideoRequest struct {
	VideoID     string `json:"video_id" uri:"video_id" binding:"required"`
	Title       string `json:"title" binding:"max=100"`
	Description string `json:"description" binding:"max=5000"`
	CoverURL    string `json:"cover_url"`
	IsPublic    int32  `json:"is_public" binding:"omitempty,oneof=0 1"`
}

// DeleteVideoRequest 删除视频请求
type DeleteVideoRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// DeleteVideoResponse 删除视频响应
type DeleteVideoResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ListUserPublishedVideosRequest 获取用户公开视频列表请求
type ListUserPublishedVideosRequest struct {
	UserID string `json:"user_id" uri:"user_id" binding:"required"`
	Page   int32  `json:"page" form:"page" binding:"omitempty,min=1"`
	Size   int32  `json:"size" form:"size" binding:"omitempty,min=1,max=50"`
}

// ListUserPublishedVideosResponse 获取用户公开视频列表响应
type ListUserPublishedVideosResponse struct {
	Videos []PublicVideoInfo `json:"videos"`
	Total  int64             `json:"total"`
}

// ListMyVideosRequest 获取我的视频列表请求
type ListMyVideosRequest struct {
	Page int32 `json:"page" form:"page" binding:"min=1"`
	Size int32 `json:"size" form:"size" binding:"min=1,max=50"`
}

// ListMyVideosResponse 获取我的视频列表响应
type ListMyVideosResponse struct {
	Videos []AuthorVideoInfo `json:"videos"`
	Total  int64             `json:"total"`
}

// ActionResponse 操作响应
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// UserView 用户视图
type UserView struct {
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}

// LikeRequest 点赞请求
type LikeRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// IsLikeRequest 是否点赞请求
type IsLikeRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// IsLikeResponse 是否点赞响应
type IsLikeResponse struct {
	IsLike bool `json:"is_like"`
}

// ListLikesRequest 获取点赞列表请求
type ListLikesRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
	Page    int32  `json:"page" form:"page" binding:"omitempty,min=1"`
	Size    int32  `json:"size" form:"size" binding:"omitempty,min=1,max=50"`
}

// ListLikesResponse 获取点赞列表响应
type ListLikesResponse struct {
	Users []UserView `json:"users"`
	Total int64      `json:"total"`
}

// FavoriteRequest 收藏请求
type FavoriteRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// IsFavoriteRequest 是否收藏请求
type IsFavoriteRequest struct {
	VideoID string `json:"video_id" uri:"video_id" binding:"required"`
}

// IsFavoriteResponse 是否收藏响应
type IsFavoriteResponse struct {
	IsFavorite bool `json:"is_favorite"`
}

// FollowRequest 关注请求
type FollowRequest struct {
	TargetUserID string `json:"target_user_id" uri:"target_user_id" binding:"required"`
}

// IsFollowRequest 是否关注请求
type IsFollowRequest struct {
	TargetUserID string `json:"target_user_id" uri:"target_user_id" binding:"required"`
}

// IsFollowResponse 是否关注响应
type IsFollowResponse struct {
	IsFollow bool `json:"is_follow"`
}

// ListFollowersRequest 获取粉丝列表请求
type ListFollowersRequest struct {
	UserID string `json:"user_id" uri:"user_id" binding:"required"`
	Page   int32  `json:"page" form:"page" binding:"omitempty,min=1"`
	Size   int32  `json:"size" form:"size" binding:"omitempty,min=1,max=50"`
}

// ListFollowersResponse 获取粉丝列表响应
type ListFollowersResponse struct {
	Users []UserView `json:"users"`
	Total int64      `json:"total"`
}

// ListFollowingsRequest 获取关注列表请求
type ListFollowingsRequest struct {
	UserID string `json:"user_id" uri:"user_id" binding:"required"`
	Page   int32  `json:"page" form:"page" binding:"omitempty,min=1"`
	Size   int32  `json:"size" form:"size" binding:"omitempty,min=1,max=50"`
}

// ListFollowingsResponse 获取关注列表响应
type ListFollowingsResponse struct {
	Users []UserView `json:"users"`
	Total int64      `json:"total"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	VideoID       string `json:"video_id" binding:"required"`
	Content       string `json:"content" binding:"required,max=1000"`
	ParentID      string `json:"parent_id" binding:"omitempty"`
	ReplyToUserID string `json:"reply_to_user_id" binding:"omitempty"`
}

// DeleteCommentRequest 删除评论请求
type DeleteCommentRequest struct {
	CommentID string `json:"comment_id" uri:"comment_id" binding:"required"`
}

// Comment 评论
type Comment struct {
	ID            string `json:"id"`
	VideoID       string `json:"video_id"`
	UserID        string `json:"user_id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	Content       string `json:"content"`
	ParentID      string `json:"parent_id"`
	ReplyToUserID string `json:"reply_to_user_id"`
	LikeCount     int64  `json:"like_count"`
	ReplyCount    int64  `json:"reply_count"`
	CreateTime    int64  `json:"create_time"`
}

// ListCommentsRequest 获取评论列表请求
type ListCommentsRequest struct {
	VideoID  string `json:"video_id" uri:"video_id" binding:"required"`
	Page     int32  `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int32  `json:"page_size" form:"page_size" binding:"omitempty,min=1,max=50"`
	ParentID string `json:"parent_id" form:"parent_id" binding:"omitempty"`
}

// ListCommentsResponse 获取评论列表响应
type ListCommentsResponse struct {
	Comments []Comment `json:"comments"`
	Total    int64     `json:"total"`
	HasMore  bool      `json:"has_more"`
}
