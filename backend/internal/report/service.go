package report

import (
	"context"

	"time"
)

// Service memegang dashboard dan laporan (ReportService lama di
// service/report.go, BE-F10). Keluar dari SettingsService
// sejak BE-B5; badan metodenya dipindah persis (validasi periode + zona waktu
// toko dengan fallback Asia/Makassar), hanya pemiliknya yang berganti.
//
// Catatan: brief BE-B5 menulis ReportService{reports ReportRepository} dengan
// NewReportService(reports), tetapi logika saat ini wajib membaca zona waktu
// toko (StoreRepository) sebelum memanggil repo - pola yang sama persis
// dengan TaxReportService{stores, reports}. Tanpa stores, zona waktu tidak
// bisa didapat dan perilaku laporan akan berubah, jadi stores ikut dibawa.
// Di slice ini kedua kontrak itu menjadi interface lokal Stores dan Reports.
// validPeriods adalah daftar periode laporan yang sah (dipindahkan
// dari service/settings.go lama di BE-F1d; pemakainya memang hanya
// ReportService dan ExportService di paket ini).
var validPeriods = map[string]bool{"": true, "today": true, "yesterday": true, "week": true, "month": true, "all": true}

// Stores adalah kontrak toko yang dipakai Service, TaxService, dan
// ExportService slice ini - hanya GetTimezone dari StoreRepository
// lama yang benar-benar dipakai (diimplementasikan outlet.StoreRepo).
type Stores interface {
	GetTimezone(ctx context.Context, storeID uint) (string, error)
}

// Reports adalah kontrak penyimpanan laporan yang dipakai Service dan
// TaxService (ReportRepository lama di service/interfaces.go, bertipe
// slice ini).
type Reports interface {
	Dashboard(ctx context.Context, storeID, cashierID uint, cashierView bool, outletID *uint, tz, dateStr, fromStr, toStr string) (any, error)
	Report(ctx context.Context, storeID uint, outletID *uint, period, tz, dateStr, fromStr, toStr string) (*ReportBundle, error)
	TaxReport(ctx context.Context, storeID uint, outletID *uint, start, end time.Time, timezone string) (*TaxReport, error)
}

type Service struct {
	stores  Stores
	reports Reports
}

func NewService(stores Stores, reports Reports) *Service {
	return &Service{stores: stores, reports: reports}
}

func (s *Service) Dashboard(ctx context.Context, storeID, cashierID uint, cashierView bool, outletID *uint, dateStr, fromStr, toStr string) (any, error) {
	tz, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil {
		tz = "Asia/Makassar"
	}
	return s.reports.Dashboard(ctx, storeID, cashierID, cashierView, outletID, tz, dateStr, fromStr, toStr)
}

func (s *Service) Report(ctx context.Context, storeID uint, outletID *uint, period, dateStr, fromStr, toStr string) (*ReportBundle, error) {
	if !validPeriods[period] {
		return nil, ErrBadPeriod
	}
	tz, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil {
		tz = "Asia/Makassar"
	}
	return s.reports.Report(ctx, storeID, outletID, period, tz, dateStr, fromStr, toStr)
}
