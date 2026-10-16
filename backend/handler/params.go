package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// pathUint parses a numeric path param, writing 400 on garbage.
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

// queryID parses an optional numeric query filter.
// Returns 0 when absent. When present but invalid, returns max uint so the
// filter matches nothing (same visible result as before with bad input).
func queryID(q func(string) string, name string) uint {
	v := q(name)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return ^uint(0)
	}
	return uint(n)
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
