package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL     string   `mapstructure:"DATABASE_URL"`
	RedisURL        string   `mapstructure:"REDIS_URL"`
	ProxyURLs       []string `mapstructure:"PROXY_URLS"`
	BrowserPoolSize int      `mapstructure:"BROWSER_POOL_SIZE"`
	CrawlQueueConc  int      `mapstructure:"CRAWL_QUEUE_CONC"`
	NormQueueConc   int      `mapstructure:"NORM_QUEUE_CONC"`
	TwoCaptchaKey   string   `mapstructure:"TWO_CAPTCHA_KEY"`
	AntiCaptchaKey  string   `mapstructure:"ANTI_CAPTCHA_KEY"`
	OTelEndpoint    string   `mapstructure:"OTEL_ENDPOINT"`
	RateLimitPerSec float64  `mapstructure:"RATE_LIMIT_PER_SEC"`
}

func MustLoad() *Config {
	viper.AutomaticEnv()
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()

	viper.SetDefault("BROWSER_POOL_SIZE", 3)
	viper.SetDefault("CRAWL_QUEUE_CONC", 5)
	viper.SetDefault("NORM_QUEUE_CONC", 10)
	viper.SetDefault("RATE_LIMIT_PER_SEC", 0.033)

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic("config: " + err.Error())
	}
	return &cfg
}
