package staff

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
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

// Repo adalah repositori kasir (CashierRepo lama di repo/cashier.go).
type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

// GetByIDUnscoped mengambil kasir berdasarkan ID saja, TANPA scope toko.
// Pemanggil WAJIB menurunkan & memverifikasi store dari record hasil
// (mis. membandingkan c.StoreID dengan storeID dari token) sebelum memakai
// atau mengubah datanya.
func (r *Repo) GetByIDUnscoped(ctx context.Context, id uint) (*Cashier, error) {
	var c Cashier
	err := r.DB(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *Repo) ListByStore(ctx context.Context, storeID uint) ([]*Cashier, error) {
	out := make([]*Cashier, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).Order("created_at ASC").Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) Create(ctx context.Context, storeID, outletID uint, name string) (*Cashier, error) {
	c := Cashier{
		StoreID:  storeID,
		OutletID: &outletID,
		Name:     name,
		Active:   true,
	}
	err := r.DB(ctx).Create(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *Repo) SetActive(ctx context.Context, id uint, active bool) error {
	res := r.DB(ctx).Model(&Cashier{}).Where("id = ?", id).Update("active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) SetPasscode(ctx context.Context, id uint, hash *string) error {
	res := r.DB(ctx).Model(&Cashier{}).Where("id = ?", id).Update("passcode_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) GetOrCreateByName(ctx context.Context, storeID uint, name string) (uint, error) {
	var c Cashier
	err := r.DB(ctx).Where("store_id = ? AND name = ?", storeID, name).First(&c).Error
	if err == nil {
		return c.ID, nil
	}
	var outlet outlet.Outlet
	if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return 0, mapDBErr(err)
	}
	created, err := r.Create(ctx, storeID, outlet.ID, name)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

func (r *Repo) Delete(ctx context.Context, id uint) error {
	res := r.DB(ctx).Delete(&Cashier{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateLastSeenAt(ctx context.Context, id uint, lastSeenAt *time.Time) error {
	res := r.DB(ctx).Model(&Cashier{}).Where("id = ?", id).Update("last_seen_at", lastSeenAt)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateName(ctx context.Context, id uint, name string) error {
	res := r.DB(ctx).Model(&Cashier{}).Where("id = ?", id).Update("name", name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) UpdateProfile(ctx context.Context, storeID, id uint, name *string, outletID *uint) error {
	updates := make(map[string]any, 2)
	if name != nil {
		updates["name"] = *name
	}
	if outletID != nil {
		updates["outlet_id"] = *outletID
	}
	if len(updates) == 0 {
		return nil
	}
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var cashier Cashier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", id, storeID).First(&cashier).Error; err != nil {
			return mapDBErr(err)
		}

		if outletID != nil && (cashier.OutletID == nil || *cashier.OutletID != *outletID) {
			var openShifts int64
			if err := tx.Table("cashier_shifts").Where("store_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, id).Count(&openShifts).Error; err != nil {
				return err
			}
			if openShifts > 0 {
				return ErrCashierOpenShift
			}
		}

		result := tx.Model(&cashier).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
