package scraper

import (
	"context"
	"fmt"
	"time"

	"review-curator/pkg/platform/database"
	"review-curator/pkg/platform/queue"

	"github.com/google/uuid"
)

type CrawlService struct {
	repo  Repository
	queue *queue.Client
}

func NewCrawlService(repo Repository, q *queue.Client) *CrawlService {
	return &CrawlService{repo: repo, queue: q}
}

func (s *CrawlService) EnqueueJob(ctx context.Context, platform Platform, productURL, productID string, maxPages int) (*CrawlJob, error) {
	job := CrawlJob{
		ID:         uuid.New().String(),
		Platform:   platform,
		ProductURL: productURL,
		ProductID:  productID,
		MaxPages:   maxPages,
		Status:     "pending",
		EnqueuedAt: time.Now(),
	}
	if err := s.repo.UpsertCrawlJob(ctx, job); err != nil {
		return nil, fmt.Errorf("crawl service: upsert job: %w", err)
	}
	if err := s.queue.EnqueueCrawlJob(job.ID); err != nil {
		return nil, fmt.Errorf("crawl service: enqueue: %w", err)
	}
	return &job, nil
}

func (s *CrawlService) CommitResult(ctx context.Context, jobID string, result CrawlResult) error {
	rv := RawReview{
		ID:         uuid.New().String(),
		JobID:      jobID,
		Platform:   result.Platform,
		ProductURL: result.ProductURL,
		Payload:    result.RawJSON,
		DedupeHash: dedupeHash(result),
		CrawledAt:  result.CrawledAt,
	}
	if err := s.repo.UpsertRawReview(ctx, rv); err != nil {
		return fmt.Errorf("crawl service: upsert raw review: %w", err)
	}
	return s.queue.EnqueueNormalise(rv.ID)
}

func (s *CrawlService) MarkFailed(ctx context.Context, jobID, errMsg string) error {
	msg := errMsg
	return s.repo.UpdateJobStatus(ctx, jobID, "failed", &msg)
}

func (s *CrawlService) UpdateJobStatus(ctx context.Context, jobID, status string, errMsg *string) error {
	return s.repo.UpdateJobStatus(ctx, jobID, status, errMsg)
}

func (s *CrawlService) GetJob(ctx context.Context, jobID string) (*CrawlJob, error) {
	return s.repo.GetJobByID(ctx, jobID)
}

func (s *CrawlService) ListJobs(ctx context.Context, platform, status string, p database.Page) ([]CrawlJob, int, error) {
	return s.repo.ListJobs(ctx, platform, status, p)
}
