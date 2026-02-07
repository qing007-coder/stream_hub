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

func NewInteraction(base *infra.Base) *Interaction {
	return &Interaction{
		Like:      NewLike(base),
		Favourite: NewFavourite(base),
		Follow:    NewFollow(base),
		Comment:   NewComment(base),
	}
}
