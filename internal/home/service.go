package home

import "context"

// Service menyediakan konten Beranda untuk handler.
type Service struct {
	repo Repository
}

// NewService membuat service konten Beranda.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Get mengambil konten Beranda.
func (s *Service) Get(ctx context.Context) (Content, error) {
	return s.repo.Get(ctx)
}
