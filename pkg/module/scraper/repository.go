package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"review-curator/pkg/platform/database"
)

type Repository interface {
	UpsertCrawlJob(ctx context.Context, job CrawlJob) error
	UpdateJobStatus(ctx context.Context, id, status string, errMsg *string) error
	GetJobByID(ctx context.Context, id string) (*CrawlJob, error)
	ListJobs(ctx context.Context, platform, status string, p database.Page) ([]CrawlJob, int, error)
	UpsertRawReview(ctx context.Context, r RawReview) error
	GetRawReviewByID(ctx context.Context, id string) (*RawReview, error)
}

type postgresRepository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) Repository { return &postgresRepository{db: db} }

func (r *postgresRepository) UpsertCrawlJob(ctx context.Context, job CrawlJob) error {
	q := `INSERT INTO crawl_jobs (id, platform, product_url, product_id, max_pages, status, enqueued_at)
          VALUES (:id, :platform, :product_url, :product_id, :max_pages, :status, :enqueued_at)
          ON CONFLICT (id) DO UPDATE SET
              status = EXCLUDED.status,
              retry_count = crawl_jobs.retry_count + 1`
	return database.UpsertOne(ctx, r.db, q, job)
}

func (r *postgresRepository) UpdateJobStatus(ctx context.Context, id, status string, errMsg *string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE crawl_jobs SET status=$1, error_msg=$2, completed_at=$3 WHERE id=$4`,
		status, errMsg, now, id)
	return err
}

func (r *postgresRepository) GetJobByID(ctx context.Context, id string) (*CrawlJob, error) {
	var job CrawlJob
	if err := r.db.GetContext(ctx, &job, `SELECT * FROM crawl_jobs WHERE id=$1`, id); err != nil {
		return nil, fmt.Errorf("repository: get job: %w", err)
	}
	return &job, nil
}

func (r *postgresRepository) ListJobs(ctx context.Context, platform, status string, p database.Page) ([]CrawlJob, int, error) {
	q := `SELECT * FROM crawl_jobs WHERE ($1='' OR platform=$1) AND ($2='' OR status=$2)`
	return database.PaginatedSelect[CrawlJob](ctx, r.db, q, []any{platform, status}, p)
}

func (r *postgresRepository) UpsertRawReview(ctx context.Context, rv RawReview) error {
	q := `INSERT INTO raw_reviews (id, job_id, platform, product_url, payload, dedupe_hash, crawled_at)
          VALUES (:id, :job_id, :platform, :product_url, :payload, :dedupe_hash, :crawled_at)
          ON CONFLICT (dedupe_hash) DO NOTHING`
	return database.UpsertOne(ctx, r.db, q, rv)
}

func (r *postgresRepository) GetRawReviewByID(ctx context.Context, id string) (*RawReview, error) {
	var rv RawReview
	if err := r.db.GetContext(ctx, &rv, `SELECT * FROM raw_reviews WHERE id=$1`, id); err != nil {
		return nil, fmt.Errorf("repository: get raw review: %w", err)
	}
	return &rv, nil
}
