package profile

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Junx27/kicaugo-backend/internal/entities"
	dbutil "github.com/Junx27/kicaugo-backend/internal/utils/db"
	apperrors "github.com/Junx27/kicaugo-backend/internal/utils/errors"
)

// Repository menyimpan dan membaca profil pengguna.
type Repository interface {
	ResolveUserID(ctx context.Context, identifier string) (int64, error)
	EnsureAndGet(ctx context.Context, userID int64) (entities.UserProfile, error)
	FavoriteCategories(ctx context.Context, userID int64) ([]entities.BirdCategory, error)
	City(ctx context.Context, cityID int64) (entities.City, error)
	CountValidCategories(ctx context.Context, categoryIDs []int64) (int, error)
	Update(ctx context.Context, userID int64, in entities.ProfileUpdate) error
	UpdateLocation(ctx context.Context, userID int64, latitude, longitude *float64, accuracyM *int) error
	ReplaceFavorites(ctx context.Context, userID int64, categoryIDs []int64) error
}

// repository adalah implementasi Repository berbasis pgx.
type repository struct {
	db dbutil.Querier
}

// NewRepository membuat repository dengan connection pool yang sudah ada.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

// ResolveUserID menerjemahkan identifier dari token menjadi ID user.
func (r *repository) ResolveUserID(ctx context.Context, identifier string) (int64, error) {
	return dbutil.ResolveUserID(ctx, r.db, identifier)
}

// EnsureAndGet membuat profil bila belum ada, lalu mengembalikannya beserta kota dan provinsi.
func (r *repository) EnsureAndGet(ctx context.Context, userID int64) (entities.UserProfile, error) {
	const ensure = `INSERT INTO user_profiles (user_id, full_name)
		SELECT $1, name FROM users WHERE id = $1
		ON CONFLICT (user_id) DO NOTHING`
	if _, err := r.db.Exec(ctx, ensure, userID); err != nil {
		return entities.UserProfile{}, fmt.Errorf("ensure profile: %w", err)
	}

	const query = `SELECT p.id, p.user_id, p.full_name, p.phone, p.avatar_url, p.bio, p.city_id,
		p.current_latitude, p.current_longitude, p.current_location_accuracy_m,
		p.current_location_updated_at, p.search_radius_km, p.created_at, p.updated_at,
		COALESCE(c.id, 0), COALESCE(c.name, ''), COALESCE(c.type, ''), COALESCE(c.slug, ''),
		COALESCE(pr.id, 0), COALESCE(pr.code, ''), COALESCE(pr.name, ''), (c.id IS NOT NULL)
		FROM user_profiles p
		LEFT JOIN cities c ON c.id = p.city_id
		LEFT JOIN provinces pr ON pr.id = c.province_id
		WHERE p.user_id = $1`

	var p entities.UserProfile
	var cityID, provinceID int64
	var cityName, cityType, citySlug, provinceCode, provinceName string
	var hasCity bool
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.FullName, &p.Phone, &p.AvatarURL, &p.Bio, &p.CityID,
		&p.CurrentLatitude, &p.CurrentLongitude, &p.CurrentLocationAccuracyM,
		&p.CurrentLocationUpdatedAt, &p.SearchRadiusKM, &p.CreatedAt, &p.UpdatedAt,
		&cityID, &cityName, &cityType, &citySlug, &provinceID, &provinceCode, &provinceName, &hasCity,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.UserProfile{}, apperrors.ErrNotFound
	}
	if err != nil {
		return entities.UserProfile{}, fmt.Errorf("get profile: %w", err)
	}
	if hasCity {
		p.City = &entities.City{ID: cityID, Name: cityName, Type: cityType, Slug: citySlug}
		p.Province = &entities.Province{ID: provinceID, Code: provinceCode, Name: provinceName}
	}
	return p, nil
}

