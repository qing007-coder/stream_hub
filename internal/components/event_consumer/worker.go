package event_consumer

import (
	"context"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
)

type StorageWorker struct {
	conn *infra.Clickhouse
	eventChan <- chan []storage.Event
}

func NewStorageWorker(eventChan <-chan []storage.Event, conf *config.CommonConfig) (*StorageWorker, error) {
	conn, err := infra.NewClickhouse(conf)
	if err != nil {
		return nil, err
	}
	return &StorageWorker{
		conn: conn,
		eventChan: eventChan,
	}, nil
}

func (s *StorageWorker) Start() {
	go func () {
		for {
			select {
			case batch := <-s.eventChan:
				if err := s.storeEvents(batch); err != nil {
					log.Println("err:", err)
					continue
				}

				log.Println("success:", batch)
			}
		}
	}()
}


func (s *StorageWorker) storeEvents(batch []storage.Event) error {
	return s.conn.BatchInsertStruct(context.Background(), constant.StorageEvent, batch)
}