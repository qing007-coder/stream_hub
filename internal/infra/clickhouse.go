package infra

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"reflect"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
	"strings"
	"time"
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

	type fieldMeta struct {
		index   int
		convert string
	}
	var columns []string
	var metas []fieldMeta

	for i := 0; i < elemType.NumField(); i++ {
		tag := elemType.Field(i).Tag.Get("ck")
		if tag != "" && tag != "-" {
			// 拆分 tag 例如 "event_time,million2time" -> ["event_time", "milli2time"]
			parts := strings.Split(tag, ",")
			columns = append(columns, parts[0]) // 第一部分永远是列名

			meta := fieldMeta{index: i}
			if len(parts) > 1 {
				meta.convert = parts[1] // 第二部分是指令
			}
			metas = append(metas, meta)
		}
	}

	batch, err := r.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s (%s)", table, strings.Join(columns, ",")))
	if err != nil {
		return err
	}

	for i := 0; i < v.Len(); i++ {
		structVal := v.Index(i)
		var row []interface{}

		for _, meta := range metas {
			fieldVal := structVal.Field(meta.index).Interface()

			switch meta.convert {
			case "million2time":
				if ts, ok := fieldVal.(float64); ok {
					// 拆出整数秒
					sec := int64(ts)
					// 算出纳秒偏移量 (0.353486 * 1,000,000,000)
					nsec := int64((ts - float64(sec)) * 1e9)
					//  合成 time.Time
					fieldVal = time.Unix(sec, nsec)
				}
			}

			row = append(row, fieldVal)
		}

		if len(row) != 11 {
			return fmt.Errorf("字段对齐失败: 期待 11, 实际解析出 %d. 请检查 ck 标签是否写漏了", len(row))
		}

		if err := batch.Append(row...); err != nil {
			return fmt.Errorf("clickhouse append error at row %d: %w", i, err)
		}
	}

	return batch.Send()
}
