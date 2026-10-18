package apikey

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *Repo) Create(ctx context.Context, key *APIKey) error {
	return r.DB(ctx).Create(key).Error
}

func (r *Repo) List(ctx context.Context, storeID uint) ([]Metadata, error) {
	items := make([]Metadata, 0)
	err := r.DB(ctx).Model(&APIKey{}).
		Select("id, name, key_prefix, created_at, last_used_at").
		Where("store_id = ?", storeID).
		Order("created_at DESC, id DESC").
		Find(&items).Error
	return items, err
}

func (r *Repo) GetByHash(ctx context.Context, hash string) (*APIKey, error) {
	var key APIKey
	if err := r.DB(ctx).Where("key_hash = ?", hash).First(&key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *Repo) MarkUsed(ctx context.Context, keyID, storeID uint, at time.Time) error {
	result := r.DB(ctx).Model(&APIKey{}).
		Where("id = ? AND store_id = ?", keyID, storeID).
		UpdateColumn("last_used_at", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, storeID, id uint) error {
	result := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&APIKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
