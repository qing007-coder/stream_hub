package infra

import (
	"github.com/minio/minio-go/v7"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
)

type Minio *minio.Client

func NewMinio(conf *config.CommonConfig) (Minio, error) {
	client, err := db.NewMinioClient(conf)
	if err != nil {
		return nil, err
	}

	return client.Minio(), nil
}
