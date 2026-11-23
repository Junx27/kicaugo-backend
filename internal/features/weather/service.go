package weather

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/entities"
)

// Fetcher adalah kontrak sumber data cuaca.
type Fetcher interface {
	Fetch(ctx context.Context, lat, lon float64, days int) (entities.Weather, error)
}

// Service menyediakan data cuaca dengan cache in-memory.
type Service interface {
	GetForecast(ctx context.Context, lat, lon float64, days int) (entities.Weather, error)
}

type cacheEntry struct {
	value entities.Weather
	exp   time.Time
}

type service struct {
	fetcher Fetcher
	ttl     time.Duration
	mu      sync.Mutex
	cache   map[string]cacheEntry
}

// NewService membuat service dengan fetcher yang diberikan.
func NewService(fetcher Fetcher) Service {
	return &service{fetcher: fetcher, ttl: 30 * time.Minute, cache: map[string]cacheEntry{}}
}

// GetForecast mengambil cuaca dari cache bila masih valid, atau dari fetcher.
func (s *service) GetForecast(ctx context.Context, lat, lon float64, days int) (entities.Weather, error) {
	if days < 1 {
		days = 1
	}
	if days > 10 {
		days = 10
	}
	key := fmt.Sprintf("%.2f,%.2f:%d", lat, lon, days)

	s.mu.Lock()
	if e, ok := s.cache[key]; ok && time.Now().Before(e.exp) {
		s.mu.Unlock()
		return e.value, nil
	}
	s.mu.Unlock()

	w, err := s.fetcher.Fetch(ctx, lat, lon, days)
	if err != nil {
		return entities.Weather{}, err
	}

	s.mu.Lock()
	s.cache[key] = cacheEntry{value: w, exp: time.Now().Add(s.ttl)}
	s.mu.Unlock()
	return w, nil
}
