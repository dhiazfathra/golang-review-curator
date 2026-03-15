package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"review-curator/pkg/module/normaliser"
	"review-curator/pkg/module/product"
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/module/scraper/adapters"
	"review-curator/pkg/module/scraper/handler"
	"review-curator/pkg/platform/browser"
	"review-curator/pkg/platform/captcha"
	"review-curator/pkg/platform/config"
	"review-curator/pkg/platform/database"
	"review-curator/pkg/platform/observability"
	"review-curator/pkg/platform/proxy"
	"review-curator/pkg/platform/queue"
	"review-curator/pkg/platform/ratelimit"
	"review-curator/pkg/platform/selector"
)

type smokeGetterWrapper struct {
	registry *adapters.Registry
}

func (s *smokeGetterWrapper) Get(platform scraper.Platform) (any, bool) {
	return s.registry.Get(platform)
}

func main() {
	cfg := config.MustLoad()
	observability.InitLogger(false)

	db := database.MustConnect(cfg.DatabaseURL)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	sessionStore := browser.NewSessionStore(rdb)

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

	registry := adapters.NewRegistry(pool, rotator, captchaDispatcher, selectorStore, limiter, sessionStore)

	repo := scraper.NewRepository(db)
	normRepo := normaliser.NewRepository(db)
	productRepo := product.NewRepository(db)
	queueClient := queue.NewClient(cfg.RedisURL)
	defer func() { _ = queueClient.Close() }()

	crawlService := scraper.NewCrawlService(repo, queueClient)
	crawlHandler := handler.NewCrawlHandler(crawlService, registry)

	smokeTest := scraper.NewSmokeTest(&smokeGetterWrapper{registry: registry}, repo)
	smokeHandler := scraper.NewSmokeHandler(smokeTest)

	normService := normaliser.NewNormaliserService(repo, normRepo)
	normHandler := normaliser.NewNormaliserHandler(normService)

	scheduler := scraper.NewScheduler(crawlService, productRepo, 6*time.Hour)
	scheduler.Start(ctx)

	asynqScheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: cfg.RedisURL},
		&asynq.SchedulerOpts{Location: mustLoadLocation("Asia/Jakarta")},
	)
	if _, err := asynqScheduler.Register("0 2 * * 1",
		asynq.NewTask(scraper.TaskSmokeRunAll, nil),
		asynq.Queue("crawl"),
	); err != nil {
		log.Fatalf("scheduler register: %v", err)
	}
	go func() { _ = asynqScheduler.Run() }()

	srv := queue.NewServer(cfg.RedisURL, cfg.CrawlQueueConc, cfg.NormQueueConc)
	mux := asynq.NewServeMux()
	queue.RegisterHandlers(mux, map[string]asynq.Handler{
		scraper.TaskCrawlJob:        crawlHandler,
		scraper.TaskNormaliseReview: normHandler,
		scraper.TaskSmokeRunAll:     smokeHandler,
	})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
		asynqScheduler.Shutdown()
		srv.Shutdown()
	}()

	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker: %v", err)
	}
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time zone %s: %v", name, err)
	}
	return loc
}
