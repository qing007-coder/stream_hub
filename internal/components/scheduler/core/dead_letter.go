package core

import (
	"context"
	"errors"
	"log"
	"strconv"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"time"

	"github.com/go-redis/redis/v8"
)

type DeadLetter struct {
	rdb      *infra.Redis
	db       *infra.DB
	queueKey string
	enable   bool
}

func NewDeadLetter(db *infra.DB, rdb *infra.Redis, conf *config.SchedulerConfig) *DeadLetter {
	return &DeadLetter{
		rdb:      rdb,
		db:       db,
		queueKey: conf.DeadLetter.QueueKey,
		enable:   conf.DeadLetter.Enabled,
	}
}

func (d *DeadLetter) Start() {
	if !d.enable {
		return
	}

	log.Println("deadletter is consumering")
	d.consume()
}

func (d *DeadLetter) consume() {
	for {
		data, err := d.rdb.BRPop(context.Background(), time.Second*5, d.queueKey)
		if err != nil {
			if !errors.Is(err, redis.Nil)  {
				log.Println("deadletter: err:", err)
			}
			continue
		}
		taskID := data[1]

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		pipeline := d.rdb.Pipeline()
		errMsgCmd := pipeline.HGet(ctx, "task:meta:"+taskID, "error_msg")
		retryCountCmd := pipeline.HGet(ctx, "task:meta:"+taskID, "retry_count")
		pipeline.Del(ctx, "task:meta:"+taskID, "task:payload:"+taskID)	
		_, err = pipeline.Exec(ctx)
		if err != nil {
			log.Println("err:", err)
			continue
		}

		errMsg, _ := errMsgCmd.Result()
		retryCount, _ := retryCountCmd.Result()
		count, _ := strconv.Atoi(retryCount)

		if err := d.db.Model(&storage.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
			"status":      constant.TaskFailed,
			"error_msg":   errMsg,
			"retry_count": count,
		}).Error; err != nil {
			log.Println("err:", err)
		}

		cancel()
	}
}
