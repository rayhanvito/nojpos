package finance

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// Sentinel repo utang-piutang, dipindahkan verbatim dari
	// repo/debt.go (BE-F6).
	ErrDebtPaymentInvalid       = errors.New("pembayaran hutang-piutang tidak valid")
	ErrDebtPaymentExceedsRemain = errors.New("pembayaran melebihi sisa hutang-piutang")

	// Sentinel service utang-piutang, dipindahkan verbatim dari
	// service/debt.go (BE-F6).
	ErrDebtStatusInvalid = errors.New("status hutang-piutang tidak valid")
	ErrDebtPaymentInput  = errors.New("pembayaran hutang-piutang tidak valid")

	// Sentinel service pengeluaran, dipindahkan verbatim dari
	// service/expense.go (BE-F6).
	ErrExpenseNotFound = errors.New("pengeluaran tidak ditemukan")
	ErrExpenseInvalid  = errors.New("data pengeluaran tidak valid")
)
