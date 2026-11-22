package city

import (
	"context"
	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Service adalah kontrak service city.
type Service interface {
	List(ctx context.Context) ([]entities.City, error)
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

// List mengambil semua city.
func (s *service) List(ctx context.Context) ([]entities.City, error) {
	return s.repo.List(ctx)
}

// scaffold:service-methods
