package outlet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// ActivityLogger adalah kontrak pencatat aktivitas untuk
// SettingsHandler (diimplementasikan activity.Service; paket
// ini tidak mengimpor paket itu).
type ActivityLogger interface {
	Log(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string)
}

// activityProfileUpdated adalah nilai aksi aktivitas untuk pembaruan
// profil/pengaturan (verbatim dari service.ActivityProfileUpdated).
const activityProfileUpdated = "PROFILE_UPDATED"

func formatID(id uint) string {
	return fmt.Sprintf("%d", id)
}

// SettingsHandler adalah SettingsHandler lama di handler/settings.go
// untuk Get/Update pengaturan (BE-F1d; SetPasscode pindah ke
// internal/user) — badan metode verbatim.
type SettingsHandler struct {
	svc      *SettingsService
	activity ActivityLogger
}

func NewSettingsHandler(svc *SettingsService, activity ...ActivityLogger) *SettingsHandler {
	h := &SettingsHandler{svc: svc}
	if len(activity) > 0 {
		h.activity = activity[0]
	}
	return h
}

func (h *SettingsHandler) logActivity(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string) {
	if h.activity == nil {
		return
	}
	if len(detail) > 80 {
		detail = detail[:80]
	}
	h.activity.Log(ctx, storeID, actorID, actorName, action, detail, refType, refID)
}

func (h *SettingsHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	s, err := h.svc.Get(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan."})
		return
	}
	c.JSON(http.StatusOK, s)
}

// outletSettingsResponse adalah bentuk respons GET/PUT
// /outlets/:id/settings (Unit C): seluruh field pengaturan EFEKTIF
// outlet — bentuk yang sama dengan respons GET /settings — ditambah
// objek "overrides" berisi hanya field yang ditetapkan outlet itu
// sendiri (kunci JSON yang sama; kosong {} bila belum ada override).
type outletSettingsResponse struct {
	*StoreSettings
	Overrides map[string]any `json:"overrides"`
}

func outletSettingsOverridesJSON(ov *OutletSettings) map[string]any {
	out := map[string]any{}
	if ov == nil {
		return out
	}
	if ov.TaxPct != nil {
		out["taxPct"] = *ov.TaxPct
	}
	if ov.ReceiptFooter != nil {
		out["receiptFooter"] = *ov.ReceiptFooter
	}
	if ov.Hours != nil {
		out["hours"] = decodeStoreHours(*ov.Hours)
	}
	return out
}

func respondOutletSettingsErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
	case errors.Is(err, ErrBadTaxPct), errors.Is(err, ErrBadReceiptFooter), errors.Is(err, ErrBadHours):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses pengaturan outlet."})
	}
}

// GetOutletSettings melayani GET /outlets/:id/settings (grup login):
// pengaturan efektif outlet — nilai toko yang ditimpa override outlet.
func (h *SettingsHandler) GetOutletSettings(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	eff, ov, err := h.svc.EffectiveSettings(c.Request.Context(), claims.StoreID, outletID)
	if err != nil {
		respondOutletSettingsErr(c, err)
		return
	}
	c.JSON(http.StatusOK, outletSettingsResponse{StoreSettings: eff, Overrides: outletSettingsOverridesJSON(ov)})
}

// parseOutletSettingsPatch memilah body PUT /outlets/:id/settings per
// kunci: kunci absen = tidak berubah; kunci null = hapus override;
// kunci bernilai = tetapkan override. Field tak dikenal diabaikan —
// idiom yang sama dengan PUT /settings.
func parseOutletSettingsPatch(c *gin.Context) (OutletSettingsPatch, bool) {
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return OutletSettingsPatch{}, false
	}
	bad := func() (OutletSettingsPatch, bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return OutletSettingsPatch{}, false
	}
	var patch OutletSettingsPatch
	if v, present := raw["taxPct"]; present {
		patch.TaxPctSet = true
		if string(v) != "null" {
			var f float64
			if err := json.Unmarshal(v, &f); err != nil {
				return bad()
			}
			patch.TaxPct = &f
		}
	}
	if v, present := raw["receiptFooter"]; present {
		patch.ReceiptFooterSet = true
		if string(v) != "null" {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				return bad()
			}
			patch.ReceiptFooter = &s
		}
	}
	if v, present := raw["hours"]; present {
		patch.HoursSet = true
		if string(v) != "null" {
			var hours []StoreHours
			if err := json.Unmarshal(v, &hours); err != nil {
				return bad()
			}
			patch.Hours = &hours
		}
	}
	return patch, true
}

// UpdateOutletSettings melayani PUT /outlets/:id/settings (grup
// admin): menerapkan subset override outlet dan mengembalikan
// pengaturan efektif terbaru dalam bentuk yang sama dengan GET.
func (h *SettingsHandler) UpdateOutletSettings(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	outletID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	patch, ok := parseOutletSettingsPatch(c)
	if !ok {
		return
	}
	eff, ov, err := h.svc.UpdateOutletSettings(c.Request.Context(), claims.StoreID, outletID, patch)
	if err != nil {
		respondOutletSettingsErr(c, err)
		return
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, activityProfileUpdated, fmt.Sprintf("%s memperbarui pengaturan outlet %d", claims.Name, outletID), "outlet", formatID(outletID))
	c.JSON(http.StatusOK, outletSettingsResponse{StoreSettings: eff, Overrides: outletSettingsOverridesJSON(ov)})
}

