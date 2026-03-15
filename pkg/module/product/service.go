package product

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Register(ctx context.Context, name, platform, productURL, productID string) (*Product, error) {
	p := Product{
		ID:         uuid.New().String(),
		Name:       name,
		Platform:   platform,
		ProductURL: productURL,
		ProductID:  productID,
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := s.repo.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repo.List(ctx)
}
