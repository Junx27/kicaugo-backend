package organizer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Junx27/kicaugo-backend/internal/testutil"
)

func TestRepositoryResolveUserID(t *testing.T) {
	pool := testutil.OpenDB(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin transaksi test: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	identifier := fmt.Sprintf("organizer-resolve-%d@example.com", time.Now().UnixNano())
	var wantID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO users (name, identifier, password_hash)
		VALUES ('Organizer Test', $1, 'rahasia') RETURNING id`, identifier).Scan(&wantID); err != nil {
		t.Fatalf("membuat user test: %v", err)
	}

	gotID, err := (&repository{db: tx}).ResolveUserID(ctx, identifier)
	if err != nil {
		t.Fatalf("ResolveUserID() error = %v", err)
	}
	if gotID != wantID {
		t.Errorf("ResolveUserID() = %d, ingin %d", gotID, wantID)
	}
}
