package outlet

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SettingsStoreRepository adalah kontrak penyimpanan pengaturan
// toko yang dipakai SettingsService — GetSettings dan UpdateSettings
// (tipe StoreSettings paket ini), plus akses override pengaturan per
// outlet sejak Unit C.
type SettingsStoreRepository interface {
	GetSettings(ctx context.Context, storeID uint) (*StoreSettings, error)
	UpdateSettings(ctx context.Context, storeID uint, s *StoreSettings) (*StoreSettings, error)
	GetOutlet(ctx context.Context, storeID, outletID uint) (*Outlet, error)
	GetOutletSettings(ctx context.Context, storeID, outletID uint) (*OutletSettings, error)
	UpsertOutletSettings(ctx context.Context, row *OutletSettings) (*OutletSettings, error)
}

// SettingsService adalah SettingsService lama di service/settings.go
// untuk Get/Update pengaturan toko (BE-F1d; SetPasscode pindah ke
// internal/user).
type SettingsService struct {
	stores SettingsStoreRepository
}

func NewSettingsService(stores SettingsStoreRepository) *SettingsService {
	return &SettingsService{stores: stores}
}

func (s *SettingsService) Get(ctx context.Context, storeID uint) (*StoreSettings, error) {
	return s.stores.GetSettings(ctx, storeID)
}

func (s *SettingsService) Update(ctx context.Context, storeID uint, in *StoreSettings) (*StoreSettings, error) {
	if in.LowStockDefault < 0 {
		return nil, ErrBadLowStockDefault
	}
	if in.LoyaltyEarnPer <= 0 {
		return nil, ErrBadLoyaltyEarnPer
	}
	if in.LoyaltyPointValue < 0 {
		return nil, ErrBadLoyaltyPointValue
	}
	if in.CreditDueDays < 0 {
		return nil, ErrBadCreditDueDays
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, fmt.Errorf("nama toko wajib diisi")
	}
	in.Timezone = strings.TrimSpace(in.Timezone)
	if in.Timezone == "" {
		in.Timezone = "Asia/Makassar"
	}
	if _, err := time.LoadLocation(in.Timezone); err != nil {
		return nil, ErrBadTimezone
	}
	if in.Paper != "58mm" && in.Paper != "80mm" {
		in.Paper = "58mm"
	}
	if in.TaxPct < 0 {
		in.TaxPct = 0
	}
	if in.TaxPct > 100 {
		return nil, ErrBadTaxPct
	}
	if err := validateExtendedSettings(in); err != nil {
		return nil, err
	}
	return s.stores.UpdateSettings(ctx, storeID, in)
}

var validBusinessTypes = map[string]bool{
	"": true, "retail": true, "fnb": true, "fashion": true, "jasa": true, "lainnya": true,
}

// Label days jam operasional (grup tampilan frontend) — tepat 3 entri.
var validHoursDays = []string{"Senin – Jumat", "Sabtu", "Minggu"}

func validateExtendedSettings(in *StoreSettings) error {
	in.BusinessType = strings.TrimSpace(in.BusinessType)
	if !validBusinessTypes[in.BusinessType] {
		return ErrBadBusinessType
	}
	in.Email = strings.TrimSpace(in.Email)
	if in.Email != "" && !isEmail(in.Email) {
		return ErrBadStoreEmail
	}
	in.City = strings.TrimSpace(in.City)
	in.Province = strings.TrimSpace(in.Province)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	if in.Currency == "" {
		in.Currency = "IDR"
	}
	if len(in.Currency) != 3 {
		return ErrBadCurrency
	}
	for _, ch := range in.Currency {
		if ch < 'A' || ch > 'Z' {
			return ErrBadCurrency
		}
	}
	hours, err := validateHours(in.Hours)
	if err != nil {
		return err
	}
	in.Hours = hours
	if len(in.ReceiptFooter) > 200 {
		return ErrBadReceiptFooter
	}
	if len(in.TaxName) > 20 {
		return ErrBadTaxName
	}
	in.TaxName = strings.TrimSpace(in.TaxName)
	switch in.TaxRounding {
	case "", "none":
		in.TaxRounding = "none"
	case "down", "up":
	default:
		return ErrBadTaxRounding
	}
	if in.TaxApplyTo == "" {
		in.TaxApplyTo = "all"
	}
	if in.TaxApplyTo != "all" {
		return ErrBadTaxApplyTo
	}
	return nil
}

func validHourMinute(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	for _, idx := range []int{0, 1, 3, 4} {
		if s[idx] < '0' || s[idx] > '9' {
			return false
		}
	}
	hh := int(s[0]-'0')*10 + int(s[1]-'0')
	mm := int(s[3]-'0')*10 + int(s[4]-'0')
	return hh < 24 && mm < 60
}

// isEmail adalah salinan privat slice dari helper di internal/auth
// (pemilik aslinya pindah di BE-F1c; validateExtendedSettings di sini
// memakainya untuk validasi email toko — semantik identik).
func isEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1 && strings.Contains(s[at:], ".") && !strings.ContainsAny(s, " \t")
}

