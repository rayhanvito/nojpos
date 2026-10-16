package handler

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/repo"
	"github.com/rayhanvito/nojpos/backend/service"
)

type ShiftHandler struct {
	svc *service.ShiftService
}

func NewShiftHandler(svc *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{svc: svc}
}

type startShiftReq struct {
	OpeningCash int64 `json:"opening_cash"`
}

type closeShiftReq struct {
	ClosingCash *int64  `json:"closingCash"`
	Note        *string `json:"note"`
}

func (h *ShiftHandler) GetCurrentShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	data, err := h.svc.GetCashierShift(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveShift) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tidak ada shift yang berjalan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat shift."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *ShiftHandler) StartShift(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req startShiftReq
	_ = c.ShouldBindJSON(&req)

	data, err := h.svc.StartShift(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name, req.OpeningCash)
	if err != nil {
		if errors.Is(err, service.ErrShiftAlreadyOpen) {
			c.JSON(http.StatusConflict, gin.H{"error": "Shift masih aktif. Tutup shift yang berjalan dulu sebelum membuka yang baru."})
			return
		}
		if errors.Is(err, service.ErrCashierOutletChanged) {
			c.JSON(http.StatusConflict, gin.H{"error": "Outlet kasir sudah berubah. Masuk kembali ke akun kasir untuk membuka shift."})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, data)
}

func (h *ShiftHandler) CloseShift(c *gin.Context) {
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
	if errors.Is(err, service.ErrNoActiveShift) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada shift yang berjalan."})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal menutup shift."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *ShiftHandler) ListShifts(c *gin.Context) {
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
	// Keep the existing key while following the shared list response contract.
	c.JSON(http.StatusOK, gin.H{"items": items, "shifts": items})
}

func (h *ShiftHandler) GetShift(c *gin.Context) {
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

func (h *ShiftHandler) ForceCloseShift(c *gin.Context) {
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
		case errors.Is(err, service.ErrShiftNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Shift tidak ditemukan di toko Anda."})
		case errors.Is(err, service.ErrShiftAlreadyClosed):
			c.JSON(http.StatusConflict, gin.H{"error": "Shift sudah ditutup."})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menutup shift secara paksa."})
		}
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *ShiftHandler) ExportShifts(c *gin.Context) {
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
	methods := service.ShiftPaymentMethods()
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
	case errors.Is(err, service.ErrShiftNotFound), errors.Is(err, repo.ErrNotFound), errors.Is(err, repo.ErrOutletNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Shift atau outlet tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
	}
}
