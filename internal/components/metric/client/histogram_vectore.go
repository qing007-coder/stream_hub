package metric_sdk

import (
	"errors"
	"math"
	"strings"
	"sync"
	"sync/atomic"
)

type HistogramVector struct {
	mu sync.RWMutex
	Name string 
	Help string 
	labelNames []string
	bucketBoundaries []float64
	histograms map[string]*Histogram
}

func NewHistogramVector(options HistogramVectorOptions) *HistogramVector {
	return &HistogramVector{
		Name: options.Name,
		Help: options.Help,
		labelNames: options.Labels,
		bucketBoundaries: options.BucketsBoundaries,
		histograms: make(map[string]*Histogram),
	}
}

func (hv *HistogramVector) WithLabelValues (vals ...string) (*Histogram, error) {
	if len(vals) != len(hv.labelNames) {
        return nil, errors.New("label count mismatch")
    }
	
    key := strings.Join(vals, "\xff")
	hv.mu.RLock()
    histogram, ok := hv.histograms[key]
    hv.mu.RUnlock()

    if ok {
        return histogram, nil
    }

    hv.mu.Lock()
    defer hv.mu.Unlock()

    histogram = newHistogram(hv.bucketBoundaries)
    hv.histograms[key] = histogram

    return histogram, nil
}

type Histogram struct {
	mu sync.Mutex
	sum float64
	count uint64
	buckets []uint64
	bucketBoundaries []float64
}

func newHistogram(boundaries []float64) *Histogram {
	return &Histogram{
		bucketBoundaries: append(boundaries, math.Inf(1)),
		buckets: make([]uint64, len(boundaries) + 1),
	}
}

func (h *Histogram) Observe(duration float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, boundary := range h.bucketBoundaries {
		if duration <= boundary {
			atomic.AddUint64(&h.buckets[i], 1)
			break
		}
	}

	atomic.AddUint64(&h.count, 1)
	h.sum += duration
}
