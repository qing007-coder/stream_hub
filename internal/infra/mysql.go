package infra

import (
	"gorm.io/gorm"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/storage"
)

type DB struct {
	*gorm.DB
}

func NewMysql(conf *config.CommonConfig) (*DB, error) {
	client, err := db.NewMysqlClient(conf)
	if err != nil {
		return nil, err
	}

	m := client.DB()

	if err := m.AutoMigrate(
		&storage.User{},
		&storage.Task{},
	); err != nil {
		return nil, err
	}

	return &DB{
		m,
	}, err
}
