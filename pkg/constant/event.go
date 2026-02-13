package constant

// EventType 行为埋点类型（稳定枚举，不可随意新增）
const (
	EventCreateVideo = "create_video"

	EventLikeVideo     = "like_video"
	EventFavoriteVideo = "favorite_video"
	EventComment       = "comment"

	EventFollowUser   = "follow_user"
	EventUnfollowUser = "unfollow_user"
)

// ResourceType 行为作用的资源类型
const (
	ResourceVideo   = "video"
	ResourceComment = "comment"
	ResourceUser    = "user"
)

// EventSource 行为来源
const (
	SourceFeed    = "feed"
	SourceProfile = "profile"
	SourceSearch  = "search"
)

// ClientType 客户端类型
const (
	ClientWeb     = "web"
	ClientIOS     = "ios"
	ClientAndroid = "android"
)
