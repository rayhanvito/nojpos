package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Konstanta aksi aktivitas yang ditulis handler ini (nilai verbatim
// dari service/activity.go).
const (
	ActivityUserCreated     = "USER_CREATED"
	ActivityProfileUpdated  = "PROFILE_UPDATED"
	ActivityAccountDisabled = "ACCOUNT_DISABLED"
	ActivityAccountEnabled  = "ACCOUNT_ENABLED"
	ActivityPasscodeChanged = "PASSCODE_CHANGED"
)

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

// Handler adalah UserHandler lama di handler/user.go (BE-F1d), plus
// SetPasscode dari handler/settings.go — badan metode verbatim.
func formatID(id uint) string {
	return fmt.Sprintf("%d", id)
}

type Handler struct {
	svc      *Service
	activity ActivityLogger
}

func NewHandler(svc *Service, activity ...ActivityLogger) *Handler {
	h := &Handler{svc: svc}
	if len(activity) > 0 {
		h.activity = activity[0]
	}
	return h
}

func (h *Handler) logActivity(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string) {
	if h.activity == nil {
		return
	}
	if len(detail) > 80 {
		detail = detail[:80]
	}
	h.activity.Log(ctx, storeID, actorID, actorName, action, detail, refType, refID)
}

type createUserReq struct {
	Name     string `json:"name"`
	OutletID *uint  `json:"outletId"`
}

type setActiveReq struct {
	Active bool `json:"active"`
}

type renameUserReq struct {
	Name     string `json:"name"`
	OutletID *uint  `json:"outletId"`
}

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	users, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat daftar akun."})
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, users))
}

func (h *Handler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	user, err := h.svc.CreateCashier(c.Request.Context(), claims.StoreID, claims.OutletID, req.Name, req.OutletID)
	if err != nil {
		respondUserErr(c, err)
		return
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, ActivityUserCreated, fmt.Sprintf("%s ditambahkan sebagai kasir", user.Name), "user", formatID(user.ID))
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (h *Handler) SetActive(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req setActiveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	// Resolve target name for activity detail (best-effort).
	targetName := fmt.Sprintf("Akun %d", id)
	if users, err := h.svc.List(c.Request.Context(), claims.StoreID); err == nil {
		for _, u := range users {
			if u.ID == id {
				targetName = u.Name
				break
			}
		}
	}
	if err := h.svc.SetActive(c.Request.Context(), claims.StoreID, id, req.Active); err != nil {
		respondUserErr(c, err)
		return
	}
	msg := "Akun diaktifkan."
	action := ActivityAccountEnabled
	detail := fmt.Sprintf("%s diaktifkan kembali", targetName)
	if !req.Active {
		msg = "Akun dinonaktifkan."
		action = ActivityAccountDisabled
		detail = fmt.Sprintf("%s dinonaktifkan", targetName)
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, action, detail, "user", formatID(id))
	c.JSON(http.StatusOK, gin.H{"message": msg})
}

func (h *Handler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondUserErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Akun kasir berhasil dihapus."})
}

func (h *Handler) Rename(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req renameUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	if err := h.svc.UpdateCashier(c.Request.Context(), claims.StoreID, id, req.Name, req.OutletID); err != nil {
		respondUserErr(c, err)
		return
	}
	detail := "Penempatan outlet kasir diperbarui"
	message := "Outlet kasir diperbarui."
	if strings.TrimSpace(req.Name) != "" {
		detail = fmt.Sprintf("%s memperbarui profil", req.Name)
		message = "Nama kasir diperbarui."
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, ActivityProfileUpdated, detail, "user", formatID(id))
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func respondUserErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar."})
	case errors.Is(err, ErrStoreMismatch):
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan di toko Anda."})
	case errors.Is(err, ErrNotEditable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya akun kasir yang dapat dinonaktifkan."})
	case errors.Is(err, ErrNotRenamable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya akun kasir yang dapat diubah."})
	case errors.Is(err, ErrCashierOpenShift):
		c.JSON(http.StatusConflict, gin.H{"error": "Kasir tidak dapat dipindahkan ke outlet lain selama shift masih terbuka. Tutup shift terlebih dahulu."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

type passcodeReq struct {
	Passcode string `json:"passcode"`
	// Role disambiguates owner vs cashier accounts sharing one number
	// ("admin"/"cashier", optional — legacy cashier-first order when empty).
	Role string `json:"role"`
}

// SetPasscode adalah SettingsHandler.SetPasscode lama (handler/settings.go).
func (h *Handler) SetPasscode(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req passcodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	if err := h.svc.SetPasscode(c.Request.Context(), claims.StoreID, id, req.Passcode, req.Role); err != nil {
		switch {
		case errors.Is(err, ErrStoreMismatch), errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Akun tidak ditemukan di toko Anda."})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	msg := "Passcode disimpan."
	if req.Passcode == "" {
		msg = "Passcode dihapus."
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, ActivityPasscodeChanged, fmt.Sprintf("Passcode akun %d diperbarui", id), "user", formatID(id))
	c.JSON(http.StatusOK, gin.H{"message": msg})
}
