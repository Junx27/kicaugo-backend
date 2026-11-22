// Package authctx menyediakan akses ke identitas pengguna dari context request.
package authctx

import (
	"context"

	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	apphttp "github.com/Junx27/kicaugo-backend/internal/utils/http"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

// IdentifierFromContext mengambil subject token yang sudah diverifikasi middleware.
func IdentifierFromContext(ctx context.Context) (string, error) {
	claims, ok := middlewares.UserFromContext[jwt.Claims](ctx)
	if !ok || claims.Subject == "" {
		return "", apphttp.Unauthorized("Authentication required")
	}
	return claims.Subject, nil
}
