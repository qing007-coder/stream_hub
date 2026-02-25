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

func (r *Redis) SAdd(ctx context.Context, key string, member interface{}) error {
	return r.Client.SAdd(ctx, key, member).Err()
}

func (r *Redis) SMember(ctx context.Context, key string) ([]string, error) {
	return r.Client.SMembers(ctx, key).Result()
}

func (r *Redis) SCard(ctx context.Context, key string) (int64, error) {
	return r.Client.SCard(ctx, key).Result()
}

func (r *Redis) HSet(ctx context.Context, key string, value ...interface{}) error {
	return r.Client.HSet(ctx, key, value...).Err()
}

func (r *Redis) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return r.Client.HGet(ctx, key, field)
}

func (r *Redis) HExists(ctx context.Context, key string, field string) (bool, error) {
	return r.Client.HExists(ctx, key, field).Result()
}

func (r *Redis) HIncrBy(ctx context.Context, key string, field string) error {
	return r.Client.HIncrBy(ctx, key, field, 1).Err()
}

func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

func (r *Redis) HDel(ctx context.Context, key, field string) error {
	return r.Client.HDel(ctx, key, field).Err()
}

func (r *Redis) Pipeline() redis.Pipeliner {
	return r.Client.Pipeline()
}

func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.Client.SIsMember(ctx, key, member).Result()
}

func (r *Redis) ZAdd(ctx context.Context, key string, member *redis.Z) error {
	return r.Client.ZAdd(ctx, key, member).Err()
}

func (r *Redis) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.Client.ZScore(ctx, key, member).Result()
}

func (r *Redis) ZRangeArgsWithScores(ctx context.Context, z redis.ZRangeArgs) ([]redis.Z, error) {
	return r.Client.ZRangeArgsWithScores(ctx, z).Result()
}

func (r *Redis) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return r.Client.ZRangeByScore(ctx, key, opt)
}

func (r *Redis) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return r.Client.ZRem(ctx, key, members...)
}

func (r *Redis) ZCard(ctx context.Context, key string) (int64, error) {
	return r.Client.ZCard(ctx, key).Result()
}

func (r *Redis) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.Client.ZRevRange(ctx, key, start, stop).Result()
}

func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.Client.LPush(ctx, key, values...).Err()
}

func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	return r.Client.RPop(ctx, key).Result()
}

func (r *Redis) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return r.Client.BRPop(ctx, timeout, keys...).Result()
}

func (r *Redis) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	return r.Client.Scan(ctx, cursor, match, count)
}

func (r *Redis) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.Client.ScriptLoad(ctx, script).Result()
}

func (r *Redis) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (bool, error) {
	return r.Client.EvalSha(ctx, sha1, keys, args...).Bool()
}