package profile

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
	"github.com/Junx27/kicaugo-backend/internal/utils/response"
)

type fakeService struct {
	view      ProfileView
	getErr    error
	updateErr error
	locErr    error
	favErr    error

	getIdentifier string
	updatedIn     entities.ProfileUpdate
	updatedProv   *int64
	location      struct {
		latitude  *float64
		longitude *float64
		accuracy  *int
	}
	replacedIDs []int64
}

func (f *fakeService) Get(ctx context.Context, identifier string) (ProfileView, error) {
	f.getIdentifier = identifier
	return f.view, f.getErr
}

func (f *fakeService) Update(ctx context.Context, identifier string, in entities.ProfileUpdate, provinceID *int64) error {
	f.updatedIn = in
	f.updatedProv = provinceID
	return f.updateErr
}

func (f *fakeService) UpdateLocation(ctx context.Context, identifier string, latitude, longitude *float64, accuracyM *int) error {
	f.location.latitude = latitude
	f.location.longitude = longitude
	f.location.accuracy = accuracyM
	return f.locErr
}

func (f *fakeService) ReplaceFavorites(ctx context.Context, identifier string, ids []int64) error {
	f.replacedIDs = ids
	return f.favErr
}

func newTestTokens() *jwt.Manager {
	return jwt.NewManager([]byte("rahasia-untuk-test"), time.Hour)
}

func bearerHeader(t *testing.T, tokens *jwt.Manager) string {
	return bearerHeaderForSubject(t, tokens, "user@example.com")
}

func bearerHeaderForSubject(t *testing.T, tokens *jwt.Manager, subject string) string {
	t.Helper()
	token, err := tokens.Issue(subject, "Tester")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return "Authorization: Bearer " + token
}

func setupHandler(t *testing.T, svc Service) (humatest.TestAPI, *jwt.Manager) {
	t.Helper()
	_, api := humatest.New(t)
	tokens := newTestTokens()
	RegisterRoutes(api, svc, tokens)
	return api, tokens
}

func TestGetProfileHandler(t *testing.T) {
	t.Run("berhasil menampilkan profil lengkap", func(t *testing.T) {
		cityID, provinceID := int64(5), int64(11)
		lat, lng := -6.2, 106.8
		acc := 25
		updatedAt := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
		phone := "+6281234567890"
		city := &entities.City{ID: cityID, Name: "Bandung", Type: "kota", Slug: "kota-bandung"}
		province := &entities.Province{ID: provinceID, Code: "32", Name: "Jawa Barat"}
		svc := &fakeService{view: ProfileView{
			Profile: entities.UserProfile{
				FullName: "Budi", Phone: &phone, CityID: &cityID,
				CurrentLatitude: &lat, CurrentLongitude: &lng,
				CurrentLocationAccuracyM: &acc, CurrentLocationUpdatedAt: &updatedAt,
				SearchRadiusKM: 50, City: city, Province: province,
			},
			Favorites:     []entities.BirdCategory{{ID: 2, Name: "Murai Batu"}},
			IsComplete:    true,
			MissingFields: []string{},
		}}
		api, tokens := setupHandler(t, svc)

		resp := api.Get("/api/v1/me/profile", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		var body struct {
			Status int `json:"status"`
			Data   struct {
				FullName string `json:"full_name"`
				City     struct {
					Name     string `json:"name"`
					Province struct {
						Name string `json:"name"`
					} `json:"province"`
				} `json:"city"`
				CurrentLocation struct {
					Latitude  float64 `json:"latitude"`
					UpdatedAt string  `json:"updated_at"`
				} `json:"current_location"`
				FavoriteCategories []response.FavoriteCategoryData `json:"favorite_categories"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Status != http.StatusOK || body.Data.FullName != "Budi" ||
			body.Data.City.Name != "Bandung" || body.Data.City.Province.Name != "Jawa Barat" ||
			body.Data.CurrentLocation.Latitude != -6.2 || len(body.Data.FavoriteCategories) != 1 {
			t.Errorf("body = %+v", body)
		}
		if svc.getIdentifier != "user@example.com" {
			t.Errorf("identifier = %q", svc.getIdentifier)
		}
	})

	t.Run("profil kosong tanpa kota dan lokasi", func(t *testing.T) {
		svc := &fakeService{view: buildView(entities.UserProfile{}, nil)}
		api, tokens := setupHandler(t, svc)

		resp := api.Get("/api/v1/me/profile", bearerHeader(t, tokens))
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
		var body struct {
			Data struct {
				City            *struct{} `json:"city"`
				CurrentLocation *struct{} `json:"current_location"`
				IsComplete      bool      `json:"is_complete"`
			} `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
			t.Fatalf("body bukan JSON: %v", err)
		}
		if body.Data.City != nil || body.Data.CurrentLocation != nil || body.Data.IsComplete {
			t.Errorf("data = %+v", body.Data)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Get("/api/v1/me/profile")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{getErr: errors.New("db down")})
		resp := api.Get("/api/v1/me/profile", bearerHeader(t, tokens))
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})
}

func TestUpdateProfileHandler(t *testing.T) {
	body := map[string]any{
		"full_name":        "Budi",
		"phone":            "081234567890",
		"city_id":          5,
		"province_id":      11,
		"search_radius_km": 25,
	}

	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{}
		api, tokens := setupHandler(t, svc)
		resp := api.Patch("/api/v1/me/profile", bearerHeader(t, tokens), body)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		if svc.updatedIn.FullName == nil || *svc.updatedIn.FullName != "Budi" ||
			svc.updatedProv == nil || *svc.updatedProv != 11 || svc.updatedIn.SearchRadiusKM == nil {
			t.Errorf("updatedIn = %+v, prov = %v", svc.updatedIn, svc.updatedProv)
		}
	})

	t.Run("input tidak valid ditolak Huma", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{})
		resp := api.Patch("/api/v1/me/profile", bearerHeader(t, tokens), map[string]any{"search_radius_km": 999})
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("kota tidak valid", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{updateErr: apperrors.ErrInvalidCity})
		resp := api.Patch("/api/v1/me/profile", bearerHeader(t, tokens), body)
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Patch("/api/v1/me/profile", body)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestUpdateCurrentLocationHandler(t *testing.T) {
	body := map[string]any{"latitude": -6.2, "longitude": 106.8, "accuracy_m": 15}

	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{}
		api, tokens := setupHandler(t, svc)
		resp := api.Put("/api/v1/me/profile/current-location", bearerHeader(t, tokens), body)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		if svc.location.latitude == nil || *svc.location.latitude != -6.2 || svc.location.accuracy == nil {
			t.Errorf("location = %+v", svc.location)
		}
	})

	t.Run("koordinat tidak lengkap", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{locErr: apperrors.ErrInvalidCoordinates})
		resp := api.Put("/api/v1/me/profile/current-location", bearerHeader(t, tokens), map[string]any{"latitude": -6.2})
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Put("/api/v1/me/profile/current-location", body)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

