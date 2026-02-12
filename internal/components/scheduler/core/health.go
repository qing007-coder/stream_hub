package core

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
	infra_ "stream_hub/pkg/model/infra"
	"time"
)

type TaskHealth struct {
	rdb               *infra.Redis
	threshold         int
	duration          time.Duration
	blacklistDuration time.Duration
	taskDelay         time.Duration
}

func NewTaskHealth(rdb *infra.Redis, conf *config.SchedulerConfig) *TaskHealth {
	return &TaskHealth{
		rdb:               rdb,
		threshold:         conf.Health.Threshold,
		duration:          time.Duration(conf.Health.Duration) * time.Millisecond,
		blacklistDuration: time.Duration(conf.Health.BlacklistDuration) * time.Millisecond,
		taskDelay:         time.Duration(conf.Health.Delay) * time.Millisecond,
	}
}

func (t *TaskHealth) Check(task *infra_.TaskMessage) bool {
	existed, err := t.rdb.IsExisted(context.Background(), fmt.Sprintf("scheduler:blacklist:%s", task.Type))
	if err != nil {
		log.Println("err:", err)
		return existed
	}

	return existed
}

func (t *TaskHealth) HandleBlackList(task *infra_.TaskMessage) {
	if err := t.rdb.ZAdd(context.Background(), "task:delay", &redis.Z{
		Score:  float64(time.Now().Add(t.blacklistDuration).Unix()),
		Member: task.TaskID,
	}); err != nil {
		log.Println("err:", err)
	}
}

func (t *TaskHealth) HandleError(task *infra_.TaskMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	pipeline := t.rdb.Pipeline()
	pipeline.SAdd(ctx, fmt.Sprintf("scheduler:failed:%s", task.Type), task.TaskID)
	count, _ := t.rdb.SCard(ctx, fmt.Sprintf("scheduler:failed:%s", task.Type))
	if count == 1 {
		pipeline.Expire(ctx, fmt.Sprintf("scheduler:failed:%s", task.Type), t.duration)
	}

	if count >= int64(t.threshold) {
		pipeline.Del(ctx, fmt.Sprintf("scheduler:failed:%s", task.Type))
		pipeline.Set(ctx, fmt.Sprintf("scheduler:blacklist:%s", task.Type), 1, t.blacklistDuration)
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
