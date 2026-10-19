package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestImportCategoriesReportsCreateErrorPerRow(t *testing.T) {
	cats := &fakeCats{createErr: errors.New("database unavailable")}
	svc := NewService(cats, nil, nil)

	result, err := svc.ImportCategories(context.Background(), 7, [][]string{{"Kategori Baru"}})
	if err != nil {
		t.Fatalf("kesalahan satu baris tidak boleh membatalkan impor: %v", err)
	}
	if result.Total != 1 || result.Berhasil != 0 || result.Gagal != 1 {
		t.Fatalf("ringkasan salah: %+v", result)
	}
	if len(result.Kesalahan) != 1 || result.Kesalahan[0] != (ImportRowFailure{Baris: 2, Alasan: "database unavailable"}) {
		t.Fatalf("kesalahan baris tidak dilaporkan: %+v", result.Kesalahan)
	}
}
