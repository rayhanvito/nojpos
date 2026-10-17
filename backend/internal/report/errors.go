package report

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound adalah alias ke instans kanonis slice outlet (BE-F1a)
	// - instans yang sama dengan repo.ErrNotFound lama; periodStarts di
	// repo slice ini mengembalikannya dan Handler memeriksanya lewat
	// var ini.
	ErrNotFound = outlet.ErrNotFound

	// ErrBadPeriod adalah alias ke definisi kanonis slice
	// internal/outlet (BE-F1d) - instans yang sama dengan
	// service.ErrBadPeriod lama; Service dan ExportService slice ini
	// mengembalikannya dan handler memeriksanya lewat var ini.
	ErrBadPeriod = outlet.ErrBadPeriod

	// ErrOutletNotFound di slice ini BUKAN sentinel slice outlet
	// (pesannya berbeda): definisinya dipindah verbatim dari
	// repo/base.go, produsennya Repo laporan (validateReportOutlet)
	// dan pemeriksanya handler slice ini.
	ErrOutletNotFound = errors.New("outlet tidak ditemukan di toko")

	// ErrBadTaxMonth dipindahkan verbatim dari service/tax_report.go
	// (BE-F10); produsennya TaxService dan pemeriksanya TaxHandler.
	ErrBadTaxMonth = errors.New("format bulan tidak valid")
)
