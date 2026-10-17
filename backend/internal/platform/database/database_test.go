package database

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testRow struct {
	ID   uint `gorm:"primarykey"`
	Name string
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "database.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&testRow{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func countRows(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var count int64
	if err := db.Model(&testRow{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

func TestDBWithoutTxUsesBaseConnection(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	ctx := context.Background()

	if err := repo.DB(ctx).Create(&testRow{Name: "dasar"}).Error; err != nil {
		t.Fatal(err)
	}
	var row testRow
	if err := repo.DB(ctx).Where("name = ?", "dasar").First(&row).Error; err != nil {
		t.Fatal(err)
	}
	if row.Name != "dasar" {
		t.Fatalf("baris terbaca salah: %+v", row)
	}
}

func TestWithinTxCommitsOnSuccess(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	ctx := context.Background()

	err := repo.WithinTx(ctx, func(ctx context.Context) error {
		return repo.DB(ctx).Create(&testRow{Name: "commit"}).Error
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := countRows(t, db); got != 1 {
		t.Fatalf("baris harus ada sesudah commit, dapat %d", got)
	}
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	ctx := context.Background()

	err := repo.WithinTx(ctx, func(ctx context.Context) error {
		if err := repo.DB(ctx).Create(&testRow{Name: "rollback"}).Error; err != nil {
			return err
		}
		return errors.New("paksa gagal")
	})
	if err == nil {
		t.Fatal("WithinTx harus mengembalikan error dari fn")
	}
	if got := countRows(t, db); got != 0 {
		t.Fatalf("baris tidak boleh ada sesudah rollback, dapat %d", got)
	}
}

func TestWithinTxNestedJoinRollsBackEverything(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	tr := NewTransactor(db)
	ctx := context.Background()

	err := tr.WithinTx(ctx, func(ctx context.Context) error {
		if err := repo.WithinTx(ctx, func(ctx context.Context) error {
			return repo.DB(ctx).Create(&testRow{Name: "A"}).Error
		}); err != nil {
			return err
		}
		if err := repo.DB(ctx).Create(&testRow{Name: "B"}).Error; err != nil {
			return err
		}
		return errors.New("gagal di langkah terakhir")
	})
	if err == nil {
		t.Fatal("WithinTx luar harus mengembalikan error")
	}
	// A ditulis oleh WithinTx dalam: bila dalam membuka savepoint/transaksi
	// sendiri dan commit, A akan tetap ada. Semantik JOIN: keduanya hilang.
	if got := countRows(t, db); got != 0 {
		t.Fatalf("A dan B keduanya harus hilang (join, bukan savepoint), dapat %d baris", got)
	}
}

func TestWithinTxInnerErrorPropagatesAndRollsBackOuter(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	ctx := context.Background()
	sentinel := errors.New("error dari dalam")

	err := repo.WithinTx(ctx, func(ctx context.Context) error {
		if err := repo.DB(ctx).Create(&testRow{Name: "luar"}).Error; err != nil {
			return err
		}
		return repo.WithinTx(ctx, func(ctx context.Context) error {
			if err := repo.DB(ctx).Create(&testRow{Name: "dalam"}).Error; err != nil {
				return err
			}
			return sentinel
		})
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error dalam harus merambat apa adanya, dapat %v", err)
	}
	if got := countRows(t, db); got != 0 {
		t.Fatalf("semua tulisan luar harus ikut gagal, dapat %d baris", got)
	}
}

func TestWithinTxWritesViaDBJoinSameTx(t *testing.T) {
	db := newTestDB(t)
	repo := NewBaseRepo(db)
	tr := NewTransactor(db)
	ctx := context.Background()

	err := tr.WithinTx(ctx, func(ctx context.Context) error {
		if err := repo.DB(ctx).Create(&testRow{Name: "lewat-db-ctx"}).Error; err != nil {
			return err
		}
		// Tulisan lewat DB(ctx) harus terlihat oleh query dalam tx yang sama.
		var count int64
		if err := repo.DB(ctx).Model(&testRow{}).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return errors.New("tulisan DB(ctx) tidak terlihat dalam tx yang sama")
		}
		return errors.New("paksa gagal sesudah menulis")
	})
	if err == nil {
		t.Fatal("WithinTx harus mengembalikan error paksa")
	}
	if got := countRows(t, db); got != 0 {
		t.Fatalf("tulisan DB(ctx) harus hilang sesudah rollback, dapat %d baris", got)
	}
}
