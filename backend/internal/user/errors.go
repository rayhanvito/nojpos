package user

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrStoreMismatch adalah definisi kanonis (BE-F1d; sebelumnya di
	// service/user.go): paket service meng-alias var ini karena
	// service katalog/stok yang menetap masih mengembalikannya dan
	// handler katalog/stok memeriksanya lewat errors.Is.
	ErrStoreMismatch    = errors.New("akun tidak ditemukan di toko Anda")
	ErrNotEditable      = errors.New("hanya akun kasir yang dapat dinonaktifkan")
	ErrNotRenamable     = errors.New("hanya akun kasir yang dapat diubah")
	ErrCashierOpenShift = errors.New("kasir memiliki shift terbuka dan tidak dapat dipindahkan outlet")

	// ErrNotFound adalah alias ke instans kanonis slice outlet —
	// instans yang sama dengan repo.ErrNotFound, jadi pemeriksaan
	// errors.Is pada hasil repo kasir/user tetap cocok.
	ErrNotFound = outlet.ErrNotFound

	// ErrOutletNotFound adalah alias ke slice outlet; badan Service
	// di paket ini mengembalikannya.
	ErrOutletNotFound = outlet.ErrOutletNotFound
)
