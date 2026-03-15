package scraper

import (
	"context"
	"time"

	"review-curator/pkg/module/product"
	"review-curator/pkg/platform/database"

	"github.com/rs/zerolog/log"
)

const defaultReCrawlInterval = 6 * time.Hour

type Scheduler struct {
	crawl    *CrawlService
	products product.Repository
	interval time.Duration
}

func NewScheduler(crawl *CrawlService, products product.Repository, interval time.Duration) *Scheduler {
	if interval == 0 {
		interval = defaultReCrawlInterval
	}
	return &Scheduler{crawl: crawl, products: products, interval: interval}
}

func (s *Scheduler) Start(ctx context.Context) {
	go func() {
		s.runOnce(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

func (s *Scheduler) runOnce(ctx context.Context) {
	products, err := s.products.ListActive(ctx)
	if err != nil {
		log.Error().Err(err).Msg("scheduler: list active products")
		return
	}
	for _, p := range products {
		if s.hasRecentJob(ctx, p) {
			continue
		}
		_, err := s.crawl.EnqueueJob(ctx, Platform(p.Platform), p.ProductURL, p.ProductID, 10)
		if err != nil {
			log.Error().Err(err).Str("product_id", p.ID).Msg("scheduler: enqueue job")
		}
	}
}

func (s *Scheduler) hasRecentJob(ctx context.Context, p product.Product) bool {
	for _, status := range []string{"pending", "running"} {
		jobs, _, err := s.crawl.ListJobs(ctx, p.Platform, status, database.Page{Limit: 1})
		if err != nil {
			continue
		}
		for _, j := range jobs {
			if j.ProductID == p.ProductID && time.Since(j.EnqueuedAt) < s.interval {
				return true
			}
		}
	}
	return false
}
