package infra

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"reflect"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
	"strings"
)

type Clickhouse struct {
	conn driver.Conn
}

func NewClickhouse(conf *config.CommonConfig) (*Clickhouse, error) {
	client, err := db.NewClickhouseClient(conf)
	if err != nil {
		return nil, err
	}

	return &Clickhouse{
		client.Conn(),
	}, nil
}

// BatchInsertStruct 传入结构体切片，自动解析写入
func (r *Clickhouse) BatchInsertStruct(ctx context.Context, table string, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("data must be a slice")
	}
	if v.Len() == 0 {
		return nil
	}

	firstElem := v.Index(0)
	elemType := firstElem.Type()

	var columns []string
	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("ck")
		if tag != "" && tag != "-" {
			columns = append(columns, tag)
		}
	}

	batch, err := r.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s (%s)", table, strings.Join(columns, ",")))
	if err != nil {
		return err
	}

	for i := 0; i < v.Len(); i++ {
		structVal := v.Index(i)
		var row []interface{}
		for j := 0; j < elemType.NumField(); j++ {
			if elemType.Field(j).Tag.Get("ck") != "" {
				row = append(row, structVal.Field(j).Interface())
			}
		}
		if err := batch.Append(row...); err != nil {
			return err
		}
	}

	return batch.Send()
}
