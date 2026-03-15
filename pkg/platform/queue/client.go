package queue

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type Client struct {
	inner *asynq.Client
}

func NewClient(redisURL string) *Client {
	return &Client{inner: asynq.NewClient(asynq.RedisClientOpt{Addr: redisURL})}
}

func (c *Client) EnqueueCrawlJob(jobID string) error {
	payload, _ := json.Marshal(map[string]string{"job_id": jobID})
	_, err := c.inner.Enqueue(
		asynq.NewTask("crawl:job", payload),
		asynq.Queue("crawl"),
		asynq.MaxRetry(3),
		asynq.Unique(0),
	)
	if err != nil {
		return fmt.Errorf("queue: enqueue crawl job: %w", err)
	}
	return nil
}

func (c *Client) EnqueueNormalise(rawReviewID string) error {
	payload, _ := json.Marshal(map[string]string{"raw_review_id": rawReviewID})
	_, err := c.inner.Enqueue(
		asynq.NewTask("normalise:review", payload),
		asynq.Queue("normalise"),
		asynq.MaxRetry(5),
	)
	if err != nil {
		return fmt.Errorf("queue: enqueue normalise: %w", err)
	}
	return nil
}

func (c *Client) Close() error { return c.inner.Close() }
