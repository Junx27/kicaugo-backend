package city

import (
	"context"
	"strings"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Service adalah kontrak service city.
type Service interface {
	ListByProvince(ctx context.Context, provinceID int64, q string) ([]entities.City, error)
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

// ListByProvince mengambil kabupaten/kota pada provinsi tertentu.
func (s *service) ListByProvince(ctx context.Context, provinceID int64, q string) ([]entities.City, error) {
	return s.repo.ListByProvince(ctx, provinceID, strings.TrimSpace(q))
}

// scaffold:service-methods
