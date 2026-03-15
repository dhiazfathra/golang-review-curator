package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/module/scraper/adapters"
	"review-curator/pkg/module/scraper/handler"
	"review-curator/pkg/platform/browser"
	"review-curator/pkg/platform/captcha"
	"review-curator/pkg/platform/config"
	"review-curator/pkg/platform/database"
	"review-curator/pkg/platform/proxy"
	"review-curator/pkg/platform/queue"
	"review-curator/pkg/platform/ratelimit"
	"review-curator/pkg/platform/selector"
)

func main() {
	cfg := config.MustLoad()
	db := database.MustConnect(cfg.DatabaseURL)

	proxySlots := proxy.LoadFromConfig(cfg.ProxyURLs)
	rotator := proxy.NewRotator(proxySlots)
	limiter := ratelimit.NewLimiter(cfg.RateLimitPerSec)

	pool, err := browser.NewPool(cfg.BrowserPoolSize, "")
	if err != nil {
		log.Fatalf("browser pool: %v", err)
	}
	defer pool.Close()

	primaryCaptcha := captcha.NewTwoCaptcha(cfg.TwoCaptchaKey)
	secondaryCaptcha := captcha.NewAntiCaptcha(cfg.AntiCaptchaKey)
	captchaDispatcher := captcha.NewDispatcher(primaryCaptcha, secondaryCaptcha)

	selectorStore, err := selector.NewSelectorStore(db)
	if err != nil {
		log.Fatalf("selector store: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	selectorStore.StartHotReload(ctx)

	registry := adapters.NewRegistry(pool, rotator, captchaDispatcher, selectorStore, limiter)

	repo := scraper.NewRepository(db)
	queueClient := queue.NewClient(cfg.RedisURL)
	defer func() { _ = queueClient.Close() }()
	crawlService := scraper.NewCrawlService(repo, queueClient)
	crawlHandler := handler.NewCrawlHandler(crawlService, registry)

	srv := queue.NewServer(cfg.RedisURL, cfg.CrawlQueueConc, cfg.NormQueueConc)
	mux := asynq.NewServeMux()
	queue.RegisterHandlers(mux, map[string]asynq.Handler{
		scraper.TaskCrawlJob: crawlHandler,
	})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
		srv.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker: %v", err)
	}
}
