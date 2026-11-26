package event

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

func TestRepositoryCountGantanganScanErrorCoverage(t *testing.T) {
	repo := &repository{db: fakeQuerier{rowErr: errors.New("scan err")}}
	if _, err := repo.CountGantangan(context.Background()); err == nil {
		t.Fatal("ingin error")
	}
}

func TestRepositoryListNearbySchedulesQueryErrorCoverage(t *testing.T) {
	repo := &repository{db: fakeQuerier{queryErr: errors.New("query err")}}
	origin := entities.NearbyOrigin{Latitude: 1, Longitude: 1}
	if _, err := repo.ListNearbySchedules(context.Background(), origin, 50, 0, 10); err == nil {
		t.Fatal("ingin error")
	}
}

func TestRepositoryListNearbySchedulesScanErrorCoverage(t *testing.T) {
	rows := &fakeRows{nextMax: 1, scanErr: errors.New("scan err")}
	repo := &repository{db: fakeQuerier{rows: rows}}
	origin := entities.NearbyOrigin{Latitude: 1, Longitude: 1}
	if _, err := repo.ListNearbySchedules(context.Background(), origin, 50, 0, 10); err == nil {
		t.Fatal("ingin error")
	}
}

func TestRepositoryListNearbySchedulesRowsErrCoverage(t *testing.T) {
	rows := &fakeRows{nextMax: 0, err: errors.New("rows err")}
	repo := &repository{db: fakeQuerier{rows: rows}}
	origin := entities.NearbyOrigin{Latitude: 1, Longitude: 1}
	if _, err := repo.ListNearbySchedules(context.Background(), origin, 50, 0, 10); err == nil {
		t.Fatal("ingin error")
	}
}
