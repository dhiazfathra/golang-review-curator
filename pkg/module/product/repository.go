package product

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery --name=Repository --output=../../../internal/mocks --outpkg=mocks
type Repository interface {
	ListActive(ctx context.Context) ([]Product, error)
	Upsert(ctx context.Context, p Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context) ([]Product, error)
}

type postgresRepo struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) Repository { return &postgresRepo{db: db} }

func (r *postgresRepo) ListActive(ctx context.Context) ([]Product, error) {
	var products []Product
	err := r.db.SelectContext(ctx, &products,
		`SELECT * FROM products WHERE active = true ORDER BY created_at ASC`)
	return products, err
}

func (r *postgresRepo) Upsert(ctx context.Context, p Product) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO products (id, name, platform, product_url, product_id, active, created_at, updated_at)
		VALUES (:id, :name, :platform, :product_url, :product_id, :active, :created_at, :updated_at)
		ON CONFLICT (product_url) DO UPDATE SET
			name = EXCLUDED.name, active = EXCLUDED.active, updated_at = NOW()`, p)
	return err
}

func (r *postgresRepo) GetByID(ctx context.Context, id string) (*Product, error) {
	var p Product
	if err := r.db.GetContext(ctx, &p, `SELECT * FROM products WHERE id=$1`, id); err != nil {
		return nil, fmt.Errorf("product repo: get by id: %w", err)
	}
	return &p, nil
}

func (r *postgresRepo) List(ctx context.Context) ([]Product, error) {
	var products []Product
	err := r.db.SelectContext(ctx, &products, `SELECT * FROM products ORDER BY created_at DESC`)
	return products, err
}