// validateHours memvalidasi dan menormalisasi jam operasional — aturan
// yang sama persis untuk pengaturan toko (validateExtendedSettings) dan
// override per outlet: 0 entri (tidak diatur) atau tepat 3 entri
// berlabel validHoursDays; tiap entri open & close keduanya kosong atau
// keduanya HH:MM valid dengan open < close (open/close hasil trim
// ditulis kembali ke slice).
func validateHours(hours []StoreHours) ([]StoreHours, error) {
	if hours == nil {
		hours = []StoreHours{}
	}
	if len(hours) != 0 && len(hours) != 3 {
		return nil, ErrBadHours
	}
	for i, h := range hours {
		if h.Days != validHoursDays[i] {
			return nil, ErrBadHours
		}
		if h.Open == nil && h.Close == nil {
			continue
		}
		if h.Open == nil || h.Close == nil {
			return nil, ErrBadHours
		}
		open, close := strings.TrimSpace(*h.Open), strings.TrimSpace(*h.Close)
		if !validHourMinute(open) || !validHourMinute(close) || open >= close {
			return nil, ErrBadHours
		}
		hours[i].Open, hours[i].Close = &open, &close
	}
	return hours, nil
}

// ── Pengaturan per outlet (Unit C) ──────────────────────────────────

// OutletSettingsPatch adalah perubahan parsial override outlet dari
// PUT /outlets/:id/settings. Flag *Set menandai kunci hadir di body:
// hadir dengan nilai = tetapkan override; hadir dengan null (pointer
// nilai nil) = hapus override (kembali mewarisi); tidak hadir = tidak
// berubah.
type OutletSettingsPatch struct {
	TaxPctSet        bool
	TaxPct           *float64
	ReceiptFooterSet bool
	ReceiptFooter    *string
	HoursSet         bool
	Hours            *[]StoreHours
}

// EffectiveSettings menyelesaikan pengaturan efektif sebuah outlet:
// pengaturan toko dengan override outlet diterapkan di atasnya untuk
// ketiga field yang dapat dioverride (taxPct, receiptFooter, hours);
// field lain selalu nilai toko. Hasil kedua adalah baris override
// mentah (nil bila outlet belum punya override) supaya pemanggil tahu
// field mana yang berasal dari outlet. Outlet yang tidak ada di toko
// mengembalikan ErrNotFound.
func (s *SettingsService) EffectiveSettings(ctx context.Context, storeID, outletID uint) (*StoreSettings, *OutletSettings, error) {
	if _, err := s.stores.GetOutlet(ctx, storeID, outletID); err != nil {
		return nil, nil, err
	}
	eff, err := s.stores.GetSettings(ctx, storeID)
	if err != nil {
		return nil, nil, err
	}
	ov, err := s.stores.GetOutletSettings(ctx, storeID, outletID)
	if err != nil {
		return nil, nil, err
	}
	applyOutletOverrides(eff, ov)
	return eff, ov, nil
}

// applyOutletOverrides menimpa field yang dioverride outlet pada
// pengaturan toko eff (diubah di tempat).
func applyOutletOverrides(eff *StoreSettings, ov *OutletSettings) {
	if ov == nil {
		return
	}
	if ov.TaxPct != nil {
		eff.TaxPct = *ov.TaxPct
	}
	if ov.ReceiptFooter != nil {
		eff.ReceiptFooter = *ov.ReceiptFooter
	}
	if ov.Hours != nil {
		eff.Hours = decodeStoreHours(*ov.Hours)
	}
}

// UpdateOutletSettings menerapkan patch override outlet dengan validasi
// yang meniru persis SettingsService toko (pajak: negatif dibulatkan ke
// 0, di atas 100 ditolak ErrBadTaxPct; footer maksimal 200 karakter;
// jam operasional aturan validateHours), menyimpan baris override
// (upsert), lalu mengembalikan pengaturan efektif terbaru beserta baris
// override-nya. Outlet yang tidak ada di toko mengembalikan
// ErrNotFound.
func (s *SettingsService) UpdateOutletSettings(ctx context.Context, storeID, outletID uint, patch OutletSettingsPatch) (*StoreSettings, *OutletSettings, error) {
	if _, err := s.stores.GetOutlet(ctx, storeID, outletID); err != nil {
		return nil, nil, err
	}
	row, err := s.stores.GetOutletSettings(ctx, storeID, outletID)
	if err != nil {
		return nil, nil, err
	}
	if row == nil {
		row = &OutletSettings{StoreID: storeID, OutletID: outletID}
	}
	if patch.TaxPctSet {
		if patch.TaxPct == nil {
			row.TaxPct = nil
		} else {
			pct := *patch.TaxPct
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				return nil, nil, ErrBadTaxPct
			}
			row.TaxPct = &pct
		}
	}
	if patch.ReceiptFooterSet {
		if patch.ReceiptFooter == nil {
			row.ReceiptFooter = nil
		} else {
			if len(*patch.ReceiptFooter) > 200 {
				return nil, nil, ErrBadReceiptFooter
			}
			footer := *patch.ReceiptFooter
			row.ReceiptFooter = &footer
		}
	}
	if patch.HoursSet {
		if patch.Hours == nil {
			row.Hours = nil
		} else {
			normalized, err := validateHours(*patch.Hours)
			if err != nil {
				return nil, nil, err
			}
			encoded := encodeStoreHours(normalized)
			row.Hours = &encoded
		}
	}
	saved, err := s.stores.UpsertOutletSettings(ctx, row)
	if err != nil {
		return nil, nil, err
	}
	eff, err := s.stores.GetSettings(ctx, storeID)
	if err != nil {
		return nil, nil, err
	}
	applyOutletOverrides(eff, saved)
	return eff, saved, nil
}
