package authctx

import (
	"context"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/middlewares"
	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

func TestIdentifierFromContext(t *testing.T) {
	t.Run("tanpa claims", func(t *testing.T) {
		if _, err := IdentifierFromContext(context.Background()); err == nil {
			t.Error("IdentifierFromContext() ingin error")
		}
	})

	t.Run("subject kosong", func(t *testing.T) {
		ctx := middlewares.WithUser(context.Background(), jwt.Claims{})
		if _, err := IdentifierFromContext(ctx); err == nil {
			t.Error("IdentifierFromContext() ingin error")
		}
	})

	t.Run("berhasil", func(t *testing.T) {
		claims := jwt.Claims{}
		claims.Subject = "user@example.com"
		ctx := middlewares.WithUser(context.Background(), claims)
		got, err := IdentifierFromContext(ctx)
		if err != nil || got != "user@example.com" {
			t.Errorf("IdentifierFromContext() = %q, %v", got, err)
		}
	})
}
