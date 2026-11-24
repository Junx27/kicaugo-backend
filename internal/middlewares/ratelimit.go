package middlewares

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"

	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

const (
	defaultRatePerMin    = 300
	rateWindow           = time.Minute
	retryAfterHeader     = "Retry-After"
	defaultMaxBuckets    = 10000
	defaultBucketIdleTTL = 10 * time.Minute
)

type rateBucket struct {
	tokens float64
	last   time.Time
}

// RateLimiter membatasi request per identifier dengan token bucket.
type RateLimiter struct {
	capacity     float64
	refillPerSec float64
	mu           sync.Mutex
	buckets      map[string]*rateBucket
	now          func() time.Time
	maxBuckets   int
	idleTTL      time.Duration
}

// NewRateLimiter membuat limiter dengan kapasitas per menit tertentu.
func NewRateLimiter(perMin int) *RateLimiter {
	if perMin <= 0 {
		perMin = defaultRatePerMin
	}
	return &RateLimiter{
		capacity:     float64(perMin),
		refillPerSec: float64(perMin) / rateWindow.Seconds(),
		buckets:      map[string]*rateBucket{},
		now:          time.Now,
		maxBuckets:   defaultMaxBuckets,
		idleTTL:      defaultBucketIdleTTL,
	}
}

// Middleware mengembalikan middleware Huma yang memasang limit pada route terkait.
func (r *RateLimiter) Middleware(api apphttp.API) apphttp.Middleware {
	return func(ctx huma.Context, next func(huma.Context)) {
		allowed, wait := r.allow(identifierFromRequest(ctx.Context(), ctx.RemoteAddr()))
		if !allowed {
			ctx.SetHeader(retryAfterHeader, strconv.Itoa(retryAfterSeconds(wait)))
			apphttp.WriteError(api, ctx, apphttp.Error(http.StatusTooManyRequests, "Too many requests"))
			return
		}
		next(ctx)
	}
}

func (r *RateLimiter) allow(identifier string) (bool, time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	bucket, ok := r.buckets[identifier]
	if !ok {
		bucket = &rateBucket{tokens: r.capacity, last: now}
		r.buckets[identifier] = bucket
	}
	bucket.tokens = math.Min(r.capacity, bucket.tokens+now.Sub(bucket.last).Seconds()*r.refillPerSec)
	bucket.last = now

	if len(r.buckets) > r.maxBuckets {
		r.pruneLocked(now)
	}
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true, 0
	}
	wait := time.Duration((1 - bucket.tokens) / r.refillPerSec * float64(time.Second))
	return false, wait
}

func (r *RateLimiter) pruneLocked(now time.Time) {
	for id, bucket := range r.buckets {
		if now.Sub(bucket.last) > r.idleTTL {
			delete(r.buckets, id)
		}
	}
	for id := range r.buckets {
		if len(r.buckets) <= r.maxBuckets {
			break
		}
		delete(r.buckets, id)
	}
}

func identifierFromRequest(ctx context.Context, remoteAddr string) string {
	if claims, ok := UserFromContext[jwt.Claims](ctx); ok && claims.Subject != "" {
		return claims.Subject
	}
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	if remoteAddr != "" {
		return remoteAddr
	}
	return "anonymous"
}

func retryAfterSeconds(wait time.Duration) int {
	seconds := int(math.Ceil(wait.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
