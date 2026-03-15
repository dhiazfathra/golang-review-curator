package adapters

import (
	"context"
	"encoding/json"
	"regexp"
	"time"

	"review-curator/pkg/module/scraper"
	"review-curator/pkg/platform/selector"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var blibliReviewPattern = regexp.MustCompile(`blibli\.com/api/reviews/products/`)

type BlibliAdapter struct {
	BaseScraper
}

func (a *BlibliAdapter) Platform() scraper.Platform { return scraper.PlatformBlibli }

func (a *BlibliAdapter) FetchReviews(ctx context.Context, job scraper.CrawlJob) ([]scraper.CrawlResult, error) {
	page, release, proxyURL, err := a.NavigateWithRetry(ctx, job.ProductURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		release()
		a.Rotator.ReportSuccess(proxyURL)
	}()

	var results []scraper.CrawlResult

	wait := page.EachEvent(func(e *proto.NetworkResponseReceived) {
		if !blibliReviewPattern.MatchString(e.Response.URL) {
			return
		}
		body, err := page.GetResource(e.Response.URL)
		if err != nil {
			return
		}
		raw, _ := json.Marshal(map[string]any{
			"source": "xhr",
			"url":    e.Response.URL,
			"body":   string(body),
		})
		results = append(results, scraper.CrawlResult{
			Platform:   scraper.PlatformBlibli,
			ProductURL: job.ProductURL,
			RawJSON:    raw,
			CrawledAt:  time.Now(),
		})
	})
	defer wait()

	_ = page.WaitIdle(3e9)

	if len(results) == 0 {
		results, err = a.blibliDOM(page, job)
	}
	return results, err
}

func (a *BlibliAdapter) blibliDOM(page *rod.Page, job scraper.CrawlJob) ([]scraper.CrawlResult, error) {
	e := &selector.RodExtractor{Page: page}
	fields := []string{"review_text", "rating", "author_name", "author_id", "reviewed_at"}
	data := make(map[string]string, len(fields))
	for _, f := range fields {
		cfg := a.Selectors.Get(string(scraper.PlatformBlibli), f)
		val, _ := selector.ExtractField(e, cfg)
		data[f] = val
	}
	raw, _ := json.Marshal(map[string]any{"source": "dom", "fields": data})
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformBlibli,
		ProductURL: job.ProductURL,
		RawJSON:    raw,
		CrawledAt:  time.Now(),
	}}, nil
}
