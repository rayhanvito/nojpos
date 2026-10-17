package shift

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap (mis. service transaksi pada hasil
	// GetActiveShift) tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// ErrOutletNotFound adalah sentinel repo lama (definisi di
	// repo/base.go, pesan "outlet tidak ditemukan di toko" — BEDA
	// dari outlet.ErrOutletNotFound): produsennya validateOutletFilter
	// di repo slice ini, pemeriksanya respondShiftReadError di
	// handler slice ini.
	ErrOutletNotFound = errors.New("outlet tidak ditemukan di toko")

	// ErrCashierOutletChanged adalah alias ke sentinel kanonis slice
	// internal/staff (BE-F1b) — instans yang sama yang dulu di-alias
	// repo/base.go dan dikembalikan repo shift dari StartShift.
	// Sentinel polos service lama ("outlet kasir sudah berubah",
	// pesan identik) DISATUKAN ke instans ini: service slice kini
	// meneruskannya apa adanya (pass-through), jadi rantai errors.Is
	// handler tetap cocok pada instans yang sama.
	ErrCashierOutletChanged = staff.ErrCashierOutletChanged

	// Penyatuan dua lapis BE-F5: repo dan service lama masing-masing
	// mendefinisikan ErrShiftAlreadyOpen sebagai instans errors.New
	// terpisah dengan pesan identik ("shift masih aktif"); di slice
	// ini hanya ada satu instans kanonis — repo mengembalikannya dan
	// service meneruskannya, perilaku errors.Is identik.
	ErrShiftAlreadyOpen = errors.New("shift masih aktif")

	// ErrShiftClosed adalah sentinel tingkat repo lama (repo/shift.go):
	// dikembalikan ForceCloseShift repo; service memetakannya ke
	// ErrShiftAlreadyClosed di bawah.
	ErrShiftClosed = errors.New("shift sudah ditutup")

	// Sentinel service shift, dipindahkan verbatim dari
	// service/shift.go (BE-F5).
	ErrNoActiveShift      = errors.New("tidak ada shift yang berjalan")
	ErrShiftNotFound      = errors.New("shift tidak ditemukan")
	ErrShiftAlreadyClosed = errors.New("shift sudah ditutup")
)
