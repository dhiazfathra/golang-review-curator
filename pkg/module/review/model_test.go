package review

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReview_Fields(t *testing.T) {
	now := time.Now()
	review := Review{
		ID:             "review-123",
		Platform:       "shopee",
		ProductID:      "prod-456",
		AuthorName:     "testuser",
		Rating:         5,
		ReviewText:     "Great product!",
		Language:       "en",
		SentimentScore: 1.0,
		ReviewedAt:     now,
	}

	assert.Equal(t, "review-123", review.ID)
	assert.Equal(t, "shopee", review.Platform)
	assert.Equal(t, "prod-456", review.ProductID)
	assert.Equal(t, "testuser", review.AuthorName)
	assert.Equal(t, 5, review.Rating)
	assert.Equal(t, "Great product!", review.ReviewText)
	assert.Equal(t, "en", review.Language)
	assert.Equal(t, 1.0, review.SentimentScore)
	assert.Equal(t, now, review.ReviewedAt)
}

func TestReview_DefaultValues(t *testing.T) {
	review := Review{}

	assert.Empty(t, review.ID)
	assert.Empty(t, review.Platform)
	assert.Empty(t, review.ProductID)
	assert.Empty(t, review.AuthorName)
	assert.Zero(t, review.Rating)
	assert.Empty(t, review.ReviewText)
	assert.Empty(t, review.Language)
	assert.Zero(t, review.SentimentScore)
}

func TestReview_AllRatings(t *testing.T) {
	tests := []struct {
		name   string
		rating int
	}{
		{"Rating 1", 1},
		{"Rating 2", 2},
		{"Rating 3", 3},
		{"Rating 4", 4},
		{"Rating 5", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := Review{Rating: tt.rating}
			assert.Equal(t, tt.rating, review.Rating)
		})
	}
}

func TestReview_SentimentRange(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected float64
	}{
		{"Positive", 0.8, 0.8},
		{"Neutral", 0.0, 0.0},
		{"Negative", -0.8, -0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			review := Review{SentimentScore: tt.score}
			assert.Equal(t, tt.expected, review.SentimentScore)
		})
	}
}

func TestSummary_Fields(t *testing.T) {
	summary := Summary{
		ProductID:    "prod-123",
		Platform:     "shopee",
		TotalCount:   100,
		AvgRating:    4.5,
		CountByStar:  map[int]int{1: 5, 2: 10, 3: 20, 4: 30, 5: 35},
		AvgSentiment: 0.75,
	}

	assert.Equal(t, "prod-123", summary.ProductID)
	assert.Equal(t, "shopee", summary.Platform)
	assert.Equal(t, 100, summary.TotalCount)
	assert.Equal(t, 4.5, summary.AvgRating)
	assert.Equal(t, 35, summary.CountByStar[5])
	assert.Equal(t, 0.75, summary.AvgSentiment)
}

func TestSummary_Empty(t *testing.T) {
	summary := Summary{}

	assert.Empty(t, summary.ProductID)
	assert.Zero(t, summary.TotalCount)
	assert.Zero(t, summary.AvgRating)
	assert.Nil(t, summary.CountByStar)
	assert.Zero(t, summary.AvgSentiment)
}

func TestSummary_CountByStar(t *testing.T) {
	summary := Summary{
		CountByStar: make(map[int]int),
	}

	summary.CountByStar[1] = 10
	summary.CountByStar[2] = 20
	summary.CountByStar[3] = 30
	summary.CountByStar[4] = 40
	summary.CountByStar[5] = 50

	assert.Equal(t, 10, summary.CountByStar[1])
	assert.Equal(t, 50, summary.CountByStar[5])
	assert.Equal(t, 150, summary.CountByStar[1]+summary.CountByStar[2]+summary.CountByStar[3]+summary.CountByStar[4]+summary.CountByStar[5])
}

func TestListFilter_Fields(t *testing.T) {
	now := time.Now()
	filter := ListFilter{
		Platform:  "tokopedia",
		ProductID: "prod-789",
		Rating:    5,
		Language:  "id",
		From:      now.Add(-7 * 24 * time.Hour),
		To:        now,
	}

	assert.Equal(t, "tokopedia", filter.Platform)
	assert.Equal(t, "prod-789", filter.ProductID)
	assert.Equal(t, 5, filter.Rating)
	assert.Equal(t, "id", filter.Language)
	assert.True(t, filter.From.Before(filter.To))
}

func TestListFilter_Empty(t *testing.T) {
	filter := ListFilter{}

	assert.Empty(t, filter.Platform)
	assert.Empty(t, filter.ProductID)
	assert.Zero(t, filter.Rating)
	assert.Empty(t, filter.Language)
}

func TestListFilter_TimeRange(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	filter := ListFilter{
		From: from,
		To:   to,
	}

	assert.True(t, filter.From.Before(filter.To))
	assert.Equal(t, from, filter.From)
	assert.Equal(t, to, filter.To)
}
