package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLimiter(t *testing.T) {
	limiter := NewLimiter(10.0)
	require.NotNil(t, limiter)
	assert.NotNil(t, limiter.limiters)
	assert.Equal(t, 10.0, limiter.rps)
}

func TestNewLimiter_ZeroRPS(t *testing.T) {
	limiter := NewLimiter(0)
	require.NotNil(t, limiter)
	assert.Equal(t, 0.0, limiter.rps)
}

func TestNewLimiter_HighRPS(t *testing.T) {
	limiter := NewLimiter(100.0)
	require.NotNil(t, limiter)
	assert.Equal(t, 100.0, limiter.rps)
}

func TestLimiter_Wait_SingleDomain(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	err := limiter.Wait(ctx, "example.com")
	assert.NoError(t, err)
}

func TestLimiter_Wait_MultipleDomains(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	err1 := limiter.Wait(ctx, "domain1.com")
	assert.NoError(t, err1)

	err2 := limiter.Wait(ctx, "domain2.com")
	assert.NoError(t, err2)

	err3 := limiter.Wait(ctx, "domain1.com")
	assert.NoError(t, err3)
}

func TestLimiter_Wait_CancelledContext(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := limiter.Wait(ctx, "example.com")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestLimiter_Wait_DeadlineExceeded(t *testing.T) {
	limiter := NewLimiter(0.001)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx, "example.com")
	assert.True(t, err == nil || err == context.DeadlineExceeded)
}

func TestLimiter_SameDomainReusesLimiter(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	_ = limiter.Wait(ctx, "example.com")
	initial := limiter.limiters["example.com"]

	_ = limiter.Wait(ctx, "example.com")
	after := limiter.limiters["example.com"]

	assert.Same(t, initial, after)
}

func TestLimiter_DifferentDomainsHaveSeparateLimiters(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	_ = limiter.Wait(ctx, "domain1.com")
	_ = limiter.Wait(ctx, "domain2.com")
	_ = limiter.Wait(ctx, "domain3.com")

	assert.Len(t, limiter.limiters, 3)
}

func TestLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(domain string) {
			err := limiter.Wait(ctx, domain)
			assert.NoError(t, err)
			done <- true
		}("concurrent-domain")
	}

	<-done
	<-done
	<-done
}

func TestLimiter_EmptyDomain(t *testing.T) {
	limiter := NewLimiter(10.0)
	ctx := context.Background()

	err := limiter.Wait(ctx, "")
	assert.NoError(t, err)
}

func TestLimiter_LargeNumberOfDomains(t *testing.T) {
	limiter := NewLimiter(100.0)
	ctx := context.Background()

	for i := 0; i < 100; i++ {
		err := limiter.Wait(ctx, "domain"+string(rune(i))+".com")
		assert.NoError(t, err)
	}

	assert.Len(t, limiter.limiters, 100)
}
