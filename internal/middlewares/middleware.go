// Package middlewares berisi middleware HTTP yang dapat dipakai lintas fitur.
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type contextKey string

const (
	userKey      contextKey = "user"
	bearerPrefix            = "Bearer "
)

// TokenParser memvalidasi token mentah dan mengembalikan identitas pengguna.
type TokenParser[T any] interface {
	Parse(raw string) (T, error)
}

// AuthMiddleware membuat middleware Huma untuk autentikasi JWT.
func AuthMiddleware[T any](tokens TokenParser[T]) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		raw, ok := strings.CutPrefix(authHeader, bearerPrefix)
		if !ok || raw == "" {
			writeUnauthorized(ctx, "Authentication required")
			return
		}
		user, err := tokens.Parse(raw)
		if err != nil {
			writeUnauthorized(ctx, "Invalid token")
			return
		}
		ctx = huma.WithValue(ctx, userKey, user)
		next(ctx)
	}
}

// UserFromContext mengambil identitas pengguna yang sudah diverifikasi dari Huma context.
func UserFromContext[T any](ctx context.Context) (T, bool) {
	val := ctx.Value(userKey)
	if val == nil {
		var zero T
		return zero, false
	}
	user, ok := val.(T)
	return user, ok
}

func writeUnauthorized(ctx huma.Context, message string) {
	ctx.SetStatus(http.StatusUnauthorized)
	_, _ = ctx.BodyWriter().Write([]byte(`{"status":401,"message":"` + message + `","data":null}`))
}
