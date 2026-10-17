package promo

import "errors"

var (
	ErrNotFound  = errors.New("tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
	ErrInvalid   = errors.New("data promo tidak valid")
	ErrCodeTaken = errors.New("kode voucher sudah digunakan di toko ini")
)
