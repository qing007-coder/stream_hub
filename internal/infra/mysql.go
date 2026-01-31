package infra

import (
	"gorm.io/gorm"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
)

type DB *gorm.DB

func NewMysql(conf *config.CommonConfig) (DB, error) {
	client, err := db.NewMysqlClient(conf)
	if err != nil {
		return nil, err
	}

	m := client.DB()

	return m, err
}
