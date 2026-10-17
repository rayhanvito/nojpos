package catalog

import (
	"errors"
	"net/http"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan repo.ErrNotFound /
	// repo.ErrDuplicate, jadi pemeriksaan errors.Is konsumen menetap
	// (mis. service/inventory.go pada hasil repo katalog) tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// Sentinel katalog, dipindahkan verbatim dari service/catalog.go.
	ErrCategoryInvalid = apierr.New(http.StatusBadRequest, "CATEGORY_INVALID", "Kategori tidak ditemukan di toko Anda.")
	ErrCategoryTaken   = apierr.New(http.StatusConflict, "CATEGORY_TAKEN", "Kategori dengan nama itu sudah ada.")
	ErrSkuTaken        = apierr.New(http.StatusConflict, "SKU_TAKEN", "SKU sudah digunakan di toko ini.")

	// ErrStockInsufficient: definisi kanonis pindah dari repo/product.go
	// (BE-B). repo/base.go meng-alias var ini supaya produsen/pemeriksa
	// yang menetap (repo/stock.go, repo/transaction.go,
	// handler/transaction.go, service/inventory.go) tetap menunjuk
	// instans yang sama.
	ErrStockInsufficient = apierr.New(http.StatusBadRequest, "STOCK_INSUFFICIENT", "stok tidak cukup")

	// ErrVariantRequired: definisi kanonis pindah dari repo/transaction.go
	// karena salah satu produsennya (penyesuaian stok produk bervarian)
	// berada di jalur slice ini dan paket repo tidak dapat diimpor dari
	// sini (siklus). Sejak Unit B produsennya adalah orkestrasi
	// AdjustStock di slice inventory, yang mengimpor instans ini.
	// repo/transaction.go meng-alias var ini supaya pemeriksa menetap
	// (handler/transaction.go) tetap menunjuk instans yang sama.
	ErrVariantRequired = errors.New("varian produk wajib dipilih")
)
