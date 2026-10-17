package report

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// TaxHandler adalah TaxReportHandler lama di handler/tax_report.go
// (BE-F10), dipindah verbatim.
type TaxHandler struct {
	svc *TaxService
}

func NewTaxHandler(svc *TaxService) *TaxHandler {
	return &TaxHandler{svc: svc}
}

func (h *TaxHandler) Get(c *gin.Context) {
	report, ok := h.report(c)
	if ok {
		c.JSON(http.StatusOK, report)
	}
}

func (h *TaxHandler) Export(c *gin.Context) {
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

func (h *TaxHandler) report(c *gin.Context) (*TaxReport, bool) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := reportOutletFilter(c, claims.Role == middleware.RoleAdmin, claims.OutletID)
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
		if errors.Is(err, ErrOutletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
		} else if errors.Is(err, ErrBadTaxMonth) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Bulan tidak valid. Gunakan format YYYY-MM."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat laporan pajak."})
		}
		return nil, false
	}
	return report, true
}
