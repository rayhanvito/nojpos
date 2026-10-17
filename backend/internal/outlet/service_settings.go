package outlet

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SettingsStoreRepository adalah kontrak penyimpanan pengaturan
// toko yang dipakai SettingsService — hanya GetSettings dan
// UpdateSettings (tipe StoreSettings paket ini).
type SettingsStoreRepository interface {
	GetSettings(ctx context.Context, storeID uint) (*StoreSettings, error)
	UpdateSettings(ctx context.Context, storeID uint, s *StoreSettings) (*StoreSettings, error)
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
	if in.Hours == nil {
		in.Hours = []StoreHours{}
	}
	if len(in.Hours) != 0 && len(in.Hours) != 3 {
		return ErrBadHours
	}
	for i, h := range in.Hours {
		if h.Days != validHoursDays[i] {
			return ErrBadHours
		}
		if h.Open == nil && h.Close == nil {
			continue
		}
		if h.Open == nil || h.Close == nil {
			return ErrBadHours
		}
		open, close := strings.TrimSpace(*h.Open), strings.TrimSpace(*h.Close)
		if !validHourMinute(open) || !validHourMinute(close) || open >= close {
			return ErrBadHours
		}
		in.Hours[i].Open, in.Hours[i].Close = &open, &close
	}
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
