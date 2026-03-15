package ratelimit

import (
	"context"
	"sync"

	"golang.org/x/time/rate"
)

// Limiter manages per-domain rate limiting.
type Limiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      float64
}

// NewLimiter creates a new rate limiter with the specified requests per second.
func NewLimiter(rps float64) *Limiter {
	return &Limiter{limiters: make(map[string]*rate.Limiter), rps: rps}
}

// Wait blocks until the rate limit allows the request for the given domain.
func (l *Limiter) Wait(ctx context.Context, domain string) error {
	l.mu.Lock()
	lim, ok := l.limiters[domain]
	if !ok {
		lim = rate.NewLimiter(rate.Limit(l.rps), 1)
		l.limiters[domain] = lim
	}
	l.mu.Unlock()
	return lim.Wait(ctx)
}
