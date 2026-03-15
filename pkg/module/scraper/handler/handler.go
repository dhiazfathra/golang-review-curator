package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"review-curator/pkg/module/scraper"
	"review-curator/pkg/module/scraper/adapters"
)

type CrawlHandler struct {
	service  *scraper.CrawlService
	registry *adapters.Registry
}

func NewCrawlHandler(service *scraper.CrawlService, registry *adapters.Registry) *CrawlHandler {
	return &CrawlHandler{service: service, registry: registry}
}

func (h *CrawlHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("crawl handler: unmarshal: %w", err)
	}

	job, err := h.service.GetJob(ctx, payload.JobID)
	if err != nil {
		return fmt.Errorf("crawl handler: get job: %w", err)
	}

	_ = h.service.UpdateJobStatus(ctx, job.ID, "running", nil)

	adapter, ok := h.registry.Get(job.Platform)
	if !ok {
		msg := fmt.Sprintf("no adapter for platform: %s", job.Platform)
		_ = h.service.MarkFailed(ctx, job.ID, msg)
		return fmt.Errorf("crawl handler: %s", msg)
	}

	results, err := adapter.FetchReviews(ctx, *job)
	if err != nil {
		_ = h.service.MarkFailed(ctx, job.ID, err.Error())
		return fmt.Errorf("crawl handler: fetch reviews: %w", err)
	}

	for _, r := range results {
		r.CrawledAt = time.Now()
		if err := h.service.CommitResult(ctx, job.ID, r); err != nil {
			return fmt.Errorf("crawl handler: commit: %w", err)
		}
	}

	return h.service.UpdateJobStatus(ctx, job.ID, "done", nil)
}
