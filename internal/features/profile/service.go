package profile

import (
	"context"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// ProfileView adalah profil lengkap beserta status kelengkapannya.
type ProfileView struct {
	Profile       entities.UserProfile
	Favorites     []entities.BirdCategory
	IsComplete    bool
	MissingFields []string
}

// Service berisi aturan bisnis profil pengguna.
type Service interface {
	Get(ctx context.Context, identifier string) (ProfileView, error)
	Update(ctx context.Context, identifier string, in entities.ProfileUpdate, provinceID *int64) error
	UpdateLocation(ctx context.Context, identifier string, latitude, longitude *float64, accuracyM *int) error
	ReplaceFavorites(ctx context.Context, identifier string, categoryIDs []int64) error
}

// service adalah implementasi Service.
type service struct {
	repo Repository
}

// NewService membuat service dengan repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Get mengambil profil milik identifier, membuat profil bila belum ada.
func (s *service) Get(ctx context.Context, identifier string) (ProfileView, error) {
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return ProfileView{}, err
	}
	p, err := s.repo.EnsureAndGet(ctx, userID)
	if err != nil {
		return ProfileView{}, err
	}
	favorites, err := s.repo.FavoriteCategories(ctx, userID)
	if err != nil {
		return ProfileView{}, err
	}
	return buildView(p, favorites), nil
}

// Update menerapkan perubahan parsial setelah memvalidasi kota dan provinsi.
func (s *service) Update(ctx context.Context, identifier string, in entities.ProfileUpdate, provinceID *int64) error {
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return err
	}
	if err := s.validateCity(ctx, userID, in.CityID, provinceID); err != nil {
		return err
	}
	if in.Phone != nil {
		phone := normalizePhone(*in.Phone)
		in.Phone = &phone
	}
	return s.repo.Update(ctx, userID, in)
}

// UpdateLocation menyimpan atau menghapus lokasi terkini milik user.
func (s *service) UpdateLocation(ctx context.Context, identifier string, latitude, longitude *float64, accuracyM *int) error {
	if err := validateLocation(latitude, longitude); err != nil {
		return err
	}
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return err
	}
	return s.repo.UpdateLocation(ctx, userID, latitude, longitude, accuracyM)
}

// ReplaceFavorites mengganti seluruh kategori favorit setelah memvalidasi semua ID.
func (s *service) ReplaceFavorites(ctx context.Context, identifier string, categoryIDs []int64) error {
	userID, err := s.repo.ResolveUserID(ctx, identifier)
	if err != nil {
		return err
	}
	ids := uniqueIDs(categoryIDs)
	count, err := s.repo.CountValidCategories(ctx, ids)
	if err == nil && count != len(ids) {
		err = apperrors.ErrInvalidCategory
	}
	if err != nil {
		return err
	}
	return s.repo.ReplaceFavorites(ctx, userID, ids)
}

// validateCity memastikan city_id valid dan province_id cocok dengan kota terpilih.
func (s *service) validateCity(ctx context.Context, userID int64, cityID, provinceID *int64) error {
	var err error
	if cityID != nil || provinceID != nil {
		var id int64
		id, err = s.targetCityID(ctx, userID, cityID)
		if err == nil && id != 0 {
			var city entities.City
			city, err = s.repo.City(ctx, id)
			if err == nil && provinceID != nil && city.ProvinceID != *provinceID {
				err = apperrors.ErrProvinceMismatch
			}
		}
	}
	return err
}

// targetCityID menentukan kota yang divalidasi: payload bila ada, selain itu kota profil saat ini.
func (s *service) targetCityID(ctx context.Context, userID int64, cityID *int64) (int64, error) {
	if cityID != nil {
		return *cityID, nil
	}
	p, err := s.repo.EnsureAndGet(ctx, userID)
	if err != nil {
		return 0, err
	}
	if p.CityID == nil {
		return 0, nil
	}
	return *p.CityID, nil
}
