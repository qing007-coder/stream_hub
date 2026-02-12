package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"log"
	"strconv"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"time"
)

type Worker struct {
	id              string
	picker          *Picker       // 队列获取决策
	concurrencyChan chan struct{} // 最大并发数
	activeQueue     string        // 自己的私有队列: schedule:active:worker_{id}
	db              *gorm.DB
	rdb             *infra.Redis
	deadLetter      string
	serveMux        *ServeMux
	retry           *Retry
	taskHealth      *TaskHealth
	deathChan       chan string
}

func NewWorker(id string, db *gorm.DB, rdb *infra.Redis, conf *config.SchedulerConfig, deathChan chan string) *Worker {
	picker := NewQueuePicker(conf.Queue)
	concurrencyChan := make(chan struct{}, conf.Concurrency)
	return &Worker{
		id:              id,
		rdb:             rdb,
		concurrencyChan: concurrencyChan,
		picker:          picker,
		activeQueue:     fmt.Sprintf("scheduler:active:worker_%s", id),
		retry:           NewRetry(rdb, conf),
		serveMux:        NewServeMux(),
		db:              db,
		taskHealth:      NewTaskHealth(rdb, conf),
		deathChan:       deathChan,
	}
}

func (w *Worker) Start() {
	go func() {
		defer w.sendDeathSignal()

		go w.fetch()

		for {
			taskID, err := w.rdb.BRPop(context.Background(), time.Second*5, w.activeQueue)
			if err != nil {
				log.Println("err:", err)
				continue
			}

			data, err := w.rdb.HGet(context.Background(), " task:pool", taskID[1]).Bytes()
			if err != nil {
				log.Println("err:", err)
				continue
			}

			var task infra_.TaskMessage
			if err := json.Unmarshal(data, &task); err != nil {
				w.retryTask(&task, err)
				continue
			}

			if w.taskHealth.Check(&task) {
				w.taskHealth.HandleBlackList(&task)
				continue
			}

			w.execute(&task)
		}
	}()
}

func (w *Worker) fetch() {
	for {
		queue := fmt.Sprintf("scheduler:queue:%s", w.picker.NextQueue())
		taskID, err := w.rdb.BRPop(context.Background(), 5*time.Second, queue)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Println("BRPop err:", err)
			}
			continue
		}

		if err := w.rdb.LPush(context.Background(), queue, taskID[1]); err != nil {
			log.Println("err:", err)
			continue
		}

		w.concurrencyChan <- struct{}{}
	}
}

func (w *Worker) execute(task *infra_.TaskMessage) {
	if err := w.serveMux.Execute(context.Background(), task.Type, task); err != nil {
		w.retryTask(task, err)
		return
	}

	retryCount := w.rdb.HGet(context.Background(), "task:retry_count", task.TaskID).Val()
	pipeline := w.rdb.Pipeline()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pipeline.HDel(ctx, "task:retry_count", task.TaskID)
	pipeline.HDel(ctx, "task:pool", task.TaskID)

	_, err := pipeline.Exec(ctx)
	if err != nil {
		log.Println("err:", err)
		return
	}

	count, _ := strconv.Atoi(retryCount)

	if err := w.db.Model(&storage.Task{}).Where("id = ?", task.TaskID).Updates(map[string]interface{}{
		"status":      constant.TaskSuccess,
		"retry_count": count,
	}).Error; err != nil {
		log.Println("db.Updates err:", err)
		return
	}

	<-w.concurrencyChan
}

func (w *Worker) retryTask(task *infra_.TaskMessage, err error) {
	w.retry.retry(task, err)
	if err := w.taskHealth.HandleError(task); err != nil {
		log.Println("task health err:", err)
	}
}

func (w *Worker) sendDeathSignal() {
	w.deathChan <- w.id
	log.Printf("worker %s is dead\n", w.id)
}
