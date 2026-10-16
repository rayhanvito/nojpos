package service

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/adrr-dev/openPOS/backend/model"
)

var ErrBadTaxMonth = errors.New("format bulan tidak valid")

type TaxReportService struct {
	stores  StoreRepository
	reports ReportRepository
}

func NewTaxReportService(stores StoreRepository, reports ReportRepository) *TaxReportService {
	return &TaxReportService{stores: stores, reports: reports}
}

func (s *TaxReportService) Report(ctx context.Context, storeID uint, outletID *uint, month string) (*model.TaxReport, error) {
	timezone, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil || timezone == "" {
		timezone = "Asia/Makassar"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location, _ = time.LoadLocation("Asia/Makassar")
	}
	if month == "" {
		month = time.Now().In(location).Format("2006-01")
	}
	parsed, err := time.Parse("2006-01", month)
	if err != nil || len(month) != len("2006-01") || parsed.Format("2006-01") != month {
		return nil, ErrBadTaxMonth
	}
	start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, location)
	end := start.AddDate(0, 1, 0)

	report, err := s.reports.TaxReport(ctx, storeID, outletID, start, end, timezone)
	if err != nil {
		return nil, err
	}
	report.Month = month
	report.Net = report.Gross - report.RefundTotal
	report.PPhFinalEstimate = pphFinalEstimate(report.Net)
	return report, nil
}

func pphFinalEstimate(net int64) int64 {
	amount := new(big.Int).Mul(big.NewInt(net), big.NewInt(5))
	half := big.NewInt(500)
	if amount.Sign() < 0 {
		amount.Sub(amount, half)
	} else {
		amount.Add(amount, half)
	}
	return amount.Quo(amount, big.NewInt(1000)).Int64()
}
