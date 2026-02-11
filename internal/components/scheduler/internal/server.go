package internal

import (
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"stream_hub/internal/infra"
	"stream_hub/pkg/model/config"
	"stream_hub/pkg/model/scheduler"
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

	s.heartbeatTicker = time.NewTicker(s.heartbeatExpiry)

	for {
		select {
		case <-s.heartbeatTicker.C:
			if err := s.SendHeartbeat(); err != nil {
				log.Println("err:", err)
			}

		case workerID := <-s.workerDeathChan:
			s.mu.Lock()
			delete(s.workerPool, workerID)
			s.mu.Unlock()
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

	if err := s.rdb.HSet(context.Background(), "scheduler:active_workers", workers); err != nil {
		return err
	}

	return nil
}

// SendHeartbeat 假设心跳周期是 5 秒，我们把过期时间设为 15 秒（3倍容错）
func (s *Server) SendHeartbeat() error {
	s.mu.Lock()
	var payload scheduler.HeartbeatPayload
	payload.NodeID = s.id
	for workerID, _ := range s.workerPool {
		payload.WorkersID = append(payload.WorkersID, workerID)
	}
	s.mu.Unlock()

	data, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	if err := s.rdb.Set(context.Background(), "scheduler:heartbeat:node_"+s.id, data, s.heartbeatExpiry); err != nil {
		return err
	}

	return nil
}
