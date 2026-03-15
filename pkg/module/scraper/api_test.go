package scraper_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueueJobHandler_Returns201(t *testing.T) {
	e := echo.New()
	body := `{"platform":"shopee","product_url":"https://shopee.co.id/test","product_id":"123","max_pages":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/crawl/jobs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	_ = e.NewContext(req, rec)

	t.Skip("requires mock CrawlService — implement with mockery")
}

func TestListJobsHandler_Pagination(t *testing.T) {
	t.Skip("requires mock CrawlService")
}

func TestGetJobHandler_NotFound(t *testing.T) {
	t.Skip("requires mock CrawlService")
}

func TestRetryJob_OnlyFailedJobs(t *testing.T) {
	t.Skip("requires mock CrawlService — assert 400 for non-failed job")
}

func TestEnqueueJobIntegration(t *testing.T) {
	t.Skip("requires running server on localhost:8080")
	body := `{"platform":"shopee","product_url":"https://shopee.co.id/test","product_id":"abc","max_pages":3}`
	resp, err := http.Post("http://localhost:8080/api/v1/crawl/jobs",
		"application/json", strings.NewReader(body))
	require.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["id"])
}
