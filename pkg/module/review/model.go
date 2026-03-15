package review

import "time"

type Review struct {
	ID             string    `json:"id" db:"id"`
	Platform       string    `json:"platform" db:"platform"`
	ProductID      string    `json:"product_id" db:"product_id"`
	AuthorName     string    `json:"author_name" db:"author_name"`
	Rating         int       `json:"rating" db:"rating"`
	ReviewText     string    `json:"review_text" db:"review_text"`
	Language       string    `json:"language" db:"language"`
	SentimentScore float64   `json:"sentiment_score" db:"sentiment_score"`
	ReviewedAt     time.Time `json:"reviewed_at" db:"reviewed_at"`
}

type Summary struct {
	ProductID    string      `json:"product_id"`
	Platform     string      `json:"platform"`
	TotalCount   int         `json:"total_count"`
	AvgRating    float64     `json:"avg_rating"`
	CountByStar  map[int]int `json:"count_by_star"`
	AvgSentiment float64     `json:"avg_sentiment"`
}

type ListFilter struct {
	Platform  string
	ProductID string
	Rating    int
	Language  string
	From      time.Time
	To        time.Time
}
