package shift

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah ShiftHandler lama di handler/shift.go, dipindah
// verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// pathUint, reportOutletFilter, clampPage, clampLimit, writeCSV,
// safeCSVText, dan exportLocation adalah salinan privat slice dari
// helper paket handler lama (handler/params.go dan handler/export.go)
// — badan verbatim.
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

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

func clampLimit(limit int) int {
	if limit < 1 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func clampPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func writeCSV(c *gin.Context, filename string, header []string, writeRows func(*csv.Writer)) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w := csv.NewWriter(c.Writer)
	if err := w.Write(header); err != nil {
		return
	}
	writeRows(w)
	w.Flush()
}

func safeCSVText(value string) string {
	// Prevent spreadsheet apps from treating user text as a formula.
	trimmed := strings.TrimLeftFunc(value, unicode.IsSpace)
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

func exportLocation(timezone string) *time.Location {
	loc, err := time.LoadLocation(timezone)
	if err == nil {
		return loc
	}
	loc, err = time.LoadLocation("Asia/Makassar")
	if err == nil {
		return loc
	}
	return time.UTC
}

func (h *Handler) GetCurrentShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	data, err := h.svc.GetCashierShift(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name)
	if err != nil {
		if errors.Is(err, ErrNoActiveShift) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tidak ada shift yang berjalan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat shift."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) StartShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req startShiftReq
	_ = c.ShouldBindJSON(&req)

	data, err := h.svc.StartShift(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name, req.OpeningCash)
	if err != nil {
		if errors.Is(err, ErrShiftAlreadyOpen) {
			c.JSON(http.StatusConflict, gin.H{"error": "Shift masih aktif. Tutup shift yang berjalan dulu sebelum membuka yang baru."})
			return
		}
		if errors.Is(err, ErrCashierOutletChanged) {
			c.JSON(http.StatusConflict, gin.H{"error": "Outlet kasir sudah berubah. Masuk kembali ke akun kasir untuk membuka shift."})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, data)
}

func (h *Handler) CloseShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req closeShiftReq
	if err := c.ShouldBindJSON(&req); err != nil || req.ClosingCash == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "closingCash wajib diisi dan tidak boleh negatif."})
		return
	}
	if *req.ClosingCash < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "closingCash tidak boleh negatif."})
		return
	}
	data, err := h.svc.CloseShift(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name, *req.ClosingCash, optionalShiftNote(req.Note))
	if errors.Is(err, ErrNoActiveShift) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada shift yang berjalan."})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal menutup shift."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) ListShifts(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := reportOutletFilter(c, true, claims.OutletID)
	if !ok {
		return
	}
	items, err := h.svc.ListShifts(c.Request.Context(), claims.StoreID, outletID, c.Query("from"), c.Query("to"))
	if err != nil {
		respondShiftReadError(c, err, "Gagal memuat daftar shift.")
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}

func (h *Handler) GetShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	shiftID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	outletID, ok := reportOutletFilter(c, true, claims.OutletID)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	detail, err := h.svc.GetShift(c.Request.Context(), claims.StoreID, shiftID, outletID, clampPage(page), clampLimit(limit))
	if err != nil {
		respondShiftReadError(c, err, "Gagal memuat detail shift.")
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) ForceCloseShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	shiftID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req closeShiftReq
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body JSON tidak valid."})
		return
	}
	if req.ClosingCash != nil && *req.ClosingCash < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "closingCash tidak boleh negatif."})
		return
	}
	data, err := h.svc.ForceCloseShift(c.Request.Context(), claims.StoreID, shiftID, req.ClosingCash, optionalShiftNote(req.Note), claims.Name)
	if err != nil {
		switch {
		case errors.Is(err, ErrShiftNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Shift tidak ditemukan di toko Anda."})
		case errors.Is(err, ErrShiftAlreadyClosed):
			c.JSON(http.StatusConflict, gin.H{"error": "Shift sudah ditutup."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menutup shift secara paksa."})
		}
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) ExportShifts(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := reportOutletFilter(c, true, claims.OutletID)
	if !ok {
		return
	}
	shifts, timezone, err := h.svc.ExportShifts(c.Request.Context(), claims.StoreID, outletID, c.Query("from"), c.Query("to"))
	if err != nil {
		respondShiftReadError(c, err, "Gagal mengekspor laporan shift.")
		return
	}
	location := exportLocation(timezone)
	filename := "laporan-shift-" + time.Now().In(location).Format("2006-01-02") + ".csv"
	if c.Query("from") != "" && c.Query("to") != "" {
		filename = "laporan-shift-" + c.Query("from") + "_" + c.Query("to") + ".csv"
	}
	methods := ShiftPaymentMethods()
	header := []string{"id", "kasir", "mulai", "selesai", "modal awal"}
	for _, method := range methods {
		header = append(header, "penjualan "+method)
	}
	header = append(header, "refund", "total penjualan", "jumlah transaksi", "expected", "aktual", "selisih")
	writeCSV(c, filename, header, func(writer *csv.Writer) {
		for _, shift := range shifts {
			row := []string{
				strconv.FormatUint(uint64(shift.ID), 10),
				safeCSVText(shift.CashierName),
				shift.StartedAt.In(location).Format(time.RFC3339),
				formatShiftTime(shift.ClosedAt, location),
				strconv.FormatInt(shift.OpeningCash, 10),
			}
			for _, method := range methods {
				row = append(row, strconv.FormatInt(shift.SalesByMethod[method], 10))
			}
			row = append(row,
				strconv.FormatInt(shift.RefundTotal, 10),
				strconv.FormatInt(shift.Sales, 10),
				strconv.FormatInt(shift.TrxCount, 10),
				strconv.FormatInt(shift.ExpectedCash, 10),
				formatShiftInt(shift.ClosingCashActual),
				formatShiftInt(shift.CashVariance),
			)
			if err := writer.Write(row); err != nil {
				return
			}
		}
	})
}

func optionalShiftNote(note *string) *string {
	if note == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func formatShiftTime(value *time.Time, location *time.Location) string {
	if value == nil {
		return ""
	}
	return value.In(location).Format(time.RFC3339)
}

func formatShiftInt(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
}

func respondShiftReadError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, ErrShiftNotFound), errors.Is(err, ErrNotFound), errors.Is(err, ErrOutletNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Shift atau outlet tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
	}
}
