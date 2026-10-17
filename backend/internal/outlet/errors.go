package outlet

import (
	"errors"
	"fmt"
	"net/http"

	apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"
)

var (
	// ErrNotFound dan ErrDuplicate adalah sentinel apierr yang sama
	// persis dengan yang dulu didefinisikan di repo/base.go (status,
	// code, dan pesan identik). repo/base.go kini meng-alias kedua
	// var ini (transisional) supaya semua konsumen menetap yang
	// memeriksa repo.ErrNotFound / repo.ErrDuplicate — misalnya
	// respondOutletErr di handler/outlet.go yang tinggal untuk
	// Switch — tetap cocok lewat errors.Is.
	ErrNotFound  = apierr.New(http.StatusNotFound, "NOT_FOUND", "tidak ditemukan")
	ErrDuplicate = apierr.New(http.StatusConflict, "DUPLICATE", "data sudah ada")

	// ErrOutletDefault dipindahkan verbatim dari repo/outlet.go;
	// repo/base.go meng-alias var ini untuk konsumen lama.
	ErrOutletDefault = errors.New("outlet default tidak dapat dinonaktifkan")

	// ErrOutletNotFound dan ErrOutletForbidden dipindahkan verbatim
	// dari service/outlet.go; paket service meng-alias keduanya
	// (transisional) karena AuthService dan UserService yang menetap
	// masih mengembalikannya.
	ErrOutletNotFound  = errors.New("outlet tidak ditemukan di toko Anda")
	ErrOutletForbidden = errors.New("kasir hanya dapat menggunakan outlet yang ditetapkan")
)

// Sentinel validasi pengaturan toko (definisi kanonis sejak BE-F1d;
// dipindahkan verbatim dari service/settings.go). Konsumen menetap di
// paket service (report/export memakai ErrBadPeriod) meng-alias
// instans di sini supaya errors.Is tetap cocok.
var (
	ErrBadTimezone          = fmt.Errorf("zona waktu tidak valid")
	ErrBadPeriod            = errors.New("periode tidak valid")
	ErrBadBusinessType      = errors.New("Jenis usaha tidak valid.")
	ErrBadStoreEmail        = errors.New("Email tidak valid.")
	ErrBadCurrency          = errors.New("Kode mata uang tidak valid.")
	ErrBadHours             = errors.New("Jam operasional tidak valid.")
	ErrBadReceiptFooter     = errors.New("Pesan footer maksimal 200 karakter.")
	ErrBadTaxName           = errors.New("Nama pajak maksimal 20 karakter.")
	ErrBadTaxPct            = errors.New("Tarif pajak maksimal 100 persen.")
	ErrBadTaxRounding       = errors.New("Pembulatan pajak tidak valid.")
	ErrBadTaxApplyTo        = errors.New("Cakupan pajak tidak valid.")
	ErrBadLoyaltyEarnPer    = errors.New("loyaltyEarnPer harus lebih dari 0.")
	ErrBadLoyaltyPointValue = errors.New("Nilai tukar poin tidak boleh negatif.")
	ErrBadCreditDueDays     = errors.New("Jangka waktu piutang tidak boleh negatif.")
	ErrBadLowStockDefault   = errors.New("Nilai stok minimum bawaan tidak boleh negatif.")
)
