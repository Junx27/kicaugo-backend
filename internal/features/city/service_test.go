package city

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeRepository struct {
	result     []entities.City
	err        error
	provinceID int64
	query      string
}

func (f *fakeRepository) ListByProvince(ctx context.Context, provinceID int64, q string) ([]entities.City, error) {
	f.provinceID = provinceID
	f.query = q
	return f.result, f.err
}

func TestServiceListByProvince(t *testing.T) {
	t.Run("berhasil dan memangkas spasi", func(t *testing.T) {
		repo := &fakeRepository{result: []entities.City{{ID: 1}}}
		got, err := NewService(repo).ListByProvince(context.Background(), 11, "  banda  ")
		if err != nil || len(got) != 1 {
			t.Errorf("ListByProvince() = %+v, %v", got, err)
		}
		if repo.query != "banda" || repo.provinceID != 11 {
			t.Errorf("repo menerima provinceID=%d q=%q", repo.provinceID, repo.query)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		svc := NewService(&fakeRepository{err: errors.New("db down")})
		if _, err := svc.ListByProvince(context.Background(), 11, ""); err == nil {
			t.Error("ListByProvince() ingin error")
		}
	})
}
