package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/infra"
	"stream_hub/pkg/model/storage"
	"time"

	"github.com/go-redis/redis/v8"
)

type TaskSender struct {
	rdb *Redis
	db  *DB
}

func NewTaskSender(conf *config.CommonConfig) (*TaskSender, error) {
	db, err := NewMysql(conf)
	if err != nil {
		return nil, err
	}

	taskSender := new(TaskSender)
	taskSender.rdb = NewRedis(conf)
	taskSender.db = db

	return taskSender, nil
}

func (t *TaskSender) SendTask(message infra.TaskMessage) error {
	payload, err := json.Marshal(&message.Payload)
	if err != nil {
		return err
	}

	task := storage.Task{
		Type:   message.Type,
		BizID:  message.BizID,
		Status: constant.TaskPending,
		Payload: string(payload),
	}

	t.db.Create(&task)
	message.TaskID = task.ID
	
	queue := fmt.Sprintf("scheduler:queue:%s", message.Priority)
	meta := message.StructToMap()

	pipeline := t.rdb.Pipeline()
	pipeline.LPush(context.Background(), queue, message.TaskID)
	pipeline.HSet(context.Background(), "task:meta:"+message.TaskID, meta)
	pipeline.Set(context.Background(), "task:payload:"+message.TaskID, payload, -1)

	_, err = pipeline.Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (t *TaskSender) SendDelayTask(message infra.TaskMessage, delayDuration time.Duration) error {
	ctx := context.Background()
	payload, err := json.Marshal(&message.Payload)
	if err != nil {
		return err
	}

	task := storage.Task{
		Type:   message.Type,
		BizID:  message.BizID,
		Status: constant.TaskPending,
		Payload: string(payload),
	}

	t.db.Create(&task)
	meta := message.StructToMap()
	message.TaskID = task.ID
	nextRunTime := time.Now().Add(delayDuration)

	pipeline := t.rdb.Pipeline()
	pipeline.ZAdd(ctx, "task:delay", &redis.Z{
		Member: task.ID,
		Score: float64(nextRunTime.Unix()),
	})
	pipeline.HSet(context.Background(), "task:meta:"+message.TaskID, meta)
	pipeline.Set(context.Background(), "task:payload:"+message.TaskID, payload, -1)

	_, err = pipeline.Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}