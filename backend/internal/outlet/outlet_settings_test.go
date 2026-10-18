package outlet

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Tes Unit C: resolusi pengaturan efektif per outlet, patch override
// (termasuk null = hapus override), validasi, dan 404 outlet asing —
// di level SettingsService (fake repo) dan SettingsHandler.

func floatPtr(f float64) *float64 { return &f }

func storeSettingsForOutletTests() *StoreSettings {
	return &StoreSettings{
		Name:          "Toko Uji",
		Timezone:      "Asia/Jakarta",
		TaxEnabled:    true,
		TaxPct:        11,
		ReceiptFooter: "Footer toko",
		Hours: []StoreHours{
			{Days: "Senin – Jumat", Open: strPtr("08:00"), Close: strPtr("17:00")},
			{Days: "Sabtu", Open: strPtr("08:00"), Close: strPtr("12:00")},
			{Days: "Minggu"},
		},
		LoyaltyEarnPer:    10000,
		LoyaltyPointValue: 100,
		CreditDueDays:     14,
	}
}

func TestEffectiveSettingsFullInheritance(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests()}
	svc := NewSettingsService(repo)

	eff, ov, err := svc.EffectiveSettings(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("EffectiveSettings: %v", err)
	}
	if ov != nil {
		t.Fatalf("override = %+v, want nil (belum ada baris override)", ov)
	}
	if eff.TaxPct != 11 || eff.ReceiptFooter != "Footer toko" || len(eff.Hours) != 3 {
		t.Fatalf("efektif tidak mewarisi penuh dari toko: %+v", eff)
	}
}

func TestEffectiveSettingsPartialOverride(t *testing.T) {
	repo := &fakeSettingsRepo{
		getOut: storeSettingsForOutletTests(),
		ovOut:  &OutletSettings{StoreID: 7, OutletID: 9, TaxPct: floatPtr(5)},
	}
	svc := NewSettingsService(repo)

	eff, ov, err := svc.EffectiveSettings(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("EffectiveSettings: %v", err)
	}
	if ov == nil || ov.TaxPct == nil || *ov.TaxPct != 5 {
		t.Fatalf("override dikembalikan salah: %+v", ov)
	}
	if eff.TaxPct != 5 {
		t.Fatalf("TaxPct efektif = %v, want 5 (override)", eff.TaxPct)
	}
	if eff.ReceiptFooter != "Footer toko" || len(eff.Hours) != 3 || *eff.Hours[0].Open != "08:00" {
		t.Fatalf("field tanpa override harus mewarisi toko: %+v", eff)
	}
}

func TestEffectiveSettingsHoursOverride(t *testing.T) {
	custom := []StoreHours{
		{Days: "Senin – Jumat", Open: strPtr("09:00"), Close: strPtr("21:00")},
		{Days: "Sabtu"},
		{Days: "Minggu"},
	}
	encoded := encodeStoreHours(custom)
	repo := &fakeSettingsRepo{
		getOut: storeSettingsForOutletTests(),
		ovOut:  &OutletSettings{StoreID: 7, OutletID: 9, Hours: &encoded},
	}
	svc := NewSettingsService(repo)

	eff, _, err := svc.EffectiveSettings(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("EffectiveSettings: %v", err)
	}
	if len(eff.Hours) != 3 || eff.Hours[0].Open == nil || *eff.Hours[0].Open != "09:00" {
		t.Fatalf("Hours efektif tidak memakai override: %+v", eff.Hours)
	}
	if eff.TaxPct != 11 || eff.ReceiptFooter != "Footer toko" {
		t.Fatalf("field lain harus tetap mewarisi: %+v", eff)
	}
}

