package customer

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap (mis. service transaksi pada hasil
	// repo pelanggan) tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// Sentinel service pelanggan, dipindahkan verbatim dari
	// service/customer.go (BE-F4).
	ErrCustomerNotFound     = errors.New("pelanggan tidak ditemukan")
	ErrCustomerPhoneTaken   = errors.New("nomor telepon sudah digunakan di toko ini")
	ErrCustomerNameRequired = errors.New("nama pelanggan wajib diisi")
)
