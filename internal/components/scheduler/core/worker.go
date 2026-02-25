package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	infra_ "stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"time"

	"github.com/go-redis/redis/v8"
)

type Worker struct {
	id              string
	picker          *Picker       // 队列获取决策
	concurrencyChan chan struct{} // 最大并发数
	activeQueue     string        // 自己的私有队列: schedule:active:worker_{id}
	db              *infra.DB
	rdb             *infra.Redis
	deadLetter      string
	serveMux        *ServeMux
	retry           *Retry
	taskHealth      *TaskHealth
	deathChan       chan string
}

func NewWorker(id string, db *infra.DB, rdb *infra.Redis, conf *config.SchedulerConfig, deathChan chan string) *Worker {
	picker := NewQueuePicker(conf.Queue)
	concurrencyChan := make(chan struct{}, conf.Concurrency)
	return &Worker{
		id:              id,
		rdb:             rdb,
		concurrencyChan: concurrencyChan,
		picker:          picker,
		activeQueue:     fmt.Sprintf("scheduler:active:worker_%s", id),
		retry:           NewRetry(rdb, conf),
		db:              db,
		taskHealth:      NewTaskHealth(rdb, conf),
		deathChan:       deathChan,
	}
}

func (w *Worker) Start() {
	go func() {
		log.Printf("worker %s is running\n", w.id)

		defer w.sendDeathSignal()

		go w.fetch()

		for {
			taskID, err := w.rdb.BRPop(context.Background(), time.Second*5, w.activeQueue)
			if err != nil {
				if !errors.Is(err, redis.Nil) {
					log.Println("err:", err)
				}
				continue
			}

			log.Printf("worker %s is handling the task %s\n", w.id, taskID[1])

			ctx := context.Background()
			pipeline := w.rdb.Pipeline()
			metaCmd := pipeline.HGetAll(ctx, "task:meta:"+taskID[1])
			dataCmd := pipeline.Get(ctx, "task:payload:"+taskID[1])

			_, err = pipeline.Exec(ctx)
			if err != nil {
				log.Println("err:", err)
				continue
			}

			meta, _ := metaCmd.Result()
			data, _ := dataCmd.Bytes()

			fmt.Println("meta:", meta)
			fmt.Println("payload:", string(data))

			var payload infra_.TaskPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				log.Println("err:", err)
				continue
			}

			task := new(infra_.TaskMessage)
			task.TransformByMap(meta)
			task.Payload = payload

			if w.taskHealth.Check(task) {
				w.taskHealth.HandleBlackList(task)
				continue
			}

			w.execute(task)
		}
	}()
}

func (w *Worker) fetch() {
	for {
		queue := w.picker.NextQueue()
		log.Printf("worker %s is fetching, the queue is %s\n", w.id, queue)
		taskID, err := w.rdb.BRPop(context.Background(), 5*time.Second, queue)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Println("BRPop err:", err)
			}
			continue
		}

		if err := w.rdb.LPush(context.Background(), w.activeQueue, taskID[1]); err != nil {
			log.Println("err:", err)
			continue
		}

		w.concurrencyChan <- struct{}{}
	}
}

func (w *Worker) RegisterMux(mux *ServeMux) {
	w.serveMux = mux
}

func (w *Worker) execute(task *infra_.TaskMessage) {
	if err := w.serveMux.Execute(context.Background(), task.Type, task); err != nil {
		w.retryTask(task, err)
		return
	}

	if err := w.rdb.Del(context.Background(), "task:meta:"+task.TaskID, "task:payload:"+task.TaskID); err != nil {
		log.Println("err:", err)
		return
	}

	if err := w.db.Model(&storage.Task{}).Where("id = ?", task.TaskID).Updates(map[string]interface{}{
		"status":      constant.TaskSuccess,
		"retry_count": task.RetryCount,
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
