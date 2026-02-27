package metric_sdk

import (
	"errors"
	"strings"
	"sync"
    "sync/atomic"
)

type CounterVector struct {
    mu sync.RWMutex

    Name       string
    Help       string
    labelNames []string

    counters map[string]*Counter
}

func NewCounterVector(options CounterVectorOptions) *CounterVector {
    return &CounterVector{
        Name:       options.Name,
        Help:       options.Help,
        labelNames: options.Labels,
        counters:   make(map[string]*Counter),
    }
}

func (cv *CounterVector) WithLabelValues(vals ...string) (*Counter, error) {
    if len(vals) != len(cv.labelNames) {
        return nil, errors.New("label count mismatch")
    }

    key := strings.Join(vals, "\xff")

    cv.mu.RLock()
    counter, ok := cv.counters[key]
    cv.mu.RUnlock()

    if ok {
        return counter, nil
    }

    cv.mu.Lock()
    defer cv.mu.Unlock()

    counter = &Counter{}
    cv.counters[key] = counter

    return counter, nil
}

type Counter struct {
	value uint64
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *Counter) Add(v uint64) {
	atomic.AddUint64(&c.value, v)
}

func (c *Counter) Load() uint64 {
	return atomic.LoadUint64(&c.value)
}