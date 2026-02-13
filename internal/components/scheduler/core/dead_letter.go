package core

import (
	"context"
	"log"
	"strconv"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"time"

	"gorm.io/gorm"
)

type DeadLetter struct {
	rdb      *infra.Redis
	db       *gorm.DB
	queueKey string
	enable   bool
}

func NewDeadLetter(db *gorm.DB, rdb *infra.Redis, conf *config.SchedulerConfig) *DeadLetter {
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

	d.consume()
}

func (d *DeadLetter) consume() {
	for {
		data, err := d.rdb.BRPop(context.Background(), time.Second*5, d.queueKey)
		if err != nil {
			log.Println("err:", err)
			continue
		}
		taskID := data[1]

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		pipeline := d.rdb.Pipeline()
		errMsg := pipeline.HGet(ctx, "task:meta:"+taskID, "error_msg").Val()
		retryCount := pipeline.HGet(ctx, "task:meta:"+taskID, "retry_count").Val()
		count, _ := strconv.Atoi(retryCount)
		pipeline.Del(ctx, "task:meta:"+taskID, "task:payload:"+taskID)	
		_, err = pipeline.Exec(ctx)
		if err != nil {
			log.Println("err:", err)
			continue
		}

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