func TestEffectiveSettingsForeignOutletNotFound(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests(), outletErr: ErrNotFound}
	svc := NewSettingsService(repo)

	_, _, err := svc.EffectiveSettings(context.Background(), 7, 9)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdateOutletSettingsSetOverride(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests()}
	svc := NewSettingsService(repo)

	eff, saved, err := svc.UpdateOutletSettings(context.Background(), 7, 9, OutletSettingsPatch{
		TaxPctSet: true, TaxPct: floatPtr(7),
	})
	if err != nil {
		t.Fatalf("UpdateOutletSettings: %v", err)
	}
	if repo.upsertCalls != 1 || saved.TaxPct == nil || *saved.TaxPct != 7 {
		t.Fatalf("baris tersimpan salah: %+v (upsert %d kali)", saved, repo.upsertCalls)
	}
	if eff.TaxPct != 7 || eff.ReceiptFooter != "Footer toko" {
		t.Fatalf("efektif sesudah update salah: %+v", eff)
	}
}

func TestUpdateOutletSettingsNullClearsOverride(t *testing.T) {
	repo := &fakeSettingsRepo{
		getOut: storeSettingsForOutletTests(),
		ovOut:  &OutletSettings{StoreID: 7, OutletID: 9, TaxPct: floatPtr(7), ReceiptFooter: strPtr("Footer outlet")},
	}
	svc := NewSettingsService(repo)

	// taxPct: null menghapus override pajak saja; footer outlet utuh.
	eff, saved, err := svc.UpdateOutletSettings(context.Background(), 7, 9, OutletSettingsPatch{TaxPctSet: true})
	if err != nil {
		t.Fatalf("UpdateOutletSettings: %v", err)
	}
	if saved.TaxPct != nil {
		t.Fatalf("TaxPct override = %v, want nil sesudah null", *saved.TaxPct)
	}
	if saved.ReceiptFooter == nil || *saved.ReceiptFooter != "Footer outlet" {
		t.Fatalf("footer override tidak boleh berubah: %+v", saved.ReceiptFooter)
	}
	if eff.TaxPct != 11 {
		t.Fatalf("TaxPct efektif = %v, want 11 (kembali mewarisi)", eff.TaxPct)
	}
	if eff.ReceiptFooter != "Footer outlet" {
		t.Fatalf("footer efektif = %q, want override outlet", eff.ReceiptFooter)
	}
}

func TestUpdateOutletSettingsValidation(t *testing.T) {
	hoursOneEntry := []StoreHours{{Days: "Senin – Jumat", Open: strPtr("08:00"), Close: strPtr("17:00")}}
	hoursOpenAfterClose := []StoreHours{
		{Days: "Senin – Jumat", Open: strPtr("17:00"), Close: strPtr("08:00")},
		{Days: "Sabtu"},
		{Days: "Minggu"},
	}
	cases := []struct {
		name  string
		patch OutletSettingsPatch
		want  error
	}{
		{"pajak di atas 100", OutletSettingsPatch{TaxPctSet: true, TaxPct: floatPtr(101)}, ErrBadTaxPct},
		{"footer 201 karakter", OutletSettingsPatch{ReceiptFooterSet: true, ReceiptFooter: strPtr(strings.Repeat("a", 201))}, ErrBadReceiptFooter},
		{"jam hanya 1 entri", OutletSettingsPatch{HoursSet: true, Hours: &hoursOneEntry}, ErrBadHours},
		{"jam buka sesudah tutup", OutletSettingsPatch{HoursSet: true, Hours: &hoursOpenAfterClose}, ErrBadHours},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests()}
			svc := NewSettingsService(repo)
			_, _, err := svc.UpdateOutletSettings(context.Background(), 7, 9, tc.patch)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if repo.upsertCalls != 0 {
				t.Fatalf("upsert tidak boleh jalan untuk input tidak valid")
			}
		})
	}
}

func TestUpdateOutletSettingsNegativeTaxClampedLikeStore(t *testing.T) {
	// Meniru persis SettingsService toko: tarif negatif dibulatkan ke 0,
	// bukan ditolak.
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests()}
	svc := NewSettingsService(repo)

	_, saved, err := svc.UpdateOutletSettings(context.Background(), 7, 9, OutletSettingsPatch{
		TaxPctSet: true, TaxPct: floatPtr(-5),
	})
	if err != nil {
		t.Fatalf("UpdateOutletSettings: %v", err)
	}
	if saved.TaxPct == nil || *saved.TaxPct != 0 {
		t.Fatalf("TaxPct tersimpan = %v, want 0 (clamp seperti toko)", saved.TaxPct)
	}
}