type settingsReq struct {
	Name                *string       `json:"storeName"`
	Address             *string       `json:"address"`
	Phone               *string       `json:"phone"`
	TaxEnabled          *bool         `json:"taxEnabled"`
	TaxPct              *float64      `json:"taxPct"`
	ReceiptHeader       *string       `json:"receiptHeader"`
	ReceiptFooter       *string       `json:"receiptFooter"`
	Paper               *string       `json:"paper"`
	Timezone            *string       `json:"timezone"`
	BusinessType        *string       `json:"businessType"`
	Email               *string       `json:"email"`
	City                *string       `json:"city"`
	Province            *string       `json:"province"`
	Currency            *string       `json:"currency"`
	Hours               *[]StoreHours `json:"hours"`
	ReceiptShowLogo     *bool         `json:"receiptShowLogo"`
	ReceiptShowCashier  *bool         `json:"receiptShowCashier"`
	ReceiptShowMethod   *bool         `json:"receiptShowMethod"`
	ReceiptShowTax      *bool         `json:"receiptShowTax"`
	ReceiptShowDiscount *bool         `json:"receiptShowDiscount"`
	ReceiptShowNote     *bool         `json:"receiptShowNote"`
	TaxName             *string       `json:"taxName"`
	TaxInclusive        *bool         `json:"taxInclusive"`
	TaxRounding         *string       `json:"taxRounding"`
	TaxApplyTo          *string       `json:"taxApplyTo"`
	LoyaltyEnabled      *bool         `json:"loyaltyEnabled"`
	LoyaltyEarnPer      *int64        `json:"loyaltyEarnPer"`
	LoyaltyPointValue   *int64        `json:"loyaltyPointValue"`
	CreditDueDays       *int          `json:"creditDueDays"`
	LowStockDefault     *int          `json:"lowStockDefault"`
}

func (h *SettingsHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req settingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	// Merge-semantics: key yang absen = tidak diubah (body parsial tetap valid,
	// field tak dikenal otomatis diabaikan oleh binding).
	cur, err := h.svc.Get(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan."})
		return
	}
	in := cur
	applyStr := func(dst *string, v *string) {
		if v != nil {
			*dst = *v
		}
	}
	applyBool := func(dst *bool, v *bool) {
		if v != nil {
			*dst = *v
		}
	}
	applyStr(&in.Name, req.Name)
	applyStr(&in.Address, req.Address)
	applyStr(&in.Phone, req.Phone)
	applyBool(&in.TaxEnabled, req.TaxEnabled)
	if req.TaxPct != nil {
		in.TaxPct = *req.TaxPct
	}
	applyStr(&in.ReceiptHeader, req.ReceiptHeader)
	applyStr(&in.ReceiptFooter, req.ReceiptFooter)
	applyStr(&in.Paper, req.Paper)
	applyStr(&in.Timezone, req.Timezone)
	applyStr(&in.BusinessType, req.BusinessType)
	applyStr(&in.Email, req.Email)
	applyStr(&in.City, req.City)
	applyStr(&in.Province, req.Province)
	applyStr(&in.Currency, req.Currency)
	if req.Hours != nil {
		in.Hours = *req.Hours
	}
	applyBool(&in.ReceiptShowLogo, req.ReceiptShowLogo)
	applyBool(&in.ReceiptShowCashier, req.ReceiptShowCashier)
	applyBool(&in.ReceiptShowMethod, req.ReceiptShowMethod)
	applyBool(&in.ReceiptShowTax, req.ReceiptShowTax)
	applyBool(&in.ReceiptShowDiscount, req.ReceiptShowDiscount)
	applyBool(&in.ReceiptShowNote, req.ReceiptShowNote)
	applyStr(&in.TaxName, req.TaxName)
	applyBool(&in.TaxInclusive, req.TaxInclusive)
	applyStr(&in.TaxRounding, req.TaxRounding)
	applyStr(&in.TaxApplyTo, req.TaxApplyTo)
	applyBool(&in.LoyaltyEnabled, req.LoyaltyEnabled)
	if req.LoyaltyEarnPer != nil {
		in.LoyaltyEarnPer = *req.LoyaltyEarnPer
	}
	if req.LoyaltyPointValue != nil {
		in.LoyaltyPointValue = *req.LoyaltyPointValue
	}
	if req.CreditDueDays != nil {
		in.CreditDueDays = *req.CreditDueDays
	}
	if req.LowStockDefault != nil {
		in.LowStockDefault = *req.LowStockDefault
	}
	s, err := h.svc.Update(c.Request.Context(), claims.StoreID, in)
	if err != nil {
		switch {
		case errors.Is(err, ErrBadTimezone):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Zona waktu tidak valid."})
		case errors.Is(err, ErrBadBusinessType),
			errors.Is(err, ErrBadStoreEmail),
			errors.Is(err, ErrBadCurrency),
			errors.Is(err, ErrBadHours),
			errors.Is(err, ErrBadReceiptFooter),
			errors.Is(err, ErrBadTaxName),
			errors.Is(err, ErrBadTaxPct),
			errors.Is(err, ErrBadTaxRounding),
			errors.Is(err, ErrBadTaxApplyTo),
			errors.Is(err, ErrBadLowStockDefault):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	h.logActivity(c.Request.Context(), claims.StoreID, formatID(claims.UserID), claims.Name, activityProfileUpdated, fmt.Sprintf("%s memperbarui profil", claims.Name), "", "")
	c.JSON(http.StatusOK, s)
}
