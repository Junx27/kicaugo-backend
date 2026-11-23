package weather

import (
	"context"
	"errors"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

type fakeFetcher struct {
	result entities.Weather
	err    error
	calls  int
}

func (f *fakeFetcher) Fetch(ctx context.Context, lat, lon float64, days int) (entities.Weather, error) {
	f.calls++
	return f.result, f.err
}

func TestServiceGetForecast(t *testing.T) {
	t.Run("cache miss lalu hit", func(t *testing.T) {
		fetcher := &fakeFetcher{result: entities.Weather{Location: "Yogya"}}
		svc := NewService(fetcher)
		got1, err := svc.GetForecast(context.Background(), -7.7, 110.3, 1)
		if err != nil || got1.Location != "Yogya" {
			t.Fatalf("got1 = %+v err=%v", got1, err)
		}
		got2, err := svc.GetForecast(context.Background(), -7.7, 110.3, 1)
		if err != nil || got2.Location != "Yogya" {
			t.Fatalf("got2 = %+v err=%v", got2, err)
		}
		if fetcher.calls != 1 {
			t.Errorf("calls = %d, ingin 1 (cache harus hit)", fetcher.calls)
		}
	})

	t.Run("error tidak di-cache", func(t *testing.T) {
		fetcher := &fakeFetcher{err: errors.New("boom")}
		svc := NewService(fetcher)
		if _, err := svc.GetForecast(context.Background(), 1, 2, 1); err == nil {
			t.Error("ingin error")
		}
		fetcher.err = nil
		fetcher.result = entities.Weather{Location: "OK"}
		if _, err := svc.GetForecast(context.Background(), 1, 2, 1); err != nil {
			t.Fatalf("panggilan kedua error: %v", err)
		}
		if fetcher.calls != 2 {
			t.Errorf("calls = %d, ingin 2", fetcher.calls)
		}
	})

	t.Run("days di-clamp", func(t *testing.T) {
		fetcher := &fakeFetcher{result: entities.Weather{Location: "X"}}
		svc := NewService(fetcher)
		if _, err := svc.GetForecast(context.Background(), 1, 2, 0); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.GetForecast(context.Background(), 1, 2, 99); err != nil {
			t.Fatal(err)
		}
	})
}
