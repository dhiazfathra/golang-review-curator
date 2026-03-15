package normaliser

import "time"

type NormalisedReview struct {
	ID             string    `db:"id"`
	RawReviewID    string    `db:"raw_review_id"`
	Platform       string    `db:"platform"`
	ProductID      string    `db:"product_id"`
	AuthorID       string    `db:"author_id"`
	AuthorName     string    `db:"author_name"`
	Rating         int       `db:"rating"`
	ReviewText     string    `db:"review_text"`
	Language       string    `db:"language"`
	SentimentScore float64   `db:"sentiment_score"`
	ReviewedAt     time.Time `db:"reviewed_at"`
	NormalisedAt   time.Time `db:"normalised_at"`
	DedupeHash     string    `db:"dedupe_hash"`
}

type NormalisationResult struct {
	Review NormalisedReview
	Err    error
}
