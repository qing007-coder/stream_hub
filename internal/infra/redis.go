package infra

import (
	"context"
	"github.com/go-redis/redis/v8"
	"stream_hub/pkg/db"
	"stream_hub/pkg/model/config"
	"time"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(conf *config.CommonConfig) *Redis {
	return &Redis{
		Client: db.NewRedisClient(conf).Client(),
	}
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	return r.Client.Get(ctx, key).Bytes()
}

// SetNX key的值如果存在，则不做任何操作
func (r *Redis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}

func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

func (r *Redis) IsExisted(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

func (r *Redis) Incr(ctx context.Context, key string) error {
	return r.Client.Incr(ctx, key).Err()
}

func (r *Redis) Decr(ctx context.Context, key string) error {
	return r.Client.Decr(ctx, key).Err()
}

func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}
