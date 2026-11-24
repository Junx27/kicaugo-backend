package middlewares

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	jwtlib "github.com/golang-jwt/jwt/v5"

	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

const (
	testSubject    = "user@example.com"
	testIDAlpha    = "a"
	testIP         = "1.2.3.4"
	testRemoteAddr = "1.2.3.4:5678"
	testTileRoute  = "/tiles"
)

func TestNewRateLimiter(t *testing.T) {
	limited := NewRateLimiter(60)
	if limited.capacity != 60 || limited.refillPerSec != 1 {
		t.Errorf("limiter = %+v", limited)
	}
	defaulted := NewRateLimiter(0)
	if defaulted.capacity != defaultRatePerMin {
		t.Errorf("capacity = %v, ingin %d", defaulted.capacity, defaultRatePerMin)
	}
}

func TestRateLimiterAllow(t *testing.T) {
	limiter := NewRateLimiter(1)
	now := time.Unix(0, 0)
	limiter.now = func() time.Time { return now }

	if allowed, _ := limiter.allow(testIDAlpha); !allowed {
		t.Fatal("request pertama harus diizinkan")
	}
	allowed, wait := limiter.allow(testIDAlpha)
	if allowed {
		t.Fatal("request kedua harus ditolak")
	}
	if wait <= 0 {
		t.Errorf("wait = %v, ingin > 0", wait)
	}

	now = now.Add(time.Minute)
	if allowed, _ := limiter.allow(testIDAlpha); !allowed {
		t.Error("request setelah refill harus diizinkan")
	}

	other := NewRateLimiter(1)
	if allowed, _ := other.allow(testIDAlpha); !allowed {
		t.Fatal("identifier a harus diizinkan")
	}
	if allowed, _ := other.allow("b"); !allowed {
		t.Fatal("identifier b harus punya bucket sendiri")
	}
}

func TestRateLimiterIdentifier(t *testing.T) {
	claimsCtx := WithUser(context.Background(), jwt.Claims{
		RegisteredClaims: jwtlib.RegisteredClaims{Subject: testSubject},
	})
	if got := identifierFromRequest(claimsCtx, testRemoteAddr); got != testSubject {
		t.Errorf("identifier = %q, ingin %q", got, testSubject)
	}

	emptySubject := WithUser(context.Background(), jwt.Claims{})
	if got := identifierFromRequest(emptySubject, testRemoteAddr); got != testIP {
		t.Errorf("identifier = %q, ingin 1.2.3.4", got)
	}
	if got := identifierFromRequest(context.Background(), testIP); got != testIP {
		t.Errorf("identifier = %q, ingin 1.2.3.4", got)
	}
	if got := identifierFromRequest(context.Background(), ""); got != "anonymous" {
		t.Errorf("identifier = %q, ingin anonymous", got)
	}
}

func TestRetryAfterSeconds(t *testing.T) {
	if got := retryAfterSeconds(0); got != 1 {
		t.Errorf("retryAfterSeconds(0) = %d, ingin 1", got)
	}
	if got := retryAfterSeconds(2 * time.Second); got != 2 {
		t.Errorf("retryAfterSeconds(2s) = %d, ingin 2", got)
	}
}

func TestRateLimiterPrune(t *testing.T) {
	limiter := NewRateLimiter(60)
	limiter.maxBuckets = 2
	limiter.idleTTL = time.Minute
	now := time.Unix(0, 0)
	limiter.now = func() time.Time { return now }

	limiter.allow(testIDAlpha)
	limiter.allow("b")
	limiter.allow("c")
	if len(limiter.buckets) > limiter.maxBuckets {
		t.Errorf("jumlah bucket = %d, ingin <= %d", len(limiter.buckets), limiter.maxBuckets)
	}

	now = now.Add(2 * time.Minute)
	limiter.allow("d")
	if len(limiter.buckets) > limiter.maxBuckets {
		t.Errorf("jumlah bucket setelah prune = %d", len(limiter.buckets))
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	limiter := NewRateLimiter(1)
	_, api := humatest.New(t)
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		next(huma.WithValue(ctx, userKey, jwt.Claims{
			RegisteredClaims: jwtlib.RegisteredClaims{Subject: testSubject},
		}))
	})
	api.UseMiddleware(limiter.Middleware(api))
	huma.Get(api, testTileRoute, func(context.Context, *struct{}) (*struct{}, error) {
		return nil, nil
	})

	if resp := api.Get(testTileRoute); resp.Code != http.StatusNoContent {
		t.Fatalf("request pertama status = %d", resp.Code)
	}
	resp := api.Get(testTileRoute)
	if resp.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, ingin 429", resp.Code)
	}
	if resp.Header().Get(retryAfterHeader) == "" {
		t.Error("Retry-After harus diisi")
	}
}
