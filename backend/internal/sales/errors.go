package sales

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap tetap cocok. Khusus gerbang shift di
	// Service: var alias warisan repo.ErrNotFound adalah instans
	// outlet.ErrNotFound yang sama, dan slice ini memeriksanya lewat
	// instans kanonis outlet secara langsung.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// ErrStockInsufficient dan ErrVariantRequired adalah alias ke
	// instans kanonis slice catalog (BE-F2) — produsen di repo lama
	// mengembalikan instans itu lewat alias repo.ErrStockInsufficient
	// / definisi repo.ErrVariantRequired = catalog.ErrVariantRequired,
	// dan pemeriksa di handler slice ini tetap menunjuk instans yang
	// sama.
	ErrStockInsufficient = catalog.ErrStockInsufficient
	ErrVariantRequired   = catalog.ErrVariantRequired

	// Sentinel repo transaksi, dipindahkan verbatim dari
	// repo/transaction.go (BE-F8) — definisi kanonisnya kini di slice
	// ini. Audit grep repo-wide sebelum pemindahan: produsennya hanya
	// repo transaksi lama dan pemeriksanya hanya handler transaksi
	// lama, jadi tidak ada alias warisan yang dipertahankan di paket
	// repo/service.
	ErrPaidInsufficient       = errors.New("jumlah bayar kurang dari total")
	ErrBadDiscount            = errors.New("diskon melebihi subtotal")
	ErrProductInactive        = errors.New("produk tidak aktif")
	ErrNotRefundable          = errors.New("transaksi ini tidak dapat direfund")
	ErrRefundTooMuch          = errors.New("qty refund melebihi jumlah terjual")
	ErrEmptyItems             = errors.New("keranjang kosong")
	ErrPromoVoucherInvalid    = errors.New("voucher tidak valid")
	ErrCustomerNotFound       = errors.New("pelanggan tidak ditemukan di toko ini")
	ErrPointsRedeemInvalid    = errors.New("penukaran poin tidak valid")
	ErrPointsCustomerRequired = errors.New("pelanggan wajib dipilih untuk menukar poin")
	ErrPointsBalance          = errors.New("saldo poin pelanggan tidak mencukupi")
	ErrPointsExceedTotal      = errors.New("nilai penukaran poin melebihi total transaksi")
	ErrLoyaltySettings        = errors.New("pengaturan loyalty tidak valid")
	ErrCreditCustomerRequired = errors.New("metode Credit wajib memilih pelanggan")
	ErrCreditPaid             = errors.New("field paid harus 0 untuk metode Credit")
	ErrVariantInvalid         = errors.New("varian produk tidak valid")

	// ErrShiftRequired: sentinel service transaksi, dipindahkan
	// verbatim dari service/transaction.go (BE-F8) — checkout ditolak
	// karena kasir pelaku belum punya shift aktif di outlet ini
	// (keputusan Rayhan 2026-10-16: tidak ada transaksi tanpa shift —
	// menggantikan gerbang lunak F2.7). Audit grep: pemeriksanya
	// hanya handler transaksi lama, jadi tidak ada alias warisan
	// service.ErrShiftRequired yang dipertahankan.
	ErrShiftRequired = errors.New("belum ada shift aktif")
)
