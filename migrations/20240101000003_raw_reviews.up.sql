-- +goose Up
CREATE TABLE raw_reviews (
    id          UUID PRIMARY KEY,
    job_id      UUID NOT NULL REFERENCES crawl_jobs(id),
    platform    TEXT NOT NULL,
    product_url TEXT NOT NULL,
    payload     JSONB NOT NULL,
    dedupe_hash TEXT NOT NULL UNIQUE,
    crawled_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_raw_reviews_job      ON raw_reviews(job_id);
CREATE INDEX idx_raw_reviews_platform ON raw_reviews(platform, crawled_at DESC);

-- +goose Down
DROP TABLE IF EXISTS raw_reviews;
