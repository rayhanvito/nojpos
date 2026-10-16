package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/service"
)

type ExportHandler struct {
	svc *service.ExportService
}

func NewExportHandler(svc *service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) Transactions(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	transactions, timezone, err := h.svc.Transactions(c.Request.Context(), claims.StoreID, c.Query("period"), c.Query("date"), c.Query("from"), c.Query("to"))
	if err != nil {
		if errors.Is(err, service.ErrBadPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Periode tidak valid."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengekspor transaksi."})
		}
		return
	}
	loc := exportLocation(timezone)
	filename := "transactions-" + time.Now().In(loc).Format("2006-01-02") + ".csv"
	writeCSV(c, filename, []string{"id", "tanggal", "kasir", "pelanggan", "metode", "status", "subtotal", "diskon", "pajak", "total", "dibayar", "kembalian", "jumlah item"}, func(w *csv.Writer) {
		for _, trx := range transactions {
			var itemCount int
			for _, item := range trx.Items {
				itemCount += item.Qty
			}
			if err := w.Write([]string{
				strconv.FormatUint(uint64(trx.ID), 10),
				trx.CreatedAt.In(loc).Format(time.RFC3339),
				safeCSVText(trx.CashierName),
				safeCSVText(trx.Customer),
				safeCSVText(trx.Method),
				safeCSVText(string(trx.Status)),
				strconv.FormatInt(trx.Subtotal, 10),
				strconv.FormatInt(trx.Discount, 10),
				strconv.FormatInt(trx.Tax, 10),
				strconv.FormatInt(trx.Total, 10),
				strconv.FormatInt(trx.Paid, 10),
				strconv.FormatInt(trx.Change, 10),
				strconv.Itoa(itemCount),
			}); err != nil {
				return
			}
		}
	})
}

func (h *ExportHandler) Products(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	products, timezone, err := h.svc.Products(c.Request.Context(), claims.StoreID, claims.OutletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengekspor produk."})
		return
	}
	loc := exportLocation(timezone)
	filename := "products-" + time.Now().In(loc).Format("2006-01-02") + ".csv"
	writeCSV(c, filename, []string{"id", "nama", "sku", "barcode", "kategori", "harga beli", "harga jual", "stok", "satuan", "aktif"}, func(w *csv.Writer) {
		for _, product := range products {
			category := ""
			if product.CategoryName != nil {
				category = *product.CategoryName
			}
			if err := w.Write([]string{
				strconv.FormatUint(uint64(product.ID), 10),
				safeCSVText(product.Name),
				safeCSVText(product.SKU),
				safeCSVText(product.Barcode),
				safeCSVText(category),
				strconv.FormatInt(product.BuyPrice, 10),
				strconv.FormatInt(product.SellPrice, 10),
				strconv.Itoa(product.Stock),
				safeCSVText(product.Unit),
				strconv.FormatBool(product.Active),
			}); err != nil {
				return
			}
		}
	})
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
