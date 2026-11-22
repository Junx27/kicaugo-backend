package city

import (
	"context"
	"strings"
	"testing"

	"github.com/Junx27/kicaugo-backend/internal/testutil"
)

func TestRepositoryListByProvince(t *testing.T) {
	pool := testutil.OpenDB(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	var provinceID int64
	if err := pool.QueryRow(ctx, `SELECT id FROM provinces WHERE code = '11'`).Scan(&provinceID); err != nil {
		t.Fatalf("mengambil provinsi Aceh: %v", err)
	}

	items, err := repo.ListByProvince(ctx, provinceID, "")
	if err != nil {
		t.Fatalf("ListByProvince() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("ListByProvince() ingin ada data provinsi 11")
	}
	for i, item := range items {
		if item.ProvinceID != provinceID {
			t.Errorf("item[%d].ProvinceID = %d, ingin %d", i, item.ProvinceID, provinceID)
		}
		if i > 0 && strings.ToLower(items[i-1].Name) > strings.ToLower(item.Name) {
			t.Errorf("urutan nama salah: %q sebelum %q", items[i-1].Name, item.Name)
		}
	}

	filtered, err := repo.ListByProvince(ctx, provinceID, "Banda")
	if err != nil {
		t.Fatalf("ListByProvince(q) error = %v", err)
	}
	if len(filtered) == 0 {
		t.Fatal("ListByProvince(q) ingin ada hasil")
	}
	for _, item := range filtered {
		if !strings.Contains(strings.ToLower(item.Name), "banda") {
			t.Errorf("nama %q tidak mengandung banda", item.Name)
		}
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := repo.ListByProvince(canceled, provinceID, ""); err == nil {
		t.Error("ListByProvince() dengan konteks dibatalkan ingin error")
	}
}
