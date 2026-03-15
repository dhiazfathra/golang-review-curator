package normaliser

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"review-curator/pkg/platform/database"
)

type Repository interface {
	UpsertNormalisedReview(ctx context.Context, r NormalisedReview) error
	GetByProductID(ctx context.Context, platform, productID string, p database.Page) ([]NormalisedReview, int, error)
	GetByID(ctx context.Context, id string) (*NormalisedReview, error)
}

type postgresRepo struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) UpsertNormalisedReview(ctx context.Context, rv NormalisedReview) error {
	q := `INSERT INTO normalised_reviews
		(id, raw_review_id, platform, product_id, author_id, author_name, rating,
		 review_text, language, sentiment_score, reviewed_at, normalised_at, dedupe_hash)
		VALUES
		(:id, :raw_review_id, :platform, :product_id, :author_id, :author_name, :rating,
		 :review_text, :language, :sentiment_score, :reviewed_at, :normalised_at, :dedupe_hash)
		ON CONFLICT (dedupe_hash) DO UPDATE SET
			sentiment_score = EXCLUDED.sentiment_score,
			language = EXCLUDED.language,
			normalised_at = NOW()`
	return database.UpsertOne(ctx, r.db, q, rv)
}

func (r *postgresRepo) GetByProductID(ctx context.Context, platform, productID string, p database.Page) ([]NormalisedReview, int, error) {
	q := `SELECT * FROM normalised_reviews WHERE platform=$1 AND product_id=$2`
	return database.PaginatedSelect[NormalisedReview](ctx, r.db, q, []any{platform, productID}, p)
}

func (r *postgresRepo) GetByID(ctx context.Context, id string) (*NormalisedReview, error) {
	var rv NormalisedReview
	if err := r.db.GetContext(ctx, &rv, `SELECT * FROM normalised_reviews WHERE id=$1`, id); err != nil {
		return nil, fmt.Errorf("normaliser repo: get by id: %w", err)
	}
	return &rv, nil
}
