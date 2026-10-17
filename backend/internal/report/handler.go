package report

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler melayani /dashboard dan /reports (ReportHandler lama di
// handler/report.go, BE-F10). Keluar dari SettingsHandler sejak BE-B5;
// badan metodenya dipindah persis dari handler/settings.go.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Dashboard(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	// Cashiers must be scoped to their own cashier ID, not the owner ID.
	cashierID := claims.UserID
	if claims.ActingAsCashierID != nil {
		cashierID = *claims.ActingAsCashierID
	}
	outletID, ok := reportOutletFilter(c, claims.Role == middleware.RoleAdmin, claims.OutletID)
	if !ok {
		return
	}
	if !validRangeOr400(c) {
		return
	}
	data, err := h.svc.Dashboard(c.Request.Context(), claims.StoreID, cashierID, claims.Role != middleware.RoleAdmin, outletID, c.Query("date"), c.Query("from"), c.Query("to"))
	if err != nil {
		if errors.Is(err, ErrOutletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat dashboard."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) Report(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := reportOutletFilter(c, claims.Role == middleware.RoleAdmin, claims.OutletID)
	if !ok {
		return
	}
	if !validRangeOr400(c) {
		return
	}
	bundle, err := h.svc.Report(c.Request.Context(), claims.StoreID, outletID, c.Query("period"), c.Query("date"), c.Query("from"), c.Query("to"))
	if err != nil {
		switch {
		case errors.Is(err, ErrOutletNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
		case errors.Is(err, ErrBadPeriod), errors.Is(err, ErrNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Periode tidak valid."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat laporan."})
		}
		return
	}
	c.JSON(http.StatusOK, bundle)
}

// reportOutletFilter dan validRangeOr400 adalah salinan privat slice
// dari helper paket handler lama (handler/params.go) - badan verbatim.
func reportOutletFilter(c *gin.Context, admin bool, activeOutletID uint) (*uint, bool) {
	if !admin {
		return &activeOutletID, true
	}
	value, provided := c.GetQuery("outletId")
	if !provided {
		return nil, true
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outletId tidak valid."})
		return nil, false
	}
	outletID := uint(id)
	return &outletID, true
}

// validRangeOr400 menolak pasangan from/to yang terbalik (keduanya tanggal
// valid tetapi to < from). Nilai yang tidak terurai mempertahankan perilaku
// lama (diabaikan oleh lapisan laporan); hanya rentang terbalik yang sah
// secara format yang kini ditolak tegas, bukan diam-diam diganti periode
// bawaan (temuan pembuktian kelompok 1, 2026-10-16).
func validRangeOr400(c *gin.Context) bool {
	from, to := c.Query("from"), c.Query("to")
	if from == "" || to == "" {
		return true
	}
	f, errF := time.Parse("2006-01-02", from)
	tt, errT := time.Parse("2006-01-02", to)
	if errF != nil || errT != nil {
		return true
	}
	if tt.Before(f) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rentang tanggal tidak valid: tanggal dari harus sebelum tanggal sampai."})
		return false
	}
	return true
}
