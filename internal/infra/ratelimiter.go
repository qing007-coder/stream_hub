package infra

import (
	"context"
	"os"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"
	"time"
)

type Ratelimter struct {
	resource string 
	rdb *Redis
	baseLimit int 
	currentLimit int
	windowSize time.Duration
	scriptSHA  string
}

func NewRatelimiter(rdb *Redis, resource string, conf *config.CommonConfig) (*Ratelimter, error) {
	content, err := os.ReadFile(conf.Ratelimiter.ScriptPath)
	if err != nil {
		return nil, err
	}

	sha, err := rdb.ScriptLoad(context.Background(), string(content))
	if err != nil {
		return nil, err
	}

	return &Ratelimter{
		resource: resource,
		rdb: rdb,
		scriptSHA: sha,
		windowSize: time.Duration(conf.Ratelimiter.WindowSize) * time.Millisecond,
		baseLimit: conf.Ratelimiter.BaseLimit,
		currentLimit: conf.Ratelimiter.BaseLimit,
	}, nil
}

func (r *Ratelimter) Allow() (bool, error) {
	memberID := utils.CreateID()
	now := time.Now().UnixMilli()

	res, err := r.rdb.EvalSha(context.Background(), r.scriptSHA, []string{r.resource}, 
		now, 
		r.windowSize.Milliseconds(), 
		r.currentLimit, 
		memberID,
	)

	if err != nil {
		return false, err
	}

	return res, nil
}