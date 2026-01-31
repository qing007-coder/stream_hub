package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"stream_hub/pkg/model/config"
)

type MysqlClient struct {
	db *gorm.DB
}

func NewMysqlClient(conf *config.CommonConfig) (*MysqlClient, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		conf.Mysql.Name,
		conf.Mysql.Password,
		conf.Mysql.Addr,
		conf.Mysql.Port,
		conf.Mysql.Database,
		conf.Mysql.Conf,
	)
	// 初始化数据库时的高级配置
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return &MysqlClient{
		db,
	}, nil
}

func (m *MysqlClient) DB() *gorm.DB {
	return m.db
}
