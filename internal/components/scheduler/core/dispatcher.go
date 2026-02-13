package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"stream_hub/internal/infra"
	"time"

	"github.com/go-redis/redis/v8"
)

type Dispatcher struct {
	rdb *infra.Redis
	queue string 
	batchSize int
}

func NewDispatcher(rdb *infra.Redis) *Dispatcher {
	return &Dispatcher{
		rdb: rdb,
	}
}

func (d *Dispatcher) Scan(ctx context.Context) error {
	now := strconv.FormatInt(time.Now().Unix(), 10)

	taskIDs, err := d.rdb.ZRangeByScore(ctx, d.queue, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   now,
		Count: int64(d.batchSize),
	}).Result()
	if err != nil || len(taskIDs) == 0 {
		return err
	}

	pipe := d.rdb.Pipeline()
	cmds := make(map[string]*redis.StringCmd, len(taskIDs))

	for _, taskID := range taskIDs {
		cmds[taskID] = pipe.HGet(ctx, "task:meta:"+taskID, "priority")
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	n, err := d.rdb.ZRem(ctx, d.queue, taskIDs).Result()
	if err != nil {
		return err
	}
	if n != int64(len(taskIDs)) {
		return errors.New("miss zset remove")
	}

	pipe = d.rdb.Pipeline()
	for taskID, cmd := range cmds {
		priority := cmd.Val()
		queue := fmt.Sprintf("scheduler:queue:%s", priority)
		pipe.LPush(ctx, queue, taskID)
	}

	_, err = pipe.Exec(ctx)
	return err
}
