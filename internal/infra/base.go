package infra

import (
	"stream_hub/pkg/constant"
	"stream_hub/pkg/model/config"
)

type Base struct {
	Clickhouse *Clickhouse
	DB         *DB
	Redis      *Redis
	Minio      *Minio
	ES         *Elasticsearch
	Mongo      *Mongo
	Logger     *Logger
	TaskSender *TaskSender
}

func NewBase(conf *config.CommonConfig) (*Base, error) {
	clickhouse, err := NewClickhouse(conf)
	if err != nil {
		return nil, err
	}

	db, err := NewMysql(conf)
	if err != nil {
		return nil, err
	}

	redis := NewRedis(conf)
	minio, err := NewMinio(conf)
	if err != nil {
		return nil, err
	}

	es, err := NewElasticSearch(conf, constant.ESVideo)
	if err != nil {
		return nil, err
	}

	mongo, err := NewMongo(conf)
	if err != nil {
		return nil, err
	}

	logger, err := NewLogger(conf)
	if err != nil {
		return nil, err
	}

	taskSender, err := NewTaskSender(conf)
	if err != nil {
		return nil, err
	}

	return &Base{
		Clickhouse: clickhouse,
		DB:         db,
		Redis:      redis,
		Minio:      minio,
		ES:         es,
		Mongo:      mongo,
		Logger:     logger,
		TaskSender: taskSender,
	}, nil
}
