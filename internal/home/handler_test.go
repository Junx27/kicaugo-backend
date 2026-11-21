package home

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/auth"
	"github.com/danielgtaylor/huma/v2/humatest"
)

// fakeRepository adalah implementasi Repository untuk testing.
type fakeRepository struct {
	content Content
	err     error
}

func (f *fakeRepository) Get(ctx context.Context) (Content, error) {
	return f.content, f.err
}

func sampleContent() Content {
	return Content{
		Summary:   Summary{Location: "Jakarta Selatan", LiveCount: 14, ActiveTicketCount: 2},
		Schedules: []ScheduleItem{{Title: "Latpres Cibubur", Status: "available", Category: "kacer"}},
	}
}

func TestGetHome(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		_, api := humatest.New(t)
		repo := &fakeRepository{content: sampleContent()}
		service := NewService(repo)
		tokens := auth.NewTokenManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, service, tokens)

		token, _ := tokens.Issue(auth.User{Name: "Kicau Mania", Identifier: "test@gmail.com"})
		resp := api.Get("/api/v1/home", "Authorization: Bearer "+token)
		if resp.Code != http.StatusOK {
			t.Fatalf("status = %d, ingin 200", resp.Code)
		}
		var result struct {
			Status  int     `json:"status"`
			Message string  `json:"message"`
			Data    Content `json:"data"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Status != http.StatusOK {
			t.Errorf("status = %d, ingin %d", result.Status, http.StatusOK)
		}
		if result.Data.Summary.Location != "Jakarta Selatan" {
			t.Errorf("content = %+v", result.Data)
		}
	})

	t.Run("tanpa token", func(t *testing.T) {
		_, api := humatest.New(t)
		repo := &fakeRepository{content: sampleContent()}
		service := NewService(repo)
		tokens := auth.NewTokenManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, service, tokens)

		resp := api.Get("/api/v1/home")
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, ingin 401", resp.Code)
		}
	})

	t.Run("gagal memuat", func(t *testing.T) {
		_, api := humatest.New(t)
		repo := &fakeRepository{err: errors.New("db down")}
		service := NewService(repo)
		tokens := auth.NewTokenManager([]byte("rahasia-untuk-test"), 24*3600*1000000000)
		RegisterRoutes(api, service, tokens)

		token, _ := tokens.Issue(auth.User{Name: "Kicau Mania", Identifier: "test@gmail.com"})
		resp := api.Get("/api/v1/home", "Authorization: Bearer "+token)
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, ingin 500", resp.Code)
		}
	})
}
