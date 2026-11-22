// Package home menyediakan kontrak ringkas untuk halaman utama aplikasi.
package home

import (
	"context"

	"github.com/Junx27/kicaugo-backend/internal/features/event"
	"github.com/Junx27/kicaugo-backend/internal/features/profile"
)

// Options adalah parameter opsional permintaan home.
type Options struct {
	RadiusKM   int
	CategoryID int64
	Limit      int
}

// ProfileReader adalah bagian profile.Service yang dipakai home.
type ProfileReader interface {
	Get(ctx context.Context, identifier string) (profile.ProfileView, error)
}

// EventReader adalah bagian event.Service yang dipakai home.
type EventReader interface {
	ListNearby(ctx context.Context, identifier string, radiusKM int, categoryID int64, limit int) (event.NearbyResult, error)
}

// View menggabungkan profil dan event terdekat untuk halaman home.
type View struct {
	Profile profile.ProfileView
	Nearby  event.NearbyResult
}

// Service menyediakan data halaman home.
type Service interface {
	Get(ctx context.Context, identifier string, opts Options) (View, error)
}

// service adalah implementasi Service yang mengorkestrasi service existing.
type service struct {
	profiles ProfileReader
	events   EventReader
}

// NewService membuat service dengan service profil dan event yang sudah ada.
func NewService(profiles ProfileReader, events EventReader) Service {
	return &service{profiles: profiles, events: events}
}

// Get mengambil profil lalu event terdekat untuk identifier tertentu.
func (s *service) Get(ctx context.Context, identifier string, opts Options) (View, error) {
	view, err := s.profiles.Get(ctx, identifier)
	if err != nil {
		return View{}, err
	}
	nearby, err := s.events.ListNearby(ctx, identifier, opts.RadiusKM, opts.CategoryID, opts.Limit)
	if err != nil {
		return View{}, err
	}
	return View{Profile: view, Nearby: nearby}, nil
}