func TestReplaceFavoritesHandler(t *testing.T) {
	body := map[string]any{"category_ids": []int64{1, 2}}

	t.Run("berhasil", func(t *testing.T) {
		svc := &fakeService{}
		api, tokens := setupHandler(t, svc)
		resp := api.Put("/api/v1/me/profile/favorite-categories", bearerHeader(t, tokens), body)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200; body = %s", resp.Code, resp.Body.String())
		}
		if len(svc.replacedIDs) != 2 {
			t.Errorf("replacedIDs = %v", svc.replacedIDs)
		}
	})

	t.Run("kategori tidak valid", func(t *testing.T) {
		api, tokens := setupHandler(t, &fakeService{favErr: apperrors.ErrInvalidCategory})
		resp := api.Put("/api/v1/me/profile/favorite-categories", bearerHeader(t, tokens), body)
		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("status = %d, ingin 422", resp.Code)
		}
	})

	t.Run("ditolak tanpa token", func(t *testing.T) {
		api, _ := setupHandler(t, &fakeService{})
		resp := api.Put("/api/v1/me/profile/favorite-categories", body)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})
}

// TestEndpointDenganSubjectKosong menutup cabang pertahanan pada keempat handler
// saat token valid tetapi tidak membawa identifier pengguna.
func TestEndpointDenganSubjectKosong(t *testing.T) {
	api, tokens := setupHandler(t, &fakeService{})
	header := bearerHeaderForSubject(t, tokens, "")
	requests := []struct {
		name string
		call func() int
	}{
		{name: "get", call: func() int { return api.Get("/api/v1/me/profile", header).Code }},
		{name: "patch", call: func() int {
			return api.Patch("/api/v1/me/profile", header, map[string]any{"bio": "x"}).Code
		}},
		{name: "location", call: func() int {
			return api.Put("/api/v1/me/profile/current-location", header, map[string]any{}).Code
		}},
		{name: "favorites", call: func() int {
			return api.Put("/api/v1/me/profile/favorite-categories", header, map[string]any{"category_ids": []int64{1}}).Code
		}},
	}
	for _, tc := range requests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.call(); got != http.StatusUnauthorized {
				t.Errorf("status = %d, ingin 401", got)
			}
		})
	}
}

func TestMapServiceError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{err: apperrors.ErrInvalidPayload, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrInvalidCity, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrProvinceMismatch, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrInvalidCoordinates, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrInvalidCategory, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrInvalidRadius, want: http.StatusUnprocessableEntity},
		{err: apperrors.ErrNotFound, want: http.StatusNotFound},
		{err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tc := range cases {
		err := mapServiceError(tc.err)
		statusErr, ok := err.(interface{ GetStatus() int })
		if !ok || statusErr.GetStatus() != tc.want {
			t.Errorf("mapServiceError(%v) = %v, ingin status %d", tc.err, err, tc.want)
		}
	}
}
