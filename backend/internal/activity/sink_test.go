package activity

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// TestActivityLogFollowsCallerTransaction membuktikan semantik sink
// §4 untuk audit: Log yang dipanggil dengan ctx dari transaksi
// pemanggil ikut transaksi itu secara atomik — rollback pemanggil
// menghapus baris activity_log, commit pemanggil menyimpannya. Tidak
// ada savepoint di slice ini: audit gagal/commit bersama bisnisnya.
func TestActivityLogFollowsCallerTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "activity-sink.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&ActivityLog{}); err != nil {
		t.Fatal(err)
	}

	svc := NewService(NewRepo(db))
	tr := database.NewTransactor(db)
	ctx := context.Background()

	countLogs := func() int64 {
		var count int64
		if err := db.Model(&ActivityLog{}).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		return count
	}

	// Rollback: Log di dalam WithinTx yang dipaksa gagal.
	err = tr.WithinTx(ctx, func(ctx context.Context) error {
		svc.Log(ctx, 1, "7", "Admin Uji", ActivityLogin, "Admin Uji login ke sistem", "", "")
		return errors.New("paksa rollback transaksi pemanggil")
	})
	if err == nil {
		t.Fatal("WithinTx harus mengembalikan error pemaksa rollback")
	}
	if got := countLogs(); got != 0 {
		t.Fatalf("baris activity_log tidak boleh ada sesudah rollback pemanggil, dapat %d", got)
	}

	// Kontrol commit: Log di dalam WithinTx yang berhasil.
	err = tr.WithinTx(ctx, func(ctx context.Context) error {
		svc.Log(ctx, 1, "7", "Admin Uji", ActivityLogin, "Admin Uji login ke sistem", "", "")
		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx kontrol harus commit: %v", err)
	}
	if got := countLogs(); got != 1 {
		t.Fatalf("baris activity_log harus ada sesudah commit pemanggil, dapat %d", got)
	}

	// Baris yang commit juga terbaca lewat List service.
	logs, total, err := svc.List(ctx, 1, ActivityFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(logs) != 1 || logs[0].Action != ActivityLogin {
		t.Fatalf("List harus membaca 1 log LOGIN, dapat total=%d logs=%+v", total, logs)
	}
}
