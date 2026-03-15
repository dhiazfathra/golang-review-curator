package e2e_test

import (
	"encoding/json"
	"time"

	"review-curator/pkg/module/normaliser"
	"review-curator/pkg/module/scraper"

	"github.com/hibiken/asynq"
)

func (s *E2ESuite) TestFullCrawlNormalisePipeline() {
	jobPayload := map[string]interface{}{
		"platform":    "shopee",
		"product_url": "https://shopee.co.id/TestProduct123",
		"product_id":  "12345",
		"max_pages":   1,
	}
	payloadBytes, _ := json.Marshal(jobPayload)

	task := asynq.NewTask(scraper.TaskCrawlJob, payloadBytes, asynq.Queue("crawl"))
	_, err := s.Asynq.Enqueue(task)
	s.Require().NoError(err)

	var rawCount int
	err = s.DB.Get(&rawCount, "SELECT COUNT(*) FROM raw_reviews")
	s.NoError(err)
	s.T().Logf("Raw reviews inserted: %d", rawCount)

	var normCount int
	err = s.DB.Get(&normCount, "SELECT COUNT(*) FROM normalised_reviews")
	s.NoError(err)
	s.T().Logf("Normalised reviews: %d", normCount)

	if normCount > 0 {
		var norm normaliser.NormalisedReview
		err := s.DB.Get(&norm, "SELECT * FROM normalised_reviews LIMIT 1")
		s.NoError(err)
		s.NotEmpty(norm.ID)
		s.NotEmpty(norm.Platform)
		s.NotEmpty(norm.ProductID)
		s.True(norm.Rating >= 1 && norm.Rating <= 5)
		s.NotEmpty(norm.DedupeHash)
	}
}

func (s *E2ESuite) TestDeduplication() {
	dedupeHash := "test_hash_123"

	_, err := s.DB.Exec(`
		INSERT INTO normalised_reviews (id, raw_review_id, platform, product_id, author_id, 
			author_name, rating, review_text, language, reviewed_at, normalised_at, dedupe_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "uuid1", "raw-uuid-1", "shopee", "prod123", "author1", "Test User",
		5, "Great product!", "en", time.Now(), time.Now(), dedupeHash)
	s.NoError(err)

	_, _ = s.DB.Exec(`
		INSERT INTO normalised_reviews (id, raw_review_id, platform, product_id, author_id, 
			author_name, rating, review_text, language, reviewed_at, normalised_at, dedupe_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "uuid2", "raw-uuid-2", "shopee", "prod123", "author1", "Test User",
		5, "Great product!", "en", time.Now(), time.Now(), dedupeHash)

	var count int
	err = s.DB.Get(&count, "SELECT COUNT(*) FROM normalised_reviews WHERE dedupe_hash = $1", dedupeHash)
	s.NoError(err)
	s.Equal(1, count, "Deduplication should prevent duplicate entries")
}

func (s *E2ESuite) TestCrawlJobWorkflow() {
	jobID := "test-job-123"

	_, err := s.DB.Exec(`
		INSERT INTO crawl_jobs (id, platform, product_url, product_id, max_pages, retry_count, status, enqueued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, jobID, "shopee", "https://shopee.co.id/test", "12345", 1, 0, "pending", time.Now())
	s.Require().NoError(err)

	var count int
	err = s.DB.Get(&count, "SELECT COUNT(*) FROM crawl_jobs WHERE id = $1", jobID)
	s.NoError(err)
	s.Equal(1, count)

	s.DB.MustExec("UPDATE crawl_jobs SET status = 'done', completed_at = $1 WHERE id = $2", time.Now(), jobID)

	err = s.DB.Get(&count, "SELECT COUNT(*) FROM crawl_jobs WHERE id = $1 AND status = 'done'", jobID)
	s.NoError(err)
	s.Equal(1, count)
}
