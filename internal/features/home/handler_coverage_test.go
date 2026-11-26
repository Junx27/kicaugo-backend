package home

import (
	"net/http"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/testutil"
)

func TestGetHomeSchedulesServiceErrorCoverage(t *testing.T) {
	svc := &fakeService{schedules: sampleSchedules(), scheduleErr: errDB()}
	api, tokens := setupHandler(t, svc)
	resp := api.Get("/api/v1/home/schedules", testutil.BearerHeader(t, tokens))
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, ingin 500", resp.Code)
	}
}

func errDB() error { return &testErr{} }

type testErr struct{}

func (testErr) Error() string { return "db down" }
