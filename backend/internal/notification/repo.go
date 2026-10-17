package notification

import (
	"context"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Repo adalah repositori notifikasi (NotificationRepo lama di
// repo/notification.go); metode-metodenya verbatim, akses DB lewat
// DB(ctx).
type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

// createNotification adalah badan transaksi Create lama, verbatim:
// insert dengan OnConflict DoNothing, lalu retensi maksimal 10 terbaru
// per store_id dan type (FIFO) untuk baris tanpa dedup_key.
func createNotification(tx *gorm.DB, n *Notification) error {
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(n)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 || n.DedupKey != nil {
		return nil
	}
	// Retention: max 10 latest per store_id and type (FIFO)
	var oldIDs []uint
	err := tx.Model(&Notification{}).
		Where("store_id = ? AND type = ? AND dedup_key IS NULL", n.StoreID, n.Type).
		Order("created_at ASC, id ASC").
		Pluck("id", &oldIDs).Error
	if err != nil {
		return err
	}
	if len(oldIDs) > 10 {
		excess := oldIDs[:len(oldIDs)-10]
		if err := tx.Where("id IN ?", excess).Unscoped().Delete(&Notification{}).Error; err != nil {
			return err
		}
	}
	return nil
}

// Create menyimpan notifikasi.
//
// Semantik sink §4 (BE-F9a): bila ctx membawa transaksi pemanggil,
// insert berjalan di dalam transaksi bersarang (savepoint GORM) dan
// error-nya DITELAN (dicatat ke log, Create mengembalikan nil) —
// kegagalan notifikasi tidak boleh menggagalkan transaksi pemanggil.
// Bila ctx tanpa transaksi, perilakunya persis seperti semula: insert
// dalam transaksinya sendiri dan error dikembalikan apa adanya.
func (r *Repo) Create(ctx context.Context, n *Notification) error {
	if database.InTx(ctx) {
		err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
			return createNotification(tx, n)
		})
		if err != nil {
			log.Printf("notification: gagal menyimpan notifikasi, diabaikan agar transaksi pemanggil tetap commit: %v", err)
		}
		return nil
	}
	return r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		return createNotification(tx, n)
	})
}

func (r *Repo) List(ctx context.Context, storeID uint, category string, status string, unreadOnly bool, page, limit int) ([]*Notification, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	query := r.DB(ctx).Model(&Notification{}).Where("store_id = ?", storeID)
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if status == "unread" || unreadOnly {
		query = query.Where("read = ?", false)
	} else if status == "read" {
		query = query.Where("read = ?", true)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []*Notification
	offset := (page - 1) * limit
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
}

func (r *Repo) UnreadCount(ctx context.Context, storeID uint) (int64, error) {
	var total int64
	err := r.DB(ctx).Model(&Notification{}).Where("store_id = ? AND read = ?", storeID, false).Count(&total).Error
	return total, err
}

func (r *Repo) GetByID(ctx context.Context, storeID, id uint) (*Notification, error) {
	var n Notification
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&n).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &n, nil
}

func (r *Repo) MarkAsRead(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Model(&Notification{}).Where("id = ? AND store_id = ?", id, storeID).Update("read", true)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) MarkAllAsRead(ctx context.Context, storeID uint) error {
	return r.DB(ctx).Model(&Notification{}).Where("store_id = ? AND read = ?", storeID, false).Update("read", true).Error
}

func (r *Repo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Notification{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) HasUnreadForReference(ctx context.Context, storeID uint, referenceID string, notifType string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&Notification{}).
		Where("store_id = ? AND reference_id = ? AND type = ? AND read = ?", storeID, referenceID, notifType, false).
		Count(&count).Error
	return count > 0, err
}
