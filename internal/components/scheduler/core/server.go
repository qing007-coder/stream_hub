package core

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/utils"
	"sync"
	"time"
)

type Server struct {
	mu                sync.Mutex
	id                string
	rdb               *infra.Redis
	workerNum         int
	workerDeathChan   chan string
	registerKey       string
	deathKey          string
	workerPool        map[string]*Worker
	heartbeatTicker   *time.Ticker
	heartbeatInterval time.Duration
	heartbeatExpiry   time.Duration
}

func NewServer(db *gorm.DB, rdb *infra.Redis, conf *config.SchedulerConfig) *Server {
	server := new(Server)
	server.id = utils.CreateID()
	server.rdb = rdb
	server.workerNum = conf.WorkerNum
	server.workerPool = make(map[string]*Worker)
	server.heartbeatExpiry = time.Duration(conf.HeartbeatExpiry) * time.Millisecond
	server.heartbeatInterval = time.Duration(conf.HeartbeatInterval) * time.Millisecond
	server.registerKey = conf.RegisterKey
	server.deathKey = conf.DeathKey
	workerDeathChan := make(chan string, 10)
	server.workerDeathChan = workerDeathChan
	for i := 0; i < server.workerNum; i++ {
		workerID := utils.CreateUUID()
		worker := NewWorker(workerID, db, rdb, conf, workerDeathChan)
		server.workerPool[workerID] = worker  
	}

	return server
}

func (s *Server) Start() error {
	for _, worker := range s.workerPool {
		worker.Start()
	}

	if err := s.RegisterWorker(); err != nil {
		return err
	}

	s.heartbeatTicker = time.NewTicker(s.heartbeatInterval)

	for {
		select {
		case <-s.heartbeatTicker.C:
			go func() {
				if err := s.SendHeartbeat(); err != nil {
					log.Println("err:", err)
				}
			}()

		case workerID := <-s.workerDeathChan:
			s.mu.Lock()
			delete(s.workerPool, workerID)
			s.mu.Unlock()
			if err := s.SendDeath(workerID); err != nil {
				log.Println("err:", err)
			}
		}
	}
}

func (s *Server) RegisterWorker() error {
	s.mu.Lock()
	workers := make(map[string]string)
	for workerID, _ := range s.workerPool {
		workers[workerID] = s.id
	}
	s.mu.Unlock()

	if err := s.rdb.HSet(context.Background(), s.registerKey+s.id, workers); err != nil {
		return err
	}

	return nil
}

// SendHeartbeat 假设心跳周期是 5 秒，过期时间设为 15 秒（3倍容错）
func (s *Server) SendHeartbeat() error {
	return s.rdb.Set(context.Background(), "scheduler:heartbeat:"+s.id, 1, s.heartbeatExpiry)
}

func (s *Server) SendDeath(workerID string) error {
	key := fmt.Sprintf("%s:%s", s.id, workerID)
	return s.rdb.LPush(context.Background(), s.deathKey, key)
}
