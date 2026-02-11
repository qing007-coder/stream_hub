package internal

import (
	"context"
	"gorm.io/gorm"
	"log"
	"strconv"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"time"
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
		errMsg := pipeline.HGet(ctx, "task:error", taskID).Val()
		retryCount := pipeline.HGet(ctx, "task:retry_count", taskID).Val()
		count, _ := strconv.Atoi(retryCount)
		pipeline.HDel(ctx, "task:error", taskID)
		pipeline.HDel(ctx, "task:pool", taskID)
		pipeline.HDel(ctx, "task:retry_count", taskID)
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
