package core

import (
	"context"
	"fmt"
	"stream_hub/internal/infra"
	"stream_hub/pkg/errors"
	errors_ "stream_hub/pkg/errors"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"
	"time"
)

type DistributedLock struct {
	rdb     *infra.Redis
	timeout time.Duration
	detectInterval time.Duration
	deadline time.Duration
}

func NewDistributedLock(rdb *infra.Redis, conf *config.SchedulerConfig) *DistributedLock {
	return &DistributedLock{
		rdb: rdb,
		timeout: time.Duration(conf.Lock.LockTimeout) * time.Millisecond,
		deadline: time.Duration(conf.Lock.WaitDeadline) * time.Millisecond,
		detectInterval: time.Duration(conf.Lock.DetectInterval) * time.Millisecond,
	}
}

func (l *DistributedLock) Lock(resource string) (string, error) {
	lock := fmt.Sprintf("lock:%s", resource)
	id := utils.CreateUUID()
	success, err := l.rdb.SetNX(context.Background(), lock, id, l.timeout)
	if err != nil {
		return "", err
	}

	if !success {
		return "", errors_.ErrKeyExists
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
		return errors_.ErrInvalidValue
	}

	return l.rdb.Del(context.Background(), lock)
}

func (l *DistributedLock) Wait(lock string) error {
	ticker := time.NewTicker(l.detectInterval)
	deadlineTimer := time.NewTimer(l.deadline)
	lock = fmt.Sprintf("lock:%s", lock)
	for {
		select {
		case <-ticker.C:
			isExisted, err := l.rdb.IsExisted(context.Background(), lock)
			if err != nil {
				return err
			}

			if isExisted {
				continue
			}

			return nil
		case <- deadlineTimer.C:
			return errors.ErrWaitTimeout
		}
	}
}

// 这个暂时不用 
func (l *DistributedLock) watchDog() {

}
