package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Page struct {
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

func PaginatedSelect[T any](ctx context.Context, db *sqlx.DB, baseQuery string, args []any, p Page) ([]T, int, error) {
	var total int
	countQ := "SELECT COUNT(*) FROM (" + baseQuery + ") _c"
	if err := db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("PaginatedSelect count: %w", err)
	}

	dir := strings.ToUpper(p.SortDir)
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}
	q := baseQuery
	if p.SortBy != "" {
		q += fmt.Sprintf(" ORDER BY %s %s", p.SortBy, dir)
	}
	q += fmt.Sprintf(" LIMIT %d OFFSET %d", p.Limit, p.Offset)

	var rows []T
	if err := db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, 0, fmt.Errorf("PaginatedSelect rows: %w", err)
	}
	return rows, total, nil
}

func UpsertOne(ctx context.Context, db *sqlx.DB, query string, arg any) error {
	_, err := db.NamedExecContext(ctx, query, arg)
	return err
}
