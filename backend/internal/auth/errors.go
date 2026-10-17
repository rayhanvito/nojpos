package auth

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan repo.ErrNotFound /
	// repo.ErrDuplicate, jadi errors.Is konsumen lama tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// ErrOutletNotFound dan ErrOutletForbidden adalah alias ke slice
	// outlet; badan Service di paket ini mengembalikannya.
	ErrOutletNotFound  = outlet.ErrOutletNotFound
	ErrOutletForbidden = outlet.ErrOutletForbidden
)

var (
	ErrInvalidCredentials  = errors.New("email atau kata sandi tidak cocok")
	ErrPasscodeRequired    = errors.New("passcode_required")
	ErrPasscodeWrong       = errors.New("passcode salah")
	ErrEmailTaken          = errors.New("email sudah terdaftar")
	ErrAccountInactive     = errors.New("akun dinonaktifkan")
	ErrTokenInvalid        = errors.New("sesi tidak valid, silakan masuk kembali")
	ErrSwitchSelf          = errors.New("tidak dapat beralih ke akun sendiri")
	ErrInvalidEmail        = errors.New("Email tidak valid.")
	ErrOtpCooldown         = errors.New("Terlalu sering meminta kode. Coba lagi dalam 60 detik.")
	ErrOtpWrong            = errors.New("Kode OTP salah.")
	ErrOtpExpired          = errors.New("Kode OTP sudah kedaluwarsa. Kirim ulang.")
	ErrOtpMaxAttempts      = errors.New("Terlalu banyak percobaan. Kirim ulang kode OTP.")
	ErrEmailNotVerified    = errors.New("Email belum diverifikasi. Silakan verifikasi kode OTP terlebih dahulu.")
	ErrGoogleInvalid       = errors.New("login Google tidak valid")
	ErrGoogleNotConfigured = errors.New("login Google belum dikonfigurasi di server")
	ErrOTPRequired         = errors.New("otp_required")
)
