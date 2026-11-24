package basemap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const testTileData = "tile"

type fakeFetcher struct {
	mu     sync.Mutex
	result []byte
	err    error
	calls  int
}

func (f *fakeFetcher) Fetch(context.Context, int, int, int, bool, string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.result, f.err
}

func (f *fakeFetcher) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

type blockingFetcher struct {
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
	once    sync.Once
	result  []byte
}

func (f *blockingFetcher) Fetch(context.Context, int, int, int, bool, string) ([]byte, error) {
	f.calls.Add(1)
	f.once.Do(func() { close(f.started) })
	<-f.release
	return f.result, nil
}

func TestServiceGet(t *testing.T) {
	t.Run("cache miss lalu hit", func(t *testing.T) {
		fetcher := &fakeFetcher{result: []byte(testTileData)}
		svc := NewService(fetcher, time.Hour, 10)

		first, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle)
		if err != nil || string(first) != testTileData {
			t.Fatalf("first = %q err=%v", first, err)
		}
		second, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle)
		if err != nil || string(second) != testTileData {
			t.Fatalf("second = %q err=%v", second, err)
		}
		if fetcher.callCount() != 1 {
			t.Errorf("calls = %d, ingin 1", fetcher.callCount())
		}
	})

	t.Run("error tidak di-cache", func(t *testing.T) {
		fetcher := &fakeFetcher{err: errors.New("boom")}
		svc := NewService(fetcher, time.Hour, 10)

		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err == nil {
			t.Fatal("Get() ingin error")
		}
		fetcher.err = nil
		fetcher.result = []byte("ok")
		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatalf("Get() kedua error = %v", err)
		}
		if fetcher.callCount() != 2 {
			t.Errorf("calls = %d, ingin 2", fetcher.callCount())
		}
	})

	t.Run("TTL kedaluwarsa", func(t *testing.T) {
		fetcher := &fakeFetcher{result: []byte(testTileData)}
		svc := NewService(fetcher, time.Hour, 10).(*service)
		now := time.Unix(0, 0)
		svc.now = func() time.Time { return now }

		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatal(err)
		}
		now = now.Add(time.Hour + time.Second)
		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatal(err)
		}
		if fetcher.callCount() != 2 {
			t.Errorf("calls = %d, ingin 2", fetcher.callCount())
		}
	})

	t.Run("batas entry dan LRU", func(t *testing.T) {
		fetcher := &fakeFetcher{result: []byte(testTileData)}
		svc := NewService(fetcher, time.Hour, 2).(*service)

		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Get(context.Background(), 1, 0, 0, false, defaultStyle); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.Get(context.Background(), 1, 0, 1, false, defaultStyle); err != nil {
			t.Fatal(err)
		}

		if len(svc.items) != 2 {
			t.Errorf("jumlah cache = %d, ingin 2", len(svc.items))
		}
		if _, ok := svc.items[cacheKey(0, 0, 0, false, defaultStyle)]; !ok {
			t.Error("entry yang baru diakses harus tetap ada")
		}
		if _, ok := svc.items[cacheKey(1, 0, 0, false, defaultStyle)]; ok {
			t.Error("entry LRU harus dievict")
		}
		if fetcher.callCount() != 3 {
			t.Errorf("calls = %d, ingin 3", fetcher.callCount())
		}
	})

	t.Run("update entry yang sama", func(t *testing.T) {
		fetcher := &fakeFetcher{result: []byte(testTileData)}
		svc := NewService(fetcher, time.Hour, 10).(*service)
		key := cacheKey(0, 0, 0, false, defaultStyle)
		svc.put(key, []byte("lama"))
		svc.put(key, []byte("baru"))

		data, ok := svc.get(key)
		if !ok || string(data) != "baru" {
			t.Errorf("data = %q ok=%v", data, ok)
		}
		if fetcher.callCount() != 0 {
			t.Errorf("calls = %d, ingin 0", fetcher.callCount())
		}
	})

	t.Run("singleflight", func(t *testing.T) {
		fetcher := &blockingFetcher{
			started: make(chan struct{}),
			release: make(chan struct{}),
			result:  []byte(testTileData),
		}
		svc := NewService(fetcher, time.Hour, 10)
		const workers = 8
		var wg sync.WaitGroup
		errCh := make(chan error, workers)

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				data, err := svc.Get(context.Background(), 0, 0, 0, false, defaultStyle)
				if err != nil {
					errCh <- err
					return
				}
				if string(data) != testTileData {
					errCh <- fmt.Errorf("data = %q", data)
				}
			}()
		}

		<-fetcher.started
		close(fetcher.release)
		wg.Wait()
		close(errCh)
		for err := range errCh {
			t.Error(err)
		}
		if calls := fetcher.calls.Load(); calls != 1 {
			t.Errorf("calls = %d, ingin 1", calls)
		}
	})

	t.Run("load memakai cache yang sudah terisi", func(t *testing.T) {
		fetcher := &fakeFetcher{result: []byte("upstream")}
		svc := NewService(fetcher, time.Hour, 10).(*service)
		key := cacheKey(0, 0, 0, false, defaultStyle)
		svc.put(key, []byte("cached"))

		data, err := svc.load(context.Background(), key, 0, 0, 0, false, defaultStyle)
		if err != nil || string(data) != "cached" {
			t.Errorf("data = %q err=%v", data, err)
		}
		if fetcher.callCount() != 0 {
			t.Errorf("calls = %d, ingin 0", fetcher.callCount())
		}
	})

	t.Run("TTL dan batas minimum", func(t *testing.T) {
		svc := NewService(&fakeFetcher{}, 2*time.Hour, 0).(*service)
		if svc.CacheTTL() != 2*time.Hour {
			t.Errorf("CacheTTL = %v", svc.CacheTTL())
		}
		if svc.maxEntries != 1 {
			t.Errorf("maxEntries = %d, ingin 1", svc.maxEntries)
		}
	})
}
