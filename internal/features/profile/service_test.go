package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

func stringPtr(v string) *string    { return &v }
func int64Ptr(v int64) *int64       { return &v }
func intPtr(v int) *int             { return &v }
func float64Ptr(v float64) *float64 { return &v }

type fakeRepository struct {
	resolvedID int64
	resolveErr error
	profile    entities.UserProfile
	ensureErr  error
	favorites  []entities.BirdCategory
	favErr     error
	city       entities.City
	cityErr    error
	validCount int
	countErr   error
	updateErr  error
	locErr     error
	replaceErr error

	resolveCalls int
	cityCalls    int
	updatedIn    entities.ProfileUpdate
	location     struct {
		latitude  *float64
		longitude *float64
		accuracy  *int
	}
	replacedIDs []int64
}

func (f *fakeRepository) ResolveUserID(ctx context.Context, identifier string) (int64, error) {
	f.resolveCalls++
	return f.resolvedID, f.resolveErr
}

func (f *fakeRepository) EnsureAndGet(ctx context.Context, userID int64) (entities.UserProfile, error) {
	return f.profile, f.ensureErr
}

func (f *fakeRepository) FavoriteCategories(ctx context.Context, userID int64) ([]entities.BirdCategory, error) {
	return f.favorites, f.favErr
}

func (f *fakeRepository) City(ctx context.Context, cityID int64) (entities.City, error) {
	f.cityCalls++
	return f.city, f.cityErr
}

func (f *fakeRepository) CountValidCategories(ctx context.Context, ids []int64) (int, error) {
	return f.validCount, f.countErr
}

func (f *fakeRepository) Update(ctx context.Context, userID int64, in entities.ProfileUpdate) error {
	f.updatedIn = in
	return f.updateErr
}

func (f *fakeRepository) UpdateLocation(ctx context.Context, userID int64, latitude, longitude *float64, accuracyM *int) error {
	f.location.latitude = latitude
	f.location.longitude = longitude
	f.location.accuracy = accuracyM
	return f.locErr
}

func (f *fakeRepository) ReplaceFavorites(ctx context.Context, userID int64, ids []int64) error {
	f.replacedIDs = ids
	return f.replaceErr
}

func TestServiceGet(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		cityID := int64(9)
		repo := &fakeRepository{
			resolvedID: 7,
			profile:    entities.UserProfile{FullName: "Budi", CityID: &cityID},
			favorites:  []entities.BirdCategory{{ID: 2, Name: "Murai"}},
		}
		view, err := NewService(repo).Get(context.Background(), "budi@example.com")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if !view.IsComplete || len(view.Favorites) != 1 || view.Profile.UserID != 0 {
			t.Errorf("view = %+v", view)
		}
	})

	t.Run("gagal mencari user", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if _, err := svc.Get(context.Background(), "x"); !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("Get() = %v, ingin ErrNotFound", err)
		}
	})

	t.Run("gagal membuat profil", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, ensureErr: errors.New("db down")})
		if _, err := svc.Get(context.Background(), "x"); err == nil {
			t.Error("Get() ingin error")
		}
	})

	t.Run("gagal mengambil favorit", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, favErr: errors.New("db down")})
		if _, err := svc.Get(context.Background(), "x"); err == nil {
			t.Error("Get() ingin error")
		}
	})
}

func TestServiceUpdate(t *testing.T) {
	t.Run("berhasil dan menormalkan telepon", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, city: entities.City{ProvinceID: 11}}
		in := entities.ProfileUpdate{
			FullName: stringPtr("Budi"),
			Phone:    stringPtr("0812-3456"),
			CityID:   int64Ptr(5),
		}
		if err := NewService(repo).Update(context.Background(), "x", in, int64Ptr(11)); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if got := *repo.updatedIn.Phone; got != "+628123456" {
			t.Errorf("phone tersimpan = %q", got)
		}
		if repo.cityCalls != 1 {
			t.Errorf("cityCalls = %d, ingin 1", repo.cityCalls)
		}
	})

	t.Run("provinsi tidak cocok", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, city: entities.City{ProvinceID: 12}}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{CityID: int64Ptr(5)}, int64Ptr(11))
		if !errors.Is(err, apperrors.ErrProvinceMismatch) {
			t.Errorf("Update() = %v, ingin ErrProvinceMismatch", err)
		}
	})

	t.Run("kota tidak valid", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, cityErr: apperrors.ErrInvalidCity}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{CityID: int64Ptr(5)}, nil)
		if !errors.Is(err, apperrors.ErrInvalidCity) {
			t.Errorf("Update() = %v, ingin ErrInvalidCity", err)
		}
	})

	t.Run("hanya provinsi tanpa kota profil", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, profile: entities.UserProfile{}}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{FullName: stringPtr("Budi")}, int64Ptr(11))
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
	})

	t.Run("gagal membaca kota profil", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, ensureErr: errors.New("db down")}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{Bio: stringPtr("hobi")}, int64Ptr(11))
		if err == nil {
			t.Error("Update() ingin error saat profil gagal dibaca")
		}
	})

	t.Run("validasi kota profil saat tidak ada payload kota", func(t *testing.T) {
		cityID := int64(6)
		repo := &fakeRepository{
			resolvedID: 1,
			profile:    entities.UserProfile{CityID: &cityID},
			city:       entities.City{ProvinceID: 11},
		}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{Bio: stringPtr("hobi")}, int64Ptr(11))
		if err != nil || repo.cityCalls != 1 {
			t.Errorf("Update() = %v, cityCalls = %d", err, repo.cityCalls)
		}
	})

	t.Run("tanpa perubahan kota tidak memanggilCity", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1}
		err := NewService(repo).Update(context.Background(), "x", entities.ProfileUpdate{Bio: stringPtr("hobi")}, nil)
		if err != nil || repo.cityCalls != 0 {
			t.Errorf("Update() = %v, cityCalls = %d", err, repo.cityCalls)
		}
	})

	t.Run("gagal mencari user", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if err := svc.Update(context.Background(), "x", entities.ProfileUpdate{}, nil); !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("Update() = %v", err)
		}
	})

	t.Run("gagal menyimpan", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, updateErr: errors.New("db down")})
		if err := svc.Update(context.Background(), "x", entities.ProfileUpdate{Bio: stringPtr("hobi")}, nil); err == nil {
			t.Error("Update() ingin error")
		}
	})
}

