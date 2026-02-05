package db

import (
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"stream_hub/pkg/model/config"
)

type MinioClient struct {
	client *minio.Client
	core   *minio.Core
}

func NewMinioClient(conf *config.CommonConfig) (*MinioClient, error) {
	endpoint := fmt.Sprintf("%s:%s", conf.Minio.Endpoint, conf.Minio.Port)

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			conf.Minio.AccessKey,
			conf.Minio.SecretKey,
			"",
		),
		Secure: false,
	})

	if err != nil {
		return nil, err
	}

	core, err := minio.NewCore(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			conf.Minio.AccessKey,
			conf.Minio.SecretKey,
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	return &MinioClient{client: client, core: core}, nil
}

func (m *MinioClient) Minio() *minio.Client {
	return m.client
}

func (m *MinioClient) Core() *minio.Core {
	return m.core
}
