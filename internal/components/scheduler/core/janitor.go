package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
	errors_ "stream_hub/pkg/errors"
	"strings"
	"time"
)

type Janitor struct {
	rdb               *infra.Redis
	heartbeatInterval time.Duration
	ticker            *time.Ticker
	registerKey       string
	deathKey          string
	lock *DistributedLock
}

func NewJanitor(rdb *infra.Redis, conf *config.SchedulerConfig) *Janitor {
	return &Janitor{
		rdb:               rdb,
		heartbeatInterval: time.Duration(conf.HeartbeatInterval) * time.Millisecond,
		registerKey:       conf.RegisterKey,
		deathKey:          conf.DeathKey,
		lock: NewDistributedLock(rdb, conf),
	}
}

func (j *Janitor) Run() {
	log.Printf("janitor is running\n")
	j.ticker = time.NewTicker(j.heartbeatInterval)
	go j.ListenDeath()
	for {
		select {
		case <-j.ticker.C:
			log.Println("janitor is scaning")
			if err := j.Scan(); err != nil {
				if !errors.Is(err, redis.Nil) {
					log.Println("err:", err)
				}
			}
		}
	}
}

func (j *Janitor) Scan() error {
	resource := "scheduler:janitor"
	id, err := j.lock.Lock(resource)
	if err != nil {
		if errors.Is(err, errors_.ErrKeyExists) {
			if err := j.lock.Wait(resource); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	defer j.lock.Unlock(resource, id)

	aliveNodes := make(map[string]struct{})
	// 心跳名单
	iter := j.rdb.Scan(context.Background(), 0, "scheduler:heartbeat:*", 500).Iterator()
	for iter.Next(context.Background()) {
		parts := strings.Split(iter.Val(), ":")
		aliveNodes[parts[len(parts)-1]] = struct{}{}
	}

	regIter := j.rdb.Scan(context.Background(), 0, j.registerKey+"*", 500).Iterator()
	for regIter.Next(context.Background()) {
		fullKey := regIter.Val()
		parts := strings.Split(fullKey, ":")
		nodeID := parts[len(parts)-1]

		// 不在存活 Map 里，说明 Node 挂了
		if _, ok := aliveNodes[nodeID]; !ok {
			log.Printf("发现失联节点: %s", nodeID)
			workerMap, _ := j.rdb.HGetAll(context.Background(), fullKey)
			for workerID := range workerMap {
				if err := j.rdb.Del(context.Background(), j.registerKey+nodeID); err != nil {
					return err
				}
				if err := j.cleanup(workerID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// node的worker若死掉，node会主动报丧
func (j *Janitor) ListenDeath() {
	for {
		key, err := j.rdb.BRPop(context.Background(), 5*time.Second, j.deathKey)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				log.Println("BRPop err:", err)
			}
			continue
		}

		keys := strings.Split(key[1], ":")
		nodeID := keys[0]
		workerID := key[1]

		if err := j.rdb.HDel(context.Background(), j.registerKey+nodeID, workerID); err != nil {
			log.Println("HDel err:", err)
			continue
		}

		if err := j.cleanup(workerID); err != nil {
			if errors.Is(err, redis.Nil) {
				return 
			}
			log.Println("err:", err)
		}
	}
}

func (j *Janitor) cleanup(workerID string) error {
	queue := fmt.Sprintf("scheduler:active:worker_%s", workerID)
	for {
		taskID, err := j.rdb.RPop(context.Background(), queue)
		if err != nil {
			return nil
		}

		priority := j.rdb.HGet(context.Background(), "task:meta:"+taskID, "priority").Val()
		queue = fmt.Sprintf("scheduler:queue:%s", priority)
		if err := j.rdb.LPush(context.Background(), queue, taskID); err != nil {
			return err
		}
	}
}
