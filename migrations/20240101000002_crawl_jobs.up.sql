-- +goose Up
CREATE TABLE crawl_jobs (
    id           UUID PRIMARY KEY,
    platform     TEXT NOT NULL,
    product_url  TEXT NOT NULL,
    product_id   TEXT NOT NULL,
    max_pages    INT NOT NULL DEFAULT 10,
    status       TEXT NOT NULL DEFAULT 'pending',
    retry_count  INT NOT NULL DEFAULT 0,
    enqueued_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_msg    TEXT
);
CREATE INDEX idx_crawl_jobs_status  ON crawl_jobs(status, enqueued_at);
CREATE INDEX idx_crawl_jobs_product ON crawl_jobs(platform, product_id);

-- +goose Down
DROP TABLE IF EXISTS crawl_jobs;
