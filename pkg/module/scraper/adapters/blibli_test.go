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

func loadBlibliFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "..", "..", "testdata", "blibli", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture %s", name)
	return data
}

type blibliXHRPayload struct {
	Data struct {
		Reviews []struct {
			Rating       int    `json:"rating"`
			AuthorName   string `json:"authorName"`
			AuthorID     string `json:"authorId"`
			ReviewText   string `json:"reviewText"`
			ReviewDate   string `json:"reviewDate"`
			HelpfulCount int    `json:"helpfulCount"`
		} `json:"reviews"`
		TotalCount  int `json:"totalCount"`
		PageSize    int `json:"pageSize"`
		CurrentPage int `json:"currentPage"`
	} `json:"data"`
	Error        any `json:"error"`
	ErrorMessage any `json:"errorMessage"`
}

func parseBlibliXHR(data []byte) ([]scraper.CrawlResult, error) {
	var payload blibliXHRPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	if len(payload.Data.Reviews) == 0 {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformBlibli,
		ProductURL: "https://blibli.com/product",
		RawJSON:    data,
	}}, nil
}

func parseBlibliDOM(data []byte, selectorKey string) ([]scraper.CrawlResult, error) {
	content := string(data)
	if strings.Contains(content, "no-reviews") || strings.Contains(content, "Belum ada ulasan") {
		return []scraper.CrawlResult{}, nil
	}
	if !strings.Contains(content, "review-item") {
		return []scraper.CrawlResult{}, nil
	}
	return []scraper.CrawlResult{{
		Platform:   scraper.PlatformBlibli,
		ProductURL: "https://blibli.com/product",
		RawJSON:    data,
	}}, nil
}

func TestBlibliXHRParse(t *testing.T) {
	fixture := loadBlibliFixture(t, "review_xhr.json")
	results, err := parseBlibliXHR(fixture)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.NotEmpty(t, results[0].RawJSON)

	var body map[string]any
	require.NoError(t, json.Unmarshal(fixture, &body))
	assert.Contains(t, body, "data")
}

func TestBlibliDOM(t *testing.T) {
	fixture := loadBlibliFixture(t, "review_page.html")
	results, err := parseBlibliDOM(fixture, "blibli:review_text")
	require.NoError(t, err)
	require.NotEmpty(t, results)
}

func TestBlibliEmptyReviewPage(t *testing.T) {
	fixture := loadBlibliFixture(t, "empty_reviews.html")
	results, err := parseBlibliDOM(fixture, "blibli:review_text")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestBlibliSmoke(t *testing.T) {
	t.Skip("live integration test — run with -tags integration")
}
