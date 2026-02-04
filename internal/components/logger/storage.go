package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
	"time"
)

type StorageWorker struct {
	conn            *infra.Clickhouse
	userBatchChan   <-chan []byte
	systemBatchChan <-chan []byte
}

func NewStorageWorker(conf *config.CommonConfig, userBatchChan <-chan []byte, systemBatchChan <-chan []byte) (*StorageWorker, error) {
	conn, err := infra.NewClickhouse(conf)
	if err != nil {
		return nil, err
	}

	s := new(StorageWorker)
	s.conn = conn
	s.userBatchChan = userBatchChan
	s.systemBatchChan = systemBatchChan

	return s, nil
}

func (s *StorageWorker) Start() {
	go func() {
		fmt.Println("start storage worker")
		for {
			select {
			case batch := <-s.userBatchChan:
				fmt.Println("user_logs:", string(batch))
				if err := s.UserBatchInsert(batch); err != nil {
					log.Println("err:", err)
				}
			case batch := <-s.systemBatchChan:
				fmt.Println("system_logs:", string(batch))
				if err := s.SystemBatchInsert(batch); err != nil {
					log.Println("err:", err)
				}
			}
		}
	}()
}

func (s *StorageWorker) UserBatchInsert(msg []byte) error {
	var logs []storage.UserLogEntry
	if err := json.Unmarshal(msg, &logs); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := s.conn.BatchInsertStruct(ctx, constant.StorageUserLog, logs); err != nil {
		return err
	}

	return nil
}

func (s *StorageWorker) SystemBatchInsert(msg []byte) error {
	var logs []storage.SystemLogEntry
	if err := json.Unmarshal(msg, &logs); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := s.conn.BatchInsertStruct(ctx, constant.StorageSystemLog, logs); err != nil {
		return err
	}

	return nil
}
