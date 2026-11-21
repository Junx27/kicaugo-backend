package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

const (
	userKey      = "user"
	bearerPrefix = "Bearer "
)

// AuthMiddleware creates a Huma middleware for JWT authentication
func AuthMiddleware(tokens *TokenManager) func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		raw, ok := strings.CutPrefix(authHeader, bearerPrefix)
		if !ok || raw == "" {
			ctx.SetStatus(http.StatusUnauthorized)
			ctx.BodyWriter().Write([]byte(`{"status":401,"message":"Authentication required","data":null}`))
			return
		}
		user, err := tokens.Parse(raw)
		if err != nil {
			ctx.SetStatus(http.StatusUnauthorized)
			ctx.BodyWriter().Write([]byte(`{"status":401,"message":"Invalid token","data":null}`))
			return
		}
		ctx = huma.WithValue(ctx, userKey, user)
		next(ctx)
	}
}

// UserFromContext mengambil identitas pengguna yang sudah diverifikasi dari Huma context
func UserFromContext(ctx context.Context) (User, bool) {
	val := ctx.Value(userKey)
	if val == nil {
		return User{}, false
	}
	user, ok := val.(User)
	return user, ok
}
