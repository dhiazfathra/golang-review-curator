package scraper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlatform_Constants(t *testing.T) {
	assert.Equal(t, Platform("shopee"), PlatformShopee)
	assert.Equal(t, Platform("tokopedia"), PlatformTokopedia)
	assert.Equal(t, Platform("blibli"), PlatformBlibli)
}

func TestTaskNames(t *testing.T) {
	assert.Equal(t, "crawl:job", TaskCrawlJob)
	assert.Equal(t, "normalise:review", TaskNormaliseReview)
}

func TestCrawlJob_Fields(t *testing.T) {
	now := time.Now()
	errMsg := "test error"
	job := CrawlJob{
		ID:          "test-id",
		Platform:    PlatformShopee,
		ProductURL:  "https://shopee.co.id/product/123",
		ProductID:   "123",
		MaxPages:    10,
		Status:      "pending",
		RetryCount:  0,
		EnqueuedAt:  now,
		StartedAt:   nil,
		CompletedAt: nil,
		ErrorMsg:    nil,
	}

	assert.Equal(t, "test-id", job.ID)
	assert.Equal(t, PlatformShopee, job.Platform)
	assert.Equal(t, "https://shopee.co.id/product/123", job.ProductURL)
	assert.Equal(t, "123", job.ProductID)
	assert.Equal(t, 10, job.MaxPages)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, 0, job.RetryCount)
	assert.Equal(t, now, job.EnqueuedAt)
	assert.Nil(t, job.StartedAt)
	assert.Nil(t, job.CompletedAt)
	assert.Nil(t, job.ErrorMsg)

	job.Status = "failed"
	job.ErrorMsg = &errMsg
	assert.Equal(t, "failed", job.Status)
	assert.NotNil(t, job.ErrorMsg)
	assert.Equal(t, "test error", *job.ErrorMsg)
}

func TestCrawlJob_WithTimestamps(t *testing.T) {
	now := time.Now()
	startedAt := now.Add(-time.Hour)
	completedAt := now

	job := CrawlJob{
		ID:          "test-id",
		Status:      "done",
		StartedAt:   &startedAt,
		CompletedAt: &completedAt,
	}

	assert.NotNil(t, job.StartedAt)
	assert.NotNil(t, job.CompletedAt)
	assert.True(t, job.CompletedAt.After(*job.StartedAt))
}

func TestRawReview_Fields(t *testing.T) {
	now := time.Now()
	review := RawReview{
		ID:         "review-id",
		JobID:      "job-id",
		Platform:   PlatformTokopedia,
		ProductURL: "https://tokopedia.com/product/456",
		Payload:    []byte(`{"review": "test"}`),
		DedupeHash: "abc123",
		CrawledAt:  now,
	}

	assert.Equal(t, "review-id", review.ID)
	assert.Equal(t, "job-id", review.JobID)
	assert.Equal(t, PlatformTokopedia, review.Platform)
	assert.Equal(t, "https://tokopedia.com/product/456", review.ProductURL)
	assert.Equal(t, []byte(`{"review": "test"}`), review.Payload)
	assert.Equal(t, "abc123", review.DedupeHash)
	assert.Equal(t, now, review.CrawledAt)
}

func TestCrawlResult_Fields(t *testing.T) {
	now := time.Now()
	result := CrawlResult{
		ID:         "result-id",
		Platform:   PlatformBlibli,
		ProductURL: "https://blibli.com/product/789",
		RawJSON:    []byte(`{"reviews": []}`),
		CrawledAt:  now,
	}

	assert.Equal(t, "result-id", result.ID)
	assert.Equal(t, PlatformBlibli, result.Platform)
	assert.Equal(t, "https://blibli.com/product/789", result.ProductURL)
	assert.Equal(t, []byte(`{"reviews": []}`), result.RawJSON)
	assert.Equal(t, now, result.CrawledAt)
}

func TestDedupeHash(t *testing.T) {
	result := CrawlResult{
		ID:         "result-id",
		Platform:   PlatformShopee,
		ProductURL: "https://shopee.co.id/product/123",
		RawJSON:    []byte(`{"text": "great product"}`),
		CrawledAt:  time.Now(),
	}

	hash := dedupeHash(result)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)

	result2 := result
	result2.RawJSON = []byte(`{"text": "different"}`)
	hash2 := dedupeHash(result2)
	assert.NotEqual(t, hash, hash2)

	result3 := result
	hash3 := dedupeHash(result3)
	assert.Equal(t, hash, hash3)
}

func TestPlatform_String(t *testing.T) {
	assert.Equal(t, "shopee", string(PlatformShopee))
	assert.Equal(t, "tokopedia", string(PlatformTokopedia))
	assert.Equal(t, "blibli", string(PlatformBlibli))
}

func TestCrawlJob_DefaultValues(t *testing.T) {
	job := CrawlJob{}

	assert.Empty(t, job.ID)
	assert.Empty(t, job.Platform)
	assert.Empty(t, job.Status)
	assert.Zero(t, job.MaxPages)
	assert.Zero(t, job.RetryCount)
}

func TestRawReview_EmptyPayload(t *testing.T) {
	review := RawReview{
		Payload: []byte{},
	}

	assert.Empty(t, review.Payload)
}