func TestServiceUpdateLocation(t *testing.T) {
	t.Run("berhasil menyimpan", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1}
		err := NewService(repo).UpdateLocation(context.Background(), "x", float64Ptr(-6.2), float64Ptr(106.8), intPtr(20))
		if err != nil {
			t.Fatalf("UpdateLocation() error = %v", err)
		}
		if *repo.location.latitude != -6.2 || *repo.location.longitude != 106.8 || *repo.location.accuracy != 20 {
			t.Errorf("location = %+v", repo.location)
		}
	})

	t.Run("berhasil menghapus lokasi", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1}
		if err := NewService(repo).UpdateLocation(context.Background(), "x", nil, nil, nil); err != nil {
			t.Errorf("UpdateLocation() error = %v", err)
		}
		if repo.location.latitude != nil {
			t.Error("latitude ingin nil")
		}
	})

	t.Run("koordinat tidak valid ditolak sebelum lookup", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1}
		err := NewService(repo).UpdateLocation(context.Background(), "x", float64Ptr(-6.2), nil, nil)
		if !errors.Is(err, apperrors.ErrInvalidCoordinates) || repo.resolveCalls != 0 {
			t.Errorf("UpdateLocation() = %v, resolveCalls = %d", err, repo.resolveCalls)
		}
	})

	t.Run("gagal mencari user", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if err := svc.UpdateLocation(context.Background(), "x", nil, nil, nil); !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("UpdateLocation() = %v", err)
		}
	})

	t.Run("gagal menyimpan", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, locErr: errors.New("db down")})
		if err := svc.UpdateLocation(context.Background(), "x", nil, nil, nil); err == nil {
			t.Error("UpdateLocation() ingin error")
		}
	})
}

func TestServiceReplaceFavorites(t *testing.T) {
	t.Run("berhasil dan menghapus duplikat", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, validCount: 2}
		err := NewService(repo).ReplaceFavorites(context.Background(), "x", []int64{1, 2, 1})
		if err != nil {
			t.Fatalf("ReplaceFavorites() error = %v", err)
		}
		if len(repo.replacedIDs) != 2 {
			t.Errorf("replacedIDs = %v", repo.replacedIDs)
		}
	})

	t.Run("kategori tidak valid", func(t *testing.T) {
		repo := &fakeRepository{resolvedID: 1, validCount: 1}
		err := NewService(repo).ReplaceFavorites(context.Background(), "x", []int64{1, 2})
		if !errors.Is(err, apperrors.ErrInvalidCategory) {
			t.Errorf("ReplaceFavorites() = %v", err)
		}
	})

	t.Run("gagal menghitung kategori", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, countErr: errors.New("db down")})
		if err := svc.ReplaceFavorites(context.Background(), "x", []int64{1}); err == nil {
			t.Error("ReplaceFavorites() ingin error")
		}
	})

	t.Run("gagal mencari user", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolveErr: apperrors.ErrNotFound})
		if err := svc.ReplaceFavorites(context.Background(), "x", []int64{1}); !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("ReplaceFavorites() = %v", err)
		}
	})

	t.Run("gagal menyimpan", func(t *testing.T) {
		svc := NewService(&fakeRepository{resolvedID: 1, validCount: 1, replaceErr: errors.New("db down")})
		if err := svc.ReplaceFavorites(context.Background(), "x", []int64{1}); err == nil {
			t.Error("ReplaceFavorites() ingin error")
		}
	})
}
