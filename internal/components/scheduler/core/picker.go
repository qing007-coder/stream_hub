package core

import (
	"math/rand"
	"sync"
	"fmt"
	"time"
)

type Picker struct {
	pool []string
	mu   sync.RWMutex
	rng  *rand.Rand
}

func NewQueuePicker(weights map[string]int) *Picker {
	var pool []string
	for priority, weight := range weights {
		for i := 0; i < weight; i++ {
			queue := fmt.Sprintf("scheduler:queue:%s", priority)
			pool = append(pool, queue)
		}
	}

	return &Picker{
		pool: pool,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Picker) NextQueue() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.pool) == 0 {
		return ""
	}

	return p.pool[p.rng.Intn(len(p.pool))]
}
