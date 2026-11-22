package birdcategory

import (
	"context"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Service adalah kontrak service kategori burung.
type Service interface {
	List(ctx context.Context) ([]entities.BirdCategory, error)
	GetByID(ctx context.Context, id int64) (entities.BirdCategory, error)
}

// service adalah implementasi Service.
type service struct {
	repo Repository
}

// NewService membuat service dengan repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// List mengambil semua kategori burung.
func (s *service) List(ctx context.Context) ([]entities.BirdCategory, error) {
	return s.repo.List(ctx)
}

// GetByID mengambil kategori burung berdasarkan id.
func (s *service) GetByID(ctx context.Context, id int64) (entities.BirdCategory, error) {
	return s.repo.FindByID(ctx, id)
}
