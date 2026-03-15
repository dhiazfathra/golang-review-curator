package review

import (
	"context"

	"review-curator/pkg/platform/database"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) List(ctx context.Context, f ListFilter, p database.Page) ([]Review, int, error) {
	if p.Limit == 0 {
		p.Limit = 20
	}
	if p.SortBy == "" {
		p.SortBy = "reviewed_at"
		p.SortDir = "desc"
	}
	return s.repo.List(ctx, f, p)
}

func (s *Service) GetSummary(ctx context.Context, platform, productID string) (*Summary, error) {
	return s.repo.GetSummary(ctx, platform, productID)
}
