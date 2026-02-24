package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"stream_hub/internal/infra"
	errors_ "stream_hub/pkg/errors"
	"stream_hub/pkg/model/config"
	"time"

	"github.com/go-redis/redis/v8"
)

type Dispatcher struct {
	rdb *infra.Redis
	queue string 
	batchSize int
	lock *DistributedLock
	ticker *time.Ticker
	scanInterval time.Duration
}

func NewDispatcher(rdb *infra.Redis, conf *config.SchedulerConfig) *Dispatcher {
	return &Dispatcher{
		rdb: rdb,
		queue: conf.Dispatcher.Queue,
		batchSize: conf.Dispatcher.BatchSize,
		lock: NewDistributedLock(rdb, conf),
		scanInterval: time.Duration(conf.Dispatcher.ScanInterval)*time.Millisecond,
	}
}

func (d *Dispatcher) Start() {
	d.ticker = time.NewTicker(d.scanInterval)
	for {
		select {
		case <- d.ticker.C:
			log.Println("dispatcher is scanning")
			if err := d.Scan(context.Background()); err != nil {
				log.Println("err:", err)
			}
		}
	}
}

func (d *Dispatcher) Scan(ctx context.Context) error {
	resource := "scheduler:dispathcer"
	id, err := d.lock.Lock(resource)
	if err != nil {
		if errors.Is(err, errors_.ErrKeyExists) {
			if err := d.lock.Wait(resource); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	defer d.lock.Unlock(resource, id)

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
