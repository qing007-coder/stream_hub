package interaction

import (
	"stream_hub/internal/infra"
)

// Interaction implements InteractionService
type Interaction struct {
	*Like
	*Favourite
	*Follow
	*Comment
}

func NewInteraction(base *infra.Base, sender *EventSender) *Interaction {
	return &Interaction{
		Like:      NewLike(base, sender),
		Favourite: NewFavourite(base, sender),
		Follow:    NewFollow(base, sender),
		Comment:   NewComment(base, sender),
	}
}
