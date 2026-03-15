package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"review-curator/pkg/module/normaliser"
	"review-curator/pkg/module/product"
	"review-curator/pkg/module/review"
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/platform/config"
	"review-curator/pkg/platform/database"
	"review-curator/pkg/platform/observability"
	"review-curator/pkg/platform/queue"
	"review-curator/pkg/platform/selector"
)

func main() {
	cfg := config.MustLoad()
	observability.InitLogger(false)

	tracerShutdown, err := observability.InitTracer(context.Background(), cfg.OTelEndpoint)
	if err != nil {
		log.Printf("tracer init (non-fatal): %v", err)
	} else {
		defer func() { _ = tracerShutdown(context.Background()) }()
	}

	db := database.MustConnect(cfg.DatabaseURL)

	selectorStore, err := selector.NewSelectorStore(db)
	if err != nil {
		log.Fatalf("selector store: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	selectorStore.StartHotReload(ctx)

	scraperRepo := scraper.NewRepository(db)
	normRepo := normaliser.NewRepository(db)
	reviewRepo := review.NewRepository(db)
	productRepo := product.NewRepository(db)
	_ = normRepo

	queueClient := queue.NewClient(cfg.RedisURL)
	defer func() {
		if err := queueClient.Close(); err != nil {
			log.Printf("queue client close: %v", err)
		}
	}()

	crawlService := scraper.NewCrawlService(scraperRepo, queueClient)
	reviewService := review.NewService(reviewRepo)
	productService := product.NewService(productRepo)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(observability.RequestLogger())

	review.NewHandler(reviewService).RegisterRoutes(e)
	product.NewHandler(productService).RegisterRoutes(e)
	scraper.NewAPIHandler(crawlService, selectorStore).RegisterRoutes(e)

	e.GET("/metrics", echo.WrapHandler(observability.MetricsHandler()))

	shutCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-shutCtx.Done()
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer timeoutCancel()
	_ = e.Shutdown(timeoutCtx)
}
