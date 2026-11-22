package testutil

import (
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/utils/jwt"
)

const (
	// TestIdentifier adalah subject token yang dipakai test handler.
	TestIdentifier = "user@example.com"
	testSecret     = "rahasia-untuk-test"
)

// NewTokenManager membuat manager JWT dengan TTL dan secret khusus test.
func NewTokenManager() *jwt.Manager {
	return jwt.NewManager([]byte(testSecret), time.Hour)
}

// BearerHeader menghasilkan header Authorization untuk subject test.
func BearerHeader(t *testing.T, tokens *jwt.Manager) string {
	t.Helper()
	token, err := tokens.Issue(TestIdentifier, "Tester")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	return "Authorization: Bearer " + token
}
