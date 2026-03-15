package adapters

import (
	"context"
	"encoding/json"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"review-curator/pkg/module/scraper"
	"review-curator/pkg/platform/selector"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var tokopediaGraphQLPattern = regexp.MustCompile(`tokopedia\.com/graphql`)

const tokopediaReviewOperation = "ProductRevGetProductReviewList"

type TokopediaAdapter struct {
	BaseScraper
}

func (a *TokopediaAdapter) Platform() scraper.Platform { return scraper.PlatformTokopedia }

func (a *TokopediaAdapter) FetchReviews(ctx context.Context, job scraper.CrawlJob) ([]scraper.CrawlResult, error) {
	page, release, proxyURL, err := a.NavigateWithRetry(ctx, job.ProductURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		release()
		a.Rotator.ReportSuccess(proxyURL)
	}()

	delay := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	var results []scraper.CrawlResult

	wait := page.EachEvent(func(e *proto.NetworkResponseReceived) {
		if !tokopediaGraphQLPattern.MatchString(e.Response.URL) {
			return
		}
		body, err := page.GetResource(e.Response.URL)
		if err != nil || !strings.Contains(string(body), tokopediaReviewOperation) {
			return
		}
		raw, _ := json.Marshal(map[string]any{
			"source":    "graphql",
			"url":       e.Response.URL,
			"body":      string(body),
			"operation": tokopediaReviewOperation,
		})
		results = append(results, scraper.CrawlResult{
			Platform:   scraper.PlatformTokopedia,
			ProductURL: job.ProductURL,
			RawJSON:    raw,
			CrawledAt:  time.Now(),
		})
	})
	defer wait()

	_ = page.WaitIdle(5e9)

	if len(results) == 0 {
		results, err = a.tokopediaDOM(page, job)
	}
	return results, err
}

func (a *TokopediaAdapter) tokopediaDOM(page *rod.Page, job scraper.CrawlJob) ([]scraper.CrawlResult, error) {
	e := &selector.RodExtractor{Page: page}
	fields := []string{"review_text", "rating", "author_name", "author_id", "reviewed_at"}
	data := make(map[string]string, len(fields))
	for _, f := range fields {
		cfg := a.Selectors.Get(string(scraper.PlatformTokopedia), f)
		val, _ := selector.ExtractField(e, cfg)
		data[f] = val
	}
	raw, _ := json.Marshal(map[string]any{"source": "dom", "fields": data})
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformTokopedia,
		ProductURL: job.ProductURL,
		RawJSON:    raw,
		CrawledAt:  time.Now(),
	}}, nil
}
