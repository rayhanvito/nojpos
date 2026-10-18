package outlet

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// mapDBErr memetakan error database ke sentinel slice, setara mapDBErr
// lama di repo/base.go: record kosong -> ErrNotFound; pelanggaran unik
// (Postgres 23505 / SQLite) -> ErrDuplicate.
func mapDBErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "23505") ||
			strings.Contains(msg, "duplicate key") ||
			strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "unique constraint") {
			return ErrDuplicate
		}
	}
	return err
}

// Repo adalah repositori outlet (OutletRepo lama di repo/outlet.go).
type Repo struct {
	database.BaseRepo
}

func NewOutletRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *Repo) List(ctx context.Context, storeID uint) ([]Outlet, error) {
	items := make([]Outlet, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).Order("is_default DESC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repo) GetByID(ctx context.Context, storeID, id uint) (*Outlet, error) {
	var outlet Outlet
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *Repo) GetDefault(ctx context.Context, storeID uint) (*Outlet, error) {
	var outlet Outlet
	if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *Repo) Create(ctx context.Context, storeID uint, outlet *Outlet) (*Outlet, error) {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var store Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var defaults int64
		if err := tx.Model(&Outlet{}).Where("store_id = ? AND is_default = ?", storeID, true).Count(&defaults).Error; err != nil {
			return err
		}
		if outlet.IsDefault || defaults == 0 {
			if err := tx.Model(&Outlet{}).Where("store_id = ?", storeID).Update("is_default", false).Error; err != nil {
				return err
			}
			outlet.IsDefault = true
		}
		outlet.StoreID = storeID
		if err := tx.Create(outlet).Error; err != nil {
			return err
		}
		if outlet.IsDefault {
			return mirrorOutletStocksToLegacy(tx, storeID, outlet.ID)
		}
		return nil
	})
	if err != nil {
		return nil, mapDBErr(err)
	}
	return r.GetByID(ctx, storeID, outlet.ID)
}

