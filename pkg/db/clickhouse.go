package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log"
	"stream_hub/pkg/model/config"
	"time"
)

type ClickhouseClient struct {
	ctx  context.Context
	conn driver.Conn
}

func NewClickhouseClient(conf *config.CommonConfig) (*ClickhouseClient, error) {
	c := new(ClickhouseClient)
	c.ctx = context.Background()

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", conf.Clickhouse.Addr, conf.Clickhouse.Port)},
		Auth: clickhouse.Auth{
			Database: conf.Clickhouse.Database,
			Username: conf.Clickhouse.Username,
			Password: "",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	log.Println("clickhouse connected")

	c.conn = conn
	return c, nil
}

func (c *ClickhouseClient) Conn() driver.Conn {
	return c.conn
}
