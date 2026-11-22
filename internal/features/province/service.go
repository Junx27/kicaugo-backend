package province

import (
	"context"
	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Service adalah kontrak service province.
type Service interface {
	List(ctx context.Context) ([]entities.Province, error)
	// scaffold:service-interface
}

// service adalah implementasi Service.
type service struct {
	repo Repository
}

// NewService membuat service dengan repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// List mengambil semua province.
func (s *service) List(ctx context.Context) ([]entities.Province, error) {
	return s.repo.List(ctx)
}

// scaffold:service-methods
