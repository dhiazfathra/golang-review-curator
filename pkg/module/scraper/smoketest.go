package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

var knownProducts = map[Platform]struct {
	URL       string
	ProductID string
}{
	PlatformShopee:    {URL: "https://shopee.co.id/product/123456/789", ProductID: "789"},
	PlatformTokopedia: {URL: "https://tokopedia.com/store/product-slug", ProductID: "prod-slug"},
	PlatformBlibli:    {URL: "https://www.blibli.com/p/product-name/ps--AX123", ProductID: "AX123"},
}

var requiredFields = []string{
	"review_text", "rating", "author_name", "author_id", "reviewed_at", "product_id", "xhr_pattern",
}

type PlatformGetter interface {
	Get(platform Platform) (any, bool)
}

type SmokeTest struct {
	getter PlatformGetter
	repo   Repository
}

func NewSmokeTest(getter PlatformGetter, repo Repository) *SmokeTest {
	return &SmokeTest{getter: getter, repo: repo}
}

func (s *SmokeTest) RunAll(ctx context.Context) error {
	var lastErr error
	for platform, known := range knownProducts {
		if err := s.runOne(ctx, platform, known.URL, known.ProductID); err != nil {
			log.Warn().
				Err(err).
				Str("platform", string(platform)).
				Msg("smoke test: platform failure")
			lastErr = err
		}
	}
	return lastErr
}

func (s *SmokeTest) runOne(ctx context.Context, platform Platform, productURL, productID string) error {
	adapterAny, ok := s.getter.Get(platform)
	if !ok {
		return fmt.Errorf("smoke: no adapter for %s", platform)
	}

	adapter, ok := adapterAny.(interface {
		Platform() Platform
		FetchReviews(ctx context.Context, job CrawlJob) ([]CrawlResult, error)
	})
	if !ok {
		return fmt.Errorf("smoke: adapter does not implement required interface")
	}

	job := CrawlJob{
		ID:         "smoke-" + string(platform),
		Platform:   platform,
		ProductURL: productURL,
		ProductID:  productID,
		MaxPages:   1,
		Status:     "running",
		EnqueuedAt: time.Now(),
	}

	results, err := adapter.FetchReviews(ctx, job)
	if err != nil {
		s.updateSelectorHealth(ctx, platform, false)
		return fmt.Errorf("smoke %s: fetch: %w", platform, err)
	}
	if len(results) == 0 {
		s.updateSelectorHealth(ctx, platform, false)
		return fmt.Errorf("smoke %s: no results returned", platform)
	}

	s.updateSelectorHealth(ctx, platform, true)
	log.Info().
		Str("platform", string(platform)).
		Int("results", len(results)).
		Msg("smoke test: pass")
	return nil
}

func (s *SmokeTest) updateSelectorHealth(ctx context.Context, platform Platform, success bool) {
	now := time.Now()
	for _, field := range requiredFields {
		if success {
			_, _ = s.repo.(*postgresRepository).db.ExecContext(ctx,
				`UPDATE selector_configs SET last_success=$1 WHERE platform=$2 AND field=$3`,
				now, string(platform), field)
		} else {
			_, _ = s.repo.(*postgresRepository).db.ExecContext(ctx,
				`UPDATE selector_configs SET last_failure=$1 WHERE platform=$2 AND field=$3`,
				now, string(platform), field)
		}
	}
}
