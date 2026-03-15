package proxy

import (
	"errors"
	"sync"
	"time"
)

var ErrNoProxiesAvailable = errors.New("proxy: no healthy proxies available")

type proxySlot struct {
	URL                 string
	healthScore         int
	consecutiveFailures int
	quarantined         time.Time
}

func (s *proxySlot) isAvailable() bool {
	if s.quarantined.IsZero() {
		return s.consecutiveFailures < 5
	}
	return time.Now().After(s.quarantined)
}

type Rotator struct {
	mu    sync.Mutex
	slots []*proxySlot
	idx   int
}

func NewRotator(slots []*proxySlot) *Rotator {
	return &Rotator{slots: slots}
}

func (r *Rotator) Next() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n := len(r.slots)
	for i := 0; i < n; i++ {
		s := r.slots[(r.idx+i)%n]
		if s.isAvailable() {
			r.idx = (r.idx + i + 1) % n
			return s.URL, nil
		}
	}

	best := r.slots[0]
	for _, s := range r.slots[1:] {
		if s.healthScore > best.healthScore {
			best = s
		}
	}
	return best.URL, ErrNoProxiesAvailable
}

func (r *Rotator) ReportSuccess(proxyURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.slots {
		if s.URL == proxyURL {
			if s.healthScore < 10 {
				s.healthScore++
			}
			s.consecutiveFailures = 0
			s.quarantined = time.Time{}
			return
		}
	}
}

func (r *Rotator) ReportFailure(proxyURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.slots {
		if s.URL == proxyURL {
			s.healthScore--
			s.consecutiveFailures++
			if s.consecutiveFailures >= 5 {
				s.quarantined = time.Now().Add(10 * time.Minute)
				s.healthScore = 0
			}
			return
		}
	}
}
