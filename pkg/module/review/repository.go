package review

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"review-curator/pkg/platform/database"
)

type Repository interface {
	List(ctx context.Context, f ListFilter, p database.Page) ([]Review, int, error)
	GetSummary(ctx context.Context, platform, productID string) (*Summary, error)
}

type postgresRepo struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) List(ctx context.Context, f ListFilter, p database.Page) ([]Review, int, error) {
	var conditions []string
	var args []any
	i := 1
	add := func(cond string, val any) {
		conditions = append(conditions, fmt.Sprintf(cond, i))
		args = append(args, val)
		i++
	}
	if f.Platform != "" {
		add("platform = $%d", f.Platform)
	}
	if f.ProductID != "" {
		add("product_id = $%d", f.ProductID)
	}
	if f.Rating > 0 {
		add("rating = $%d", f.Rating)
	}
	if f.Language != "" {
		add("language = $%d", f.Language)
	}
	if !f.From.IsZero() {
		add("reviewed_at >= $%d", f.From)
	}
	if !f.To.IsZero() {
		add("reviewed_at <= $%d", f.To)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}
	q := fmt.Sprintf("SELECT id, platform, product_id, author_name, rating, review_text, language, sentiment_score, reviewed_at FROM normalised_reviews %s", where)
	return database.PaginatedSelect[Review](ctx, r.db, q, args, p)
}

func (r *postgresRepo) GetSummary(ctx context.Context, platform, productID string) (*Summary, error) {
	var row struct {
		TotalCount   int     `db:"total_count"`
		AvgRating    float64 `db:"avg_rating"`
		AvgSentiment float64 `db:"avg_sentiment"`
	}
	err := r.db.GetContext(ctx, &row,
		`SELECT COUNT(*) AS total_count,
                COALESCE(AVG(rating), 0) AS avg_rating,
                COALESCE(AVG(sentiment_score), 0) AS avg_sentiment
         FROM normalised_reviews WHERE platform=$1 AND product_id=$2`,
		platform, productID)
	if err != nil {
		return nil, fmt.Errorf("review repo: summary: %w", err)
	}

	var starRows []struct {
		Rating int `db:"rating"`
		Count  int `db:"cnt"`
	}
	_ = r.db.SelectContext(ctx, &starRows,
		`SELECT rating, COUNT(*) AS cnt FROM normalised_reviews
		 WHERE platform=$1 AND product_id=$2 GROUP BY rating`, platform, productID)

	countByStar := make(map[int]int, 5)
	for _, s := range starRows {
		countByStar[s.Rating] = s.Count
	}

	return &Summary{
		ProductID:    productID,
		Platform:     platform,
		TotalCount:   row.TotalCount,
		AvgRating:    row.AvgRating,
		CountByStar:  countByStar,
		AvgSentiment: row.AvgSentiment,
	}, nil
}
