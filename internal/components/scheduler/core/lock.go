package core

import (
	"context"
	"errors"
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/pkg/utils"
	"time"
)

type DistributedLock struct {
	rdb     *infra.Redis
	timeout time.Duration
}

func NewDistributedLock(rdb *infra.Redis, timeout time.Duration) *DistributedLock {
	return &DistributedLock{rdb: rdb}
}

func (l *DistributedLock) Lock(key string) (string, error) {
	lock := fmt.Sprintf("lock:%s", key)
	id := utils.CreateUUID()
	success, err := l.rdb.SetNX(context.Background(), lock, id, l.timeout)
	if err != nil {
		return "", err
	}

	if !success {
		return "", errors.New("distributed lock timeout")
	}

	return id, nil
}

func (l *DistributedLock) Unlock(lock, key string) error {
	lock = fmt.Sprintf("lock:%s", lock)
	data, err := l.rdb.Get(context.Background(), lock)
	if err != nil {
		return err
	}

	if string(data) != key {
		return errors.New("unlock failed")
	}

}

func (l *DistributedLock) watchDog() {

}
