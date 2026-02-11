package internal

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
	infra_ "stream_hub/pkg/model/infra"
	"time"
)

type Retry struct {
	MaxRetry     int64
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	EnableJitter bool
	rdb          *infra.Redis
}

func NewRetry(rdb *infra.Redis, conf *config.SchedulerConfig) *Retry {
	return &Retry{
		rdb:          rdb,
		MaxRetry:     conf.Retry.MaxRetries,
		BaseDelay:    time.Duration(conf.Retry.BaseDelayMs) * time.Millisecond,
		MaxDelay:     time.Duration(conf.Retry.MaxDelayMs) * time.Millisecond,
		EnableJitter: conf.Retry.EnableJitter,
	}
}

func (r *Retry) retry(task *infra_.TaskMessage, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	pipeline := r.rdb.Pipeline()
	count, _ := pipeline.HIncrBy(ctx, "task:retry_count", task.TaskID, 1).Result()

	if count >= r.MaxRetry {
		pipeline.HSet(ctx, "task:error", task.TaskID, err.Error())
		if err := r.sendToDlq(task.TaskID); err != nil {
			log.Println("err:", err)
			return
		}

		return
	}

	// 指数退避
	delay := time.Duration(1<<count) * r.BaseDelay

	if delay > r.MaxDelay {
		delay = r.MaxDelay
	}

	if r.EnableJitter {
		delay = time.Duration(rand.Int63n(int64(delay)))
	}

	nextRunTime := time.Now().Add(delay)
	pipeline.ZAdd(ctx, "task:delay", &redis.Z{
		Score:  float64(nextRunTime.Unix()),
		Member: task.TaskID,
	})

	_, err = pipeline.Exec(ctx)
	if err != nil {
		log.Println("err:", err)
		return
	}
}

func (r *Retry) sendToDlq(taskID string) error {
	return r.rdb.LPush(context.Background(), "scheduler:dlq", taskID)
}
