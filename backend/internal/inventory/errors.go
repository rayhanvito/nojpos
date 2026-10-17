package inventory

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
	// tetap cocok pada hasil mesin stok dan repo slice ini.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// ErrBadDirection dipindahkan verbatim dari service/inventory.go
	// (BE-F3-2): pesan mengikuti render handler stok.
	ErrBadDirection = apierr.New(http.StatusBadRequest, "BAD_DIRECTION", "arah penyesuaian harus 'plus' atau 'minus'")

	// Sentinel stock opname. Tingkat repo (repo/stock_opname.go lama)
	// dan tingkat service (service/stock_opname.go lama) dulu memiliki
	// instans terpisah untuk Active dan State dengan pesan yang
	// identik; di dalam satu paket slice ini keduanya disatukan menjadi
	// satu instans kanonis per nama (pesan tidak berubah, dan
	// mapOpnameError + handler kini berada di paket yang sama).
	ErrStockOpnameNotFound = errors.New("sesi stock opname tidak ditemukan")
	ErrStockOpnameInvalid  = errors.New("data stock opname tidak valid")
	ErrStockOpnameActive   = errors.New("sesi stock opname aktif sudah ada")
	ErrStockOpnameState    = errors.New("status sesi stock opname tidak mengizinkan aksi ini")
	ErrStockOpnameItem     = errors.New("item tidak sesuai dengan sesi stock opname")
	ErrStockOpnameTarget   = errors.New("produk atau varian tujuan stok tidak tersedia")
)
