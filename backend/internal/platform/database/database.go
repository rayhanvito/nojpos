// Package database menyediakan fondasi akses data bersama: BaseRepo
// memberi repo akses *gorm.DB yang sadar konteks, dan Transactor
// menjalankan unit kerja dalam transaksi dengan semantik JOIN —
// bila ctx sudah membawa transaksi aktif, WithinTx tidak membuka
// transaksi/savepoint baru, melainkan menjalankan fn dengan ctx yang
// sama sehingga error merambat ke pemilik transaksi terluar.
package database

import (
	"context"

	"gorm.io/gorm"
)

// txKey adalah kunci context privat untuk menyimpan *gorm.DB transaksi aktif.
type txKey struct{}

// txFromContext mengembalikan transaksi aktif yang dibawa ctx, bila ada.
func txFromContext(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(txKey{}).(*gorm.DB)
	return tx, ok && tx != nil
}

// InTx melaporkan apakah ctx membawa transaksi aktif. Konsumen sink
// (mis. notifikasi, §4) memakainya untuk memilih semantik best-effort
// di dalam transaksi pemanggil versus perilaku biasa di luarnya.
func InTx(ctx context.Context) bool {
	_, ok := txFromContext(ctx)
	return ok
}

// withinTx menjalankan fn dalam transaksi milik db: join bila ctx sudah
// membawa transaksi aktif, buka transaksi baru bila belum.
func withinTx(db *gorm.DB, ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		return fn(ctx)
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// BaseRepo adalah fondasi repo: field db tidak diekspor, satu-satunya
// akses repo ke database adalah lewat DB(ctx).
type BaseRepo struct {
	db *gorm.DB
}

// NewBaseRepo membuat BaseRepo dari handle database yang ada.
func NewBaseRepo(db *gorm.DB) BaseRepo {
	return BaseRepo{db: db}
}

// DB mengembalikan handle untuk ctx ini: transaksi aktif bila ctx
// membawanya, koneksi dasar bila tidak.
func (b BaseRepo) DB(ctx context.Context) *gorm.DB {
	if tx, ok := txFromContext(ctx); ok {
		return tx.WithContext(ctx)
	}
	return b.db.WithContext(ctx)
}

// WithinTx menjalankan fn dalam transaksi (semantik JOIN, lihat paket).
func (b BaseRepo) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return withinTx(b.db, ctx, fn)
}

// Transactor menjalankan unit kerja lintas-repo dalam satu transaksi
// (semantik JOIN, lihat paket).
type Transactor struct {
	db *gorm.DB
}

// NewTransactor membuat Transactor dari handle database yang ada.
func NewTransactor(db *gorm.DB) *Transactor {
	return &Transactor{db: db}
}

// WithinTx menjalankan fn dalam transaksi (semantik JOIN, lihat paket).
func (t *Transactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return withinTx(t.db, ctx, fn)
}
