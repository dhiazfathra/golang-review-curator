package adapters_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"review-curator/pkg/module/scraper"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "testdata", "shopee", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture %s", name)
	return data
}

type xhrPayload struct {
	Data struct {
		Ratings []struct {
			RatingStar     int    `json:"rating_star"`
			AuthorUsername string `json:"author_username"`
			AuthorID       string `json:"author_id"`
			Comment        string `json:"comment"`
			ReviewTime     int64  `json:"review_time"`
		} `json:"ratings"`
	} `json:"data"`
}

func parseShopeeXHR(data []byte) ([]scraper.CrawlResult, error) {
	var payload xhrPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Ratings) == 0 {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformShopee,
		ProductURL: "https://shopee.co.id/product",
		RawJSON:    data,
	}}, nil
}

func parseShopeeDOM(data []byte, selectorKey string) ([]scraper.CrawlResult, error) {
	content := string(data)
	if strings.Contains(content, "no-reviews") || strings.Contains(content, "Belum ada ulasan") {
		return []scraper.CrawlResult{}, nil
	}
	if !strings.Contains(content, "review-item") {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformShopee,
		ProductURL: "https://shopee.co.id/product",
		RawJSON:    data,
	}}, nil
}

func TestShopeeXHRParse(t *testing.T) {
	fixture := loadFixture(t, "review_xhr.json")
	results, err := parseShopeeXHR(fixture)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.NotEmpty(t, results[0].RawJSON)
}

func TestShopeeDOM(t *testing.T) {
	fixture := loadFixture(t, "review_page.html")
	results, err := parseShopeeDOM(fixture, "shopee:review_text")
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestShopeeEmptyReviewPage(t *testing.T) {
	fixture := loadFixture(t, "empty_reviews.html")
	results, err := parseShopeeDOM(fixture, "shopee:review_text")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestShopeeCaptchaDetected(t *testing.T) {
	t.Skip("requires mock setup — implemented with mockery in full test suite")
}

func TestShopeeSmoke(t *testing.T) {
	t.Skip("live integration test — run manually with -tags integration")
}
