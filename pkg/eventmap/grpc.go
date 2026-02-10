package eventmap

import "stream_hub/pkg/constant"

var GRPCEndpointToEvent = map[string]string{
	"VideoService.CreateVideo":          constant.EventCreateVideo,
	"InteractionService.CreateLike":     constant.EventLikeVideo,
	"InteractionService.CreateFavorite": constant.EventFavoriteVideo,
	"InteractionService.CreateFollow":   constant.EventFollowUser,
	"InteractionService.DeleteFollow":   constant.EventUnfollowUser,
	"InteractionService.CreateComment":  constant.EventComment,
}
