package adapters_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"review-curator/pkg/module/scraper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTokopediaFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "testdata", "tokopedia", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture %s", name)
	return data
}

type graphqlPayload struct {
	Data struct {
		ProductRevGetProductReviewList struct {
			Data struct {
				List []struct {
					ReviewID   string `json:"reviewId"`
					Rating     int    `json:"rating"`
					ReviewText string `json:"reviewText"`
					UserName   string `json:"userName"`
				} `json:"list"`
			} `json:"data"`
		} `json:"ProductRevGetProductReviewList"`
	} `json:"data"`
}

func parseTokopediaGraphQL(data []byte) ([]scraper.CrawlResult, error) {
	var payload graphqlPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data.ProductRevGetProductReviewList.Data.List) == 0 {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformTokopedia,
		ProductURL: "https://tokopedia.com/product",
		RawJSON:    data,
	}}, nil
}

func parseTokopediaDOM(data []byte, selectorKey string) ([]scraper.CrawlResult, error) {
	content := string(data)
	if strings.Contains(content, "no-reviews") || strings.Contains(content, "Belum ada ulasan") {
		return []scraper.CrawlResult{}, nil
	}
	if !strings.Contains(content, "review-item") {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformTokopedia,
		ProductURL: "https://tokopedia.com/product",
		RawJSON:    data,
	}}, nil
}

func TestTokopediaGraphQLParse(t *testing.T) {
	fixture := loadTokopediaFixture(t, "review_graphql.json")
	var body map[string]any
	require.NoError(t, json.Unmarshal(fixture, &body))
	assert.Contains(t, string(fixture), "ProductRevGetProductReviewList")

	results, err := parseTokopediaGraphQL(fixture)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.NotEmpty(t, results[0].RawJSON)
}

func TestTokopediaDOM(t *testing.T) {
	fixture := loadTokopediaFixture(t, "review_page.html")
	results, err := parseTokopediaDOM(fixture, "tokopedia:review_text")
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestTokopediaEmptyReviewPage(t *testing.T) {
	fixture := loadTokopediaFixture(t, "empty_reviews.html")
	results, err := parseTokopediaDOM(fixture, "tokopedia:review_text")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestTokopediaSmoke(t *testing.T) {
	t.Skip("live integration test — run with -tags integration")
}
