package core

import "stream_hub/internal/infra"

type Dispatcher struct {
	rdb *infra.Redis
}