func (r *Repo) Update(ctx context.Context, storeID, id uint, updated Outlet, isDefault *bool) (*Outlet, error) {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var store Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var current Outlet
		if err := tx.Where("id = ? AND store_id = ?", id, storeID).First(&current).Error; err != nil {
			return mapDBErr(err)
		}
		if isDefault != nil && current.IsDefault && !*isDefault {
			return ErrOutletDefault
		}
		if isDefault != nil && *isDefault {
			if err := tx.Model(&Outlet{}).Where("store_id = ?", storeID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{"name": updated.Name, "address": updated.Address}
		if isDefault != nil {
			updates["is_default"] = *isDefault
		}
		result := tx.Model(&Outlet{}).Where("id = ? AND store_id = ?", id, storeID).Updates(updates)
		if result.Error != nil {
			return mapDBErr(result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		if isDefault != nil && *isDefault {
			return mirrorOutletStocksToLegacy(tx, storeID, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, id)
}

// mirrorOutletStocksToLegacy memakai nama tabel langsung, bukan tipe
// catalog.Product / catalog.ProductVariant: slice katalog mengimpor
// slice outlet (alias sentinel + tipe Outlet), jadi outlet tidak boleh
// mengimpor katalog (siklus impor). SQL yang dihasilkan identik —
// Model(&Product{}) pun hanya menunjuk tabel "products".
func mirrorOutletStocksToLegacy(tx *gorm.DB, storeID, outletID uint) error {
	if err := tx.Unscoped().Table("products").
		Where("store_id = ?", storeID).
		Update("stock", gorm.Expr("COALESCE((SELECT qty FROM product_stocks WHERE product_stocks.product_id = products.id AND product_stocks.outlet_id = ? AND product_stocks.variant_id IS NULL AND product_stocks.deleted_at IS NULL LIMIT 1), 0)", outletID)).Error; err != nil {
		return err
	}
	return tx.Unscoped().Table("product_variants").
		Where("store_id = ?", storeID).
		Update("stock", gorm.Expr("COALESCE((SELECT qty FROM product_stocks WHERE product_stocks.product_id = product_variants.product_id AND product_stocks.variant_id = product_variants.id AND product_stocks.outlet_id = ? AND product_stocks.deleted_at IS NULL LIMIT 1), 0)", outletID)).Error
}

func decodeStoreHours(raw string) []StoreHours {
	if raw == "" {
		return []StoreHours{}
	}
	var h []StoreHours
	if err := json.Unmarshal([]byte(raw), &h); err != nil || h == nil {
		return []StoreHours{}
	}
	return h
}

func encodeStoreHours(h []StoreHours) string {
	if h == nil {
		h = []StoreHours{}
	}
	buf, err := json.Marshal(h)
	if err != nil {
		return "[]"
	}
	return string(buf)
}

// StoreRepo adalah repositori pengaturan toko (StoreRepo lama di
// repo/store.go).
type StoreRepo struct {
	database.BaseRepo
}

func NewStoreRepo(db *gorm.DB) *StoreRepo {
	return &StoreRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *StoreRepo) GetSettings(ctx context.Context, storeID uint) (*StoreSettings, error) {
	var st Store
	if err := r.DB(ctx).Where("id = ?", storeID).First(&st).Error; err != nil {
		return nil, mapDBErr(err)
	}
	s := &StoreSettings{
		Name:                st.Name,
		Address:             st.Address,
		Phone:               st.Phone,
		TaxEnabled:          st.TaxEnabled,
		TaxPct:              st.TaxPct,
		ReceiptHeader:       st.ReceiptHeader,
		ReceiptFooter:       st.ReceiptFooter,
		Paper:               st.Paper,
		Timezone:            st.Timezone,
		BusinessType:        st.BusinessType,
		Email:               st.Email,
		City:                st.City,
		Province:            st.Province,
		Currency:            st.Currency,
		Hours:               decodeStoreHours(st.Hours),
		ReceiptShowLogo:     st.ReceiptShowLogo,
		ReceiptShowCashier:  st.ReceiptShowCashier,
		ReceiptShowMethod:   st.ReceiptShowMethod,
		ReceiptShowTax:      st.ReceiptShowTax,
		ReceiptShowDiscount: st.ReceiptShowDiscount,
		ReceiptShowNote:     st.ReceiptShowNote,
		TaxName:             st.TaxName,
		TaxInclusive:        st.TaxInclusive,
		TaxRounding:         st.TaxRounding,
		TaxApplyTo:          st.TaxApplyTo,
		LoyaltyEnabled:      st.LoyaltyEnabled,
		LoyaltyEarnPer:      st.LoyaltyEarnPer,
		LoyaltyPointValue:   st.LoyaltyPointValue,
		CreditDueDays:       st.CreditDueDays,
		LowStockDefault:     st.LowStockDefault,
	}
	// Normalisasi default aman untuk baris lama / nilai kosong.
	if s.Currency == "" {
		s.Currency = "IDR"
	}
	if s.TaxRounding == "" {
		s.TaxRounding = "none"
	}
	if s.TaxApplyTo == "" {
		s.TaxApplyTo = "all"
	}
	if s.LoyaltyEarnPer <= 0 {
		s.LoyaltyEarnPer = 10000
	}
	return s, nil
}

// GetOutlet mengambil satu outlet milik toko (dipakai SettingsService
// untuk memastikan outlet sasaran ada di toko pemanggil sebelum membaca
// atau menulis override-nya; ErrNotFound bila tidak ada / milik toko
// lain).
func (r *StoreRepo) GetOutlet(ctx context.Context, storeID, outletID uint) (*Outlet, error) {
	var o Outlet
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", outletID, storeID).First(&o).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &o, nil
}

// GetOutletSettings mengambil baris override pengaturan outlet. Hasil
// (nil, nil) berarti outlet belum punya baris override sama sekali —
// bukan error.
func (r *StoreRepo) GetOutletSettings(ctx context.Context, storeID, outletID uint) (*OutletSettings, error) {
	var row OutletSettings
	err := r.DB(ctx).Where("store_id = ? AND outlet_id = ?", storeID, outletID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// UpsertOutletSettings menyisipkan baris override outlet atau
// memperbarui kolom override-nya bila baris (store_id, outlet_id) sudah
// ada, lalu mengembalikan baris tersimpan.
func (r *StoreRepo) UpsertOutletSettings(ctx context.Context, row *OutletSettings) (*OutletSettings, error) {
	err := r.DB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "store_id"}, {Name: "outlet_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tax_pct", "receipt_footer", "hours", "updated_at"}),
	}).Create(row).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return r.GetOutletSettings(ctx, row.StoreID, row.OutletID)
}

func (r *StoreRepo) GetTimezone(ctx context.Context, storeID uint) (string, error) {
	var st Store
	if err := r.DB(ctx).Where("id = ?", storeID).First(&st).Error; err != nil {
		return "Asia/Makassar", mapDBErr(err)
	}
	if st.Timezone == "" {
		return "Asia/Makassar", nil
	}
	return st.Timezone, nil
}

func (r *StoreRepo) UpdateSettings(ctx context.Context, storeID uint, s *StoreSettings) (*StoreSettings, error) {
	res := r.DB(ctx).Model(&Store{}).Where("id = ?", storeID).Updates(map[string]interface{}{
		"name":                  s.Name,
		"address":               s.Address,
		"phone":                 s.Phone,
		"tax_enabled":           s.TaxEnabled,
		"tax_pct":               s.TaxPct,
		"receipt_header":        s.ReceiptHeader,
		"receipt_footer":        s.ReceiptFooter,
		"paper":                 s.Paper,
		"timezone":              s.Timezone,
		"business_type":         s.BusinessType,
		"email":                 s.Email,
		"city":                  s.City,
		"province":              s.Province,
		"currency":              s.Currency,
		"hours":                 encodeStoreHours(s.Hours),
		"receipt_show_logo":     s.ReceiptShowLogo,
		"receipt_show_cashier":  s.ReceiptShowCashier,
		"receipt_show_method":   s.ReceiptShowMethod,
		"receipt_show_tax":      s.ReceiptShowTax,
		"receipt_show_discount": s.ReceiptShowDiscount,
		"receipt_show_note":     s.ReceiptShowNote,
		"tax_name":              s.TaxName,
		"tax_inclusive":         s.TaxInclusive,
		"tax_rounding":          s.TaxRounding,
		"tax_apply_to":          s.TaxApplyTo,
		"loyalty_enabled":       s.LoyaltyEnabled,
		"loyalty_earn_per":      s.LoyaltyEarnPer,
		"loyalty_point_value":   s.LoyaltyPointValue,
		"credit_due_days":       s.CreditDueDays,
		"low_stock_default":     s.LowStockDefault,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetSettings(ctx, storeID)
}

func (r *StoreRepo) SetPasscode(ctx context.Context, storeID, userID uint, hash *string) error {
	// Tabel users kini dimiliki slice internal/auth; paket ini tidak
	// boleh mengimpor auth (auth mengimpor outlet), jadi update lewat
	// nama tabel — semantik sama persis dengan Model(&User{}) lama.
	res := r.DB(ctx).Table("users").Where("id = ? AND store_id = ?", userID, storeID).Update("passcode_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
