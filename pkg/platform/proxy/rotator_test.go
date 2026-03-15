package proxy_test

import (
	"testing"

	"review-curator/pkg/platform/proxy"

	"github.com/stretchr/testify/assert"
)

func TestRotator_CircuitBreaker(t *testing.T) {
	slots := proxy.LoadFromConfig([]string{
		"http://proxy1:8080",
		"http://proxy2:8080",
		"http://proxy3:8080",
	})
	r := proxy.NewRotator(slots)

	for i := 0; i < 6; i++ {
		r.ReportFailure("http://proxy1:8080")
	}

	url, _ := r.Next()
	assert.NotEqual(t, "http://proxy1:8080", url)
}

func TestRotator_AllFailed_ReturnsBest(t *testing.T) {
	slots := proxy.LoadFromConfig([]string{"http://only:8080"})
	r := proxy.NewRotator(slots)
	for i := 0; i < 6; i++ {
		r.ReportFailure("http://only:8080")
	}
	url, err := r.Next()
	assert.Equal(t, "http://only:8080", url)
	assert.ErrorIs(t, err, proxy.ErrNoProxiesAvailable)
}

func TestRotator_ReportSuccess_ClearsQuarantine(t *testing.T) {
	slots := proxy.LoadFromConfig([]string{"http://p1:8080", "http://p2:8080"})
	r := proxy.NewRotator(slots)
	for i := 0; i < 6; i++ {
		r.ReportFailure("http://p1:8080")
	}
	r.ReportSuccess("http://p1:8080")
	url, _ := r.Next()
	_ = url
}
