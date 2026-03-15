package scraper

import "time"

type Platform string

const (
	PlatformShopee    Platform = "shopee"
	PlatformTokopedia Platform = "tokopedia"
	PlatformBlibli    Platform = "blibli"
)

const (
	TaskCrawlJob        = "crawl:job"
	TaskNormaliseReview = "normalise:review"
)

type CrawlJob struct {
	ID          string     `db:"id"`
	Platform    Platform   `db:"platform"`
	ProductURL  string     `db:"product_url"`
	ProductID   string     `db:"product_id"`
	MaxPages    int        `db:"max_pages"`
	Status      string     `db:"status"`
	RetryCount  int        `db:"retry_count"`
	EnqueuedAt  time.Time  `db:"enqueued_at"`
	StartedAt   *time.Time `db:"started_at"`
	CompletedAt *time.Time `db:"completed_at"`
	ErrorMsg    *string    `db:"error_msg"`
}

type RawReview struct {
	ID         string    `db:"id"`
	JobID      string    `db:"job_id"`
	Platform   Platform  `db:"platform"`
	ProductURL string    `db:"product_url"`
	Payload    []byte    `db:"payload"`
	DedupeHash string    `db:"dedupe_hash"`
	CrawledAt  time.Time `db:"crawled_at"`
}

type CrawlResult struct {
	ID         string
	Platform   Platform
	ProductURL string
	RawJSON    []byte
	CrawledAt  time.Time
}
