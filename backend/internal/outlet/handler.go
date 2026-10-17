package outlet

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// storeIDFromClaims membaca StoreID dari claims yang disimpan middleware
// Auth pada konteks gin (kunci "claims", *service.Claims). Paket service
// mengimpor paket ini (interfaces.go -> Repository), jadi mengimpor
// middleware/service dari sini akan membentuk siklus impor; field dibaca
// lewat refleksi — pola yang sama dengan internal/promo. Untuk request
// terautentikasi hasilnya identik dengan middleware.ClaimsFrom(c).StoreID.
func storeIDFromClaims(c *gin.Context) uint {
	v, exists := c.Get("claims")
	if !exists || v == nil {
		return 0
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return 0
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return 0
	}
	f := rv.FieldByName("StoreID")
	if !f.IsValid() || (f.Kind() != reflect.Uint && f.Kind() != reflect.Uint64) {
		return 0
	}
	return uint(f.Uint())
}

// pathUint parses a numeric path param, writing 400 on garbage
// (semantik disalin persis dari handler/params.go lama).
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), storeIDFromClaims(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat daftar outlet."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type outletReq struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsDefault *bool  `json:"isDefault"`
}

func (h *Handler) Create(c *gin.Context) {
	var req outletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	outlet, err := h.svc.Create(c.Request.Context(), storeIDFromClaims(c), Outlet{
		Name: req.Name, Address: req.Address, IsDefault: req.IsDefault != nil && *req.IsDefault,
	})
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, outlet)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req outletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	outlet, err := h.svc.Update(c.Request.Context(), storeIDFromClaims(c), id, Outlet{
		Name: req.Name, Address: req.Address,
	}, req.IsDefault)
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusOK, outlet)
}

func respondOutletErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrOutletForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Kasir hanya dapat menggunakan outlet yang ditetapkan."})
	case errors.Is(err, ErrOutletNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
	case errors.Is(err, ErrOutletDefault):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Outlet default tidak dapat dinonaktifkan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
