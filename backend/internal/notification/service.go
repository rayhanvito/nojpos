package notification

import (
	"context"
	"fmt"
	"time"
)

// Repository adalah kontrak penyimpanan notifikasi yang dipakai
// Service (NotificationRepository lama di service/notification.go,
// bertipe slice).
type Repository interface {
	Create(ctx context.Context, n *Notification) error
	List(ctx context.Context, storeID uint, category string, status string, unreadOnly bool, page, limit int) ([]*Notification, int, error)
	UnreadCount(ctx context.Context, storeID uint) (int64, error)
	GetByID(ctx context.Context, storeID, id uint) (*Notification, error)
	MarkAsRead(ctx context.Context, storeID, id uint) error
	MarkAllAsRead(ctx context.Context, storeID uint) error
	Delete(ctx context.Context, storeID, id uint) error
	HasUnreadForReference(ctx context.Context, storeID uint, referenceID string, notifType string) (bool, error)
}

// StoreRepository adalah kontrak toko yang dipakai Service — hanya
// GetTimezone dari NotificationStoreRepository lama yang benar-benar
// dipakai (diimplementasikan outlet.StoreRepo).
type StoreRepository interface {
	GetTimezone(ctx context.Context, storeID uint) (string, error)
}

// Service adalah NotificationService lama di service/notification.go,
// dipindah verbatim.
type Service struct {
	notifRepo Repository
	stores    StoreRepository
}

func NewService(notifRepo Repository, stores ...StoreRepository) *Service {
	s := &Service{notifRepo: notifRepo}
	if len(stores) > 0 {
		s.stores = stores[0]
	}
	return s
}

func (s *Service) List(ctx context.Context, storeID uint, category, status string, unreadOnly bool, page, limit int) (*NotificationPage, error) {
	items, total, err := s.notifRepo.List(ctx, storeID, category, status, unreadOnly, page, limit)
	if err != nil {
		return nil, err
	}
	return &NotificationPage{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (s *Service) UnreadCount(ctx context.Context, storeID uint) (int64, error) {
	return s.notifRepo.UnreadCount(ctx, storeID)
}

func (s *Service) MarkAsRead(ctx context.Context, storeID, id uint) error {
	return s.notifRepo.MarkAsRead(ctx, storeID, id)
}

func (s *Service) MarkAllAsRead(ctx context.Context, storeID uint) error {
	return s.notifRepo.MarkAllAsRead(ctx, storeID)
}

func (s *Service) Delete(ctx context.Context, storeID, id uint) error {
	return s.notifRepo.Delete(ctx, storeID, id)
}

func (s *Service) Emit(ctx context.Context, storeID uint, title, message string, category NotificationCategory, notifType string, actorID, actorName, refType, refID string) error {
	n := &Notification{
		StoreID:       storeID,
		Title:         title,
		Message:       message,
		Category:      category,
		Type:          notifType,
		ActorID:       actorID,
		ActorName:     actorName,
		ReferenceType: refType,
		ReferenceID:   refID,
		Read:          false,
	}
	return s.notifRepo.Create(ctx, n)
}

func (s *Service) CheckAndNotifyLowStock(ctx context.Context, storeID, itemID uint, productName, variantName string, currentStock, minStock int, unit, referenceType string) error {
	if minStock <= 0 || currentStock > minStock {
		return nil
	}
	refIDStr := fmt.Sprintf("%d", itemID)
	notifType := "low_stock"
	if currentStock == 0 {
		notifType = "out_of_stock"
	}
	tz := "Asia/Makassar"
	if s.stores != nil {
		if storeTZ, err := s.stores.GetTimezone(ctx, storeID); err == nil && storeTZ != "" {
			tz = storeTZ
		}
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.FixedZone("UTC+8", 8*60*60)
	}
	dedupKey := fmt.Sprintf("%s:%d:%s", referenceType, itemID, time.Now().In(loc).Format("2006-01-02"))
	title := "Stok menipis"
	message := fmt.Sprintf("Produk %s tersisa %d %s.", productName, currentStock, unit)
	if referenceType == "variant" {
		message = fmt.Sprintf("Produk %s, varian %s tersisa %d %s.", productName, variantName, currentStock, unit)
	}
	if currentStock == 0 {
		title = "Stok habis"
		if referenceType == "variant" {
			message = fmt.Sprintf("Produk %s, varian %s habis (0 %s).", productName, variantName, unit)
		} else {
			message = fmt.Sprintf("Produk %s habis (0 %s).", productName, unit)
		}
	}
	n := &Notification{
		StoreID: storeID, Title: title, Message: message, Category: CategoryStok, Type: notifType,
		ReferenceType: referenceType, ReferenceID: refIDStr, DedupKey: &dedupKey,
	}
	return s.notifRepo.Create(ctx, n)
}
