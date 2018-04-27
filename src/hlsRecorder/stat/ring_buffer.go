package stat

import (
	"sync"

	percentile "github.com/montanaflynn/stats"
)

type ringBuffer struct {
	sync.RWMutex
	cap, cur int
	slice    []float64
}

func NewRingBuffer(cap int) *ringBuffer {
	return &ringBuffer{
		cap:   cap,
		slice: make([]float64, cap),
	}
}

func (rb *ringBuffer) Add(v float64) {
	rb.Lock()
	rb.slice[rb.cur] = v
	rb.cur++
	rb.cur %= rb.cap
	rb.Unlock()
}

func (rb *ringBuffer) Percentile(f float64) float64 {
	result, _ := percentile.Percentile(rb.slice, f)
	return result
}