// FavoriteCategories mengambil daftar kategori favorit user.
func (r *repository) FavoriteCategories(ctx context.Context, userID int64) ([]entities.BirdCategory, error) {
	const query = `SELECT c.id, c.name, c.slug, c.icon, c.created_at
		FROM user_favorite_categories f
		JOIN bird_categories c ON c.id = f.category_id
		WHERE f.user_id = $1 ORDER BY c.name`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite categories: %w", err)
	}
	defer rows.Close()

	return dbutil.ScanBirdCategories(rows)
}

// City mencari kabupaten/kota berdasarkan ID.
func (r *repository) City(ctx context.Context, cityID int64) (entities.City, error) {
	const query = `SELECT id, province_id, code, name, type, slug, latitude, longitude, created_at FROM cities WHERE id = $1`
	var c entities.City
	err := r.db.QueryRow(ctx, query, cityID).Scan(
		&c.ID, &c.ProvinceID, &c.Code, &c.Name, &c.Type, &c.Slug, &c.Latitude, &c.Longitude, &c.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return entities.City{}, apperrors.ErrInvalidCity
	}
	if err != nil {
		return entities.City{}, fmt.Errorf("get city: %w", err)
	}
	return c, nil
}

// CountValidCategories menghitung jumlah ID kategori yang benar-benar ada.
func (r *repository) CountValidCategories(ctx context.Context, categoryIDs []int64) (int, error) {
	const query = `SELECT count(DISTINCT id) FROM bird_categories WHERE id = ANY($1::bigint[])`
	var count int
	if err := r.db.QueryRow(ctx, query, categoryIDs).Scan(&count); err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return count, nil
}

// Update menerapkan perubahan parsial pada profil.
func (r *repository) Update(ctx context.Context, userID int64, in entities.ProfileUpdate) error {
	sets, args := buildUpdate(userID, in)
	if len(sets) == 0 {
		return nil
	}
	query := "UPDATE user_profiles SET " + strings.Join(sets, ", ") + " WHERE user_id = $1"
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

// UpdateLocation menyimpan atau menghapus lokasi terkini. Timestamp diisi server.
func (r *repository) UpdateLocation(ctx context.Context, userID int64, latitude, longitude *float64, accuracyM *int) error {
	if latitude == nil {
		const clear = `UPDATE user_profiles SET current_latitude = NULL, current_longitude = NULL,
			current_location_accuracy_m = NULL, current_location_updated_at = NULL WHERE user_id = $1`
		if _, err := r.db.Exec(ctx, clear, userID); err != nil {
			return fmt.Errorf("clear location: %w", err)
		}
		return nil
	}
	const set = `UPDATE user_profiles SET current_latitude = $2, current_longitude = $3,
		current_location_accuracy_m = $4, current_location_updated_at = now() WHERE user_id = $1`
	if _, err := r.db.Exec(ctx, set, userID, *latitude, *longitude, accuracyM); err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	return nil
}

// ReplaceFavorites mengganti seluruh kategori favorit dalam satu transaksi.
func (r *repository) ReplaceFavorites(ctx context.Context, userID int64, categoryIDs []int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin favorites: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `DELETE FROM user_favorite_categories WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete favorites: %w", err)
	}
	const insert = `INSERT INTO user_favorite_categories (user_id, category_id)
		SELECT $1, unnest($2::bigint[])`
	if _, err := tx.Exec(ctx, insert, userID, categoryIDs); err != nil {
		return fmt.Errorf("insert favorites: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit favorites: %w", err)
	}
	return nil
}

// buildUpdate menyusun klausa SET dan argumen terparameterisasi.
func buildUpdate(userID int64, in entities.ProfileUpdate) ([]string, []any) {
	args := []any{userID}
	sets := make([]string, 0, 6)
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if in.FullName != nil {
		add("full_name", *in.FullName)
	}
	if in.Phone != nil {
		add("phone", *in.Phone)
	}
	if in.AvatarURL != nil {
		add("avatar_url", *in.AvatarURL)
	}
	if in.Bio != nil {
		add("bio", *in.Bio)
	}
	if in.CityID != nil {
		add("city_id", *in.CityID)
	}
	if in.SearchRadiusKM != nil {
		add("search_radius_km", *in.SearchRadiusKM)
	}
	return sets, args
}