func TestUpdateOutletSettingsForeignOutletNotFound(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests(), outletErr: ErrNotFound}
	svc := NewSettingsService(repo)

	_, _, err := svc.UpdateOutletSettings(context.Background(), 7, 9, OutletSettingsPatch{TaxPctSet: true, TaxPct: floatPtr(5)})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
	if repo.upsertCalls != 0 {
		t.Fatalf("upsert tidak boleh jalan untuk outlet asing")
	}
}

// ── Handler ─────────────────────────────────────────────────────────

func newOutletSettingsHandlerCtx(t *testing.T, method, outletID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, "/outlets/"+outletID+"/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, "/outlets/"+outletID+"/settings", nil)
	}
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: outletID}}
	ctx.Set("claims", &middleware.Claims{StoreID: 7, UserID: 3, Name: "Admin Uji"})
	return ctx, recorder
}

func TestGetOutletSettingsHandlerForeignOutlet404(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests(), outletErr: ErrNotFound}
	h := NewSettingsHandler(NewSettingsService(repo))

	ctx, recorder := newOutletSettingsHandlerCtx(t, http.MethodGet, "9", "")
	h.GetOutletSettings(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Outlet tidak ditemukan di toko Anda.") {
		t.Fatalf("body 404 tidak sesuai idiom outlet: %s", recorder.Body.String())
	}
}

func TestUpdateOutletSettingsHandlerForeignOutlet404(t *testing.T) {
	repo := &fakeSettingsRepo{getOut: storeSettingsForOutletTests(), outletErr: ErrNotFound}
	h := NewSettingsHandler(NewSettingsService(repo))

	ctx, recorder := newOutletSettingsHandlerCtx(t, http.MethodPut, "9", `{"taxPct": 5}`)
	h.UpdateOutletSettings(ctx)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetOutletSettingsHandlerShape(t *testing.T) {
	repo := &fakeSettingsRepo{
		getOut: storeSettingsForOutletTests(),
		ovOut:  &OutletSettings{StoreID: 7, OutletID: 9, TaxPct: floatPtr(5)},
	}
	h := NewSettingsHandler(NewSettingsService(repo))

	ctx, recorder := newOutletSettingsHandlerCtx(t, http.MethodGet, "9", "")
	h.GetOutletSettings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode respons: %v", err)
	}
	if body["taxPct"] != float64(5) {
		t.Fatalf("taxPct efektif = %v, want 5 (override)", body["taxPct"])
	}
	if body["receiptFooter"] != "Footer toko" {
		t.Fatalf("receiptFooter efektif = %v, want warisan toko", body["receiptFooter"])
	}
	overrides, ok := body["overrides"].(map[string]any)
	if !ok {
		t.Fatalf("overrides hilang dari respons: %s", recorder.Body.String())
	}
	if len(overrides) != 1 || overrides["taxPct"] != float64(5) {
		t.Fatalf("overrides = %v, want hanya {taxPct: 5}", overrides)
	}
}

func TestUpdateOutletSettingsHandlerNullClears(t *testing.T) {
	repo := &fakeSettingsRepo{
		getOut: storeSettingsForOutletTests(),
		ovOut:  &OutletSettings{StoreID: 7, OutletID: 9, TaxPct: floatPtr(5)},
	}
	h := NewSettingsHandler(NewSettingsService(repo))

	ctx, recorder := newOutletSettingsHandlerCtx(t, http.MethodPut, "9", `{"taxPct": null}`)
	h.UpdateOutletSettings(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", recorder.Code, recorder.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode respons: %v", err)
	}
	if body["taxPct"] != float64(11) {
		t.Fatalf("taxPct efektif = %v, want 11 sesudah override dihapus", body["taxPct"])
	}
	overrides, _ := body["overrides"].(map[string]any)
	if len(overrides) != 0 {
		t.Fatalf("overrides = %v, want kosong sesudah null", overrides)
	}
}
