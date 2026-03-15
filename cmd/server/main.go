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
	"review-curator/pkg/platform/config"
	"review-curator/pkg/platform/database"
)

func main() {
	cfg := config.MustLoad()
	db := database.MustConnect(cfg.DatabaseURL)

	normRepo := normaliser.NewRepository(db)
	reviewRepo := review.NewRepository(db)
	productRepo := product.NewRepository(db)

	reviewService := review.NewService(reviewRepo)
	productService := product.NewService(productRepo)
	_ = normRepo

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	review.NewHandler(reviewService).RegisterRoutes(e)
	product.NewHandler(productService).RegisterRoutes(e)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
}
