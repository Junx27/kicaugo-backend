package basemap

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// Fetcher adalah kontrak sumber tile basemap.
type Fetcher interface {
	Fetch(ctx context.Context, z, x, y int, retina bool, style string) ([]byte, error)
}

// Service menyediakan tile basemap dengan cache LRU dan singleflight.
type Service interface {
	Get(ctx context.Context, z, x, y int, retina bool, style string) ([]byte, error)
	CacheTTL() time.Duration
}

type cacheEntry struct {
	key     string
	data    []byte
	expires time.Time
}

type service struct {
	fetcher    Fetcher
	ttl        time.Duration
	maxEntries int
	mu         sync.Mutex
	order      *list.List
	items      map[string]*list.Element
	group      singleflight.Group
	now        func() time.Time
}

// NewService membuat service dengan batas cache dan TTL yang diberikan.
func NewService(fetcher Fetcher, ttl time.Duration, maxEntries int) Service {
	if maxEntries < 1 {
		maxEntries = 1
	}
	return &service{
		fetcher:    fetcher,
		ttl:        ttl,
		maxEntries: maxEntries,
		order:      list.New(),
		items:      map[string]*list.Element{},
		now:        time.Now,
	}
}

// Get mengambil tile dari cache, atau dari fetcher bila belum ada/kedaluwarsa.
func (s *service) Get(ctx context.Context, z, x, y int, retina bool, style string) ([]byte, error) {
	key := cacheKey(z, x, y, retina, style)
	if data, ok := s.get(key); ok {
		return data, nil
	}

	result, err, _ := s.group.Do(key, func() (any, error) {
		data, loadErr := s.load(ctx, key, z, x, y, retina, style)
		if loadErr != nil {
			return nil, loadErr
		}
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]byte), nil
}

func (s *service) load(ctx context.Context, key string, z, x, y int, retina bool, style string) ([]byte, error) {
	if data, ok := s.get(key); ok {
		return data, nil
	}
	data, err := s.fetcher.Fetch(ctx, z, x, y, retina, style)
	if err != nil {
		return nil, err
	}
	s.put(key, data)
	return data, nil
}

func (s *service) get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	element, ok := s.items[key]
	if !ok {
		return nil, false
	}
	entry := element.Value.(*cacheEntry)
	if !s.now().Before(entry.expires) {
		s.removeLocked(element)
		return nil, false
	}
	s.order.MoveToFront(element)
	return entry.data, true
}

func (s *service) put(key string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &cacheEntry{key: key, data: data, expires: s.now().Add(s.ttl)}
	if element, ok := s.items[key]; ok {
		element.Value = entry
		s.order.MoveToFront(element)
		return
	}
	s.items[key] = s.order.PushFront(entry)
	for s.order.Len() > s.maxEntries {
		s.removeLocked(s.order.Back())
	}
}

func (s *service) removeLocked(element *list.Element) {
	s.order.Remove(element)
	delete(s.items, element.Value.(*cacheEntry).key)
}

// CacheTTL mengembalikan TTL cache untuk header Cache-Control.
func (s *service) CacheTTL() time.Duration {
	return s.ttl
}

func cacheKey(z, x, y int, retina bool, style string) string {
	return fmt.Sprintf("%d/%d/%d/%t/%s", z, x, y, retina, style)
}
