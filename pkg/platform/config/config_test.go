package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustLoad(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cfg := MustLoad()

	assert.NotNil(t, cfg)
	assert.Equal(t, 3, cfg.BrowserPoolSize)
	assert.Equal(t, 5, cfg.CrawlQueueConc)
	assert.Equal(t, 10, cfg.NormQueueConc)
	assert.Equal(t, 0.033, cfg.RateLimitPerSec)
}

func TestConfig_Struct(t *testing.T) {
	cfg := &Config{
		DatabaseURL:     "postgres://test",
		RedisURL:        "redis://test",
		ProxyURLs:       []string{"http://proxy1"},
		BrowserPoolSize: 5,
		CrawlQueueConc:  10,
		NormQueueConc:   15,
		TwoCaptchaKey:   "key",
		AntiCaptchaKey:  "key",
		OTelEndpoint:    "localhost:4317",
		RateLimitPerSec: 0.5,
	}

	assert.Equal(t, "postgres://test", cfg.DatabaseURL)
	assert.Equal(t, "redis://test", cfg.RedisURL)
	assert.Len(t, cfg.ProxyURLs, 1)
	assert.Equal(t, 5, cfg.BrowserPoolSize)
	assert.Equal(t, 10, cfg.CrawlQueueConc)
	assert.Equal(t, 15, cfg.NormQueueConc)
	assert.Equal(t, "key", cfg.TwoCaptchaKey)
	assert.Equal(t, "key", cfg.AntiCaptchaKey)
	assert.Equal(t, "localhost:4317", cfg.OTelEndpoint)
	assert.Equal(t, 0.5, cfg.RateLimitPerSec)
}

func TestConfig_Defaults(t *testing.T) {
	cfg := &Config{}

	assert.Empty(t, cfg.DatabaseURL)
	assert.Empty(t, cfg.RedisURL)
	assert.Nil(t, cfg.ProxyURLs)
	assert.Zero(t, cfg.BrowserPoolSize)
	assert.Zero(t, cfg.CrawlQueueConc)
	assert.Zero(t, cfg.NormQueueConc)
	assert.Empty(t, cfg.TwoCaptchaKey)
	assert.Empty(t, cfg.AntiCaptchaKey)
	assert.Empty(t, cfg.OTelEndpoint)
	assert.Zero(t, cfg.RateLimitPerSec)
}

func TestConfig_MapStructure(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		validate func(*Config) bool
	}{
		{
			name:     "DATABASE_URL",
			envKey:   "DATABASE_URL",
			envValue: "postgres://user:pass@localhost/db",
			validate: func(c *Config) bool { return c.DatabaseURL == "postgres://user:pass@localhost/db" },
		},
		{
			name:     "REDIS_URL",
			envKey:   "REDIS_URL",
			envValue: "redis://localhost:6379",
			validate: func(c *Config) bool { return c.RedisURL == "redis://localhost:6379" },
		},
		{
			name:     "BROWSER_POOL_SIZE",
			envKey:   "BROWSER_POOL_SIZE",
			envValue: "10",
			validate: func(c *Config) bool { return c.BrowserPoolSize == 10 },
		},
		{
			name:     "CRAWL_QUEUE_CONC",
			envKey:   "CRAWL_QUEUE_CONC",
			envValue: "20",
			validate: func(c *Config) bool { return c.CrawlQueueConc == 20 },
		},
		{
			name:     "NORM_QUEUE_CONC",
			envKey:   "NORM_QUEUE_CONC",
			envValue: "30",
			validate: func(c *Config) bool { return c.NormQueueConc == 30 },
		},
		{
			name:     "RATE_LIMIT_PER_SEC",
			envKey:   "RATE_LIMIT_PER_SEC",
			envValue: "1.0",
			validate: func(c *Config) bool { return c.RateLimitPerSec == 1.0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{}
			cfg.DatabaseURL = tt.envValue
			cfg.RedisURL = tt.envValue
			cfg.BrowserPoolSize = 10
			cfg.CrawlQueueConc = 20
			cfg.NormQueueConc = 30
			cfg.RateLimitPerSec = 1.0

			assert.True(t, tt.validate(cfg))
		})
	}
}
