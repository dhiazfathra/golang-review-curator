-- +goose Up
CREATE TABLE normalised_reviews (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_review_id   UUID NOT NULL REFERENCES raw_reviews(id),
    platform        TEXT NOT NULL,
    product_id      TEXT NOT NULL,
    author_id       TEXT NOT NULL,
    author_name     TEXT NOT NULL,
    rating          SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review_text     TEXT NOT NULL,
    language        TEXT NOT NULL DEFAULT 'unknown',
    sentiment_score NUMERIC(4,3),
    reviewed_at     TIMESTAMPTZ NOT NULL,
    normalised_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dedupe_hash     TEXT NOT NULL UNIQUE
);
CREATE INDEX idx_norm_reviews_product ON normalised_reviews(platform, product_id, reviewed_at DESC);
CREATE INDEX idx_norm_reviews_rating  ON normalised_reviews(platform, product_id, rating);

-- +goose Down
DROP TABLE IF EXISTS normalised_reviews;
