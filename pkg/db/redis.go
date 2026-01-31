package db

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"stream_hub/pkg/model/config"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(conf *config.CommonConfig) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", conf.Redis.Addr, conf.Redis.Port),
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	return &RedisClient{client: client}
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}
