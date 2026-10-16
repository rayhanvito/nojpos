package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type TaxReportHandler struct {
	svc *service.TaxReportService
}

func NewTaxReportHandler(svc *service.TaxReportService) *TaxReportHandler {
	return &TaxReportHandler{svc: svc}
}

func (h *TaxReportHandler) Get(c *gin.Context) {
	report, ok := h.report(c)
	if ok {
		c.JSON(http.StatusOK, report)
	}
}

func (h *TaxReportHandler) Export(c *gin.Context) {
	report, ok := h.report(c)
	if !ok {
		return
	}
	writeCSV(c, "laporan-pajak-"+report.Month+".csv", []string{
		"bulan", "omzet bruto", "total refund", "omzet neto", "DPP", "PPN terkumpul", "jumlah transaksi", "jumlah refund", "estimasi PPh Final",
	}, func(w *csv.Writer) {
		_ = w.Write([]string{
			report.Month,
			strconv.FormatInt(report.Gross, 10),
			strconv.FormatInt(report.RefundTotal, 10),
			strconv.FormatInt(report.Net, 10),
			strconv.FormatInt(report.DPP, 10),
			strconv.FormatInt(report.PPNCollected, 10),
			strconv.FormatInt(report.TrxCount, 10),
			strconv.FormatInt(report.RefundCount, 10),
			strconv.FormatInt(report.PPhFinalEstimate, 10),
		})
	})
}

func (h *TaxReportHandler) report(c *gin.Context) (*model.TaxReport, bool) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := reportOutletFilter(c, claims.Role == model.RoleAdmin, claims.OutletID)
	if !ok {
		return nil, false
	}
	month, provided := c.GetQuery("month")
	if provided && month == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bulan tidak valid. Gunakan format YYYY-MM."})
		return nil, false
	}
	report, err := h.svc.Report(c.Request.Context(), claims.StoreID, outletID, month)
	if err != nil {
		if errors.Is(err, repo.ErrOutletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
		} else if errors.Is(err, service.ErrBadTaxMonth) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bulan tidak valid. Gunakan format YYYY-MM."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat laporan pajak."})
		}
		return nil, false
	}
	return report, true
}
