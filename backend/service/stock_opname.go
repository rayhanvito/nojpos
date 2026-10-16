package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrStockOpnameNotFound = errors.New("sesi stock opname tidak ditemukan")
	ErrStockOpnameInvalid  = errors.New("data stock opname tidak valid")
	ErrStockOpnameActive   = errors.New("sesi stock opname aktif sudah ada")
	ErrStockOpnameState    = errors.New("status sesi stock opname tidak mengizinkan aksi ini")
)

type StockOpnameService struct {
	opnames StockOpnameRepository
}

func NewStockOpnameService(opnames StockOpnameRepository) *StockOpnameService {
	return &StockOpnameService{opnames: opnames}
}

func (s *StockOpnameService) List(ctx context.Context, storeID, outletID uint) ([]model.StockOpnameSession, error) {
	return s.opnames.List(ctx, storeID, outletID)
}

func (s *StockOpnameService) Get(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error) {
	session, err := s.opnames.GetByID(ctx, storeID, outletID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStockOpnameNotFound
	}
	return session, err
}

func (s *StockOpnameService) Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*model.StockOpnameSession, error) {
	session, err := s.opnames.Create(ctx, storeID, outletID, strings.TrimSpace(note), strings.TrimSpace(createdBy))
	return session, s.mapOpnameError(err)
}

func (s *StockOpnameService) UpdateItems(ctx context.Context, storeID, outletID, id uint, items []repo.StockOpnameItemCount) (*model.StockOpnameSession, error) {
	if len(items) == 0 {
		return nil, invalidStockOpname("Minimal satu item hitung wajib diisi.")
	}
	seen := make(map[uint]bool, len(items))
	for _, item := range items {
		if item.ItemID == 0 || item.CountedStock < 0 || seen[item.ItemID] {
			return nil, invalidStockOpname("Item hitung harus unik dan countedStock tidak boleh negatif.")
		}
		seen[item.ItemID] = true
	}
	session, err := s.opnames.UpdateItems(ctx, storeID, outletID, id, items)
	return session, s.mapOpnameError(err)
}

func (s *StockOpnameService) Post(ctx context.Context, storeID, outletID, id uint, actor string) (*model.StockOpnameSession, error) {
	session, err := s.opnames.Post(ctx, storeID, outletID, id, actor)
	return session, s.mapOpnameError(err)
}

func (s *StockOpnameService) Cancel(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error) {
	session, err := s.opnames.Cancel(ctx, storeID, outletID, id)
	return session, s.mapOpnameError(err)
}

func (s *StockOpnameService) mapOpnameError(err error) error {
	switch {
	case errors.Is(err, repo.ErrNotFound):
		return ErrStockOpnameNotFound
	case errors.Is(err, repo.ErrStockOpnameActive):
		return ErrStockOpnameActive
	case errors.Is(err, repo.ErrStockOpnameState):
		return ErrStockOpnameState
	case errors.Is(err, repo.ErrStockOpnameItem):
		return invalidStockOpname("Item tidak ditemukan dalam sesi stock opname ini.")
	case errors.Is(err, repo.ErrStockOpnameTarget):
		return invalidStockOpname("Produk atau varian tujuan stok tidak tersedia.")
	default:
		return err
	}
}

func invalidStockOpname(message string) error {
	return fmt.Errorf("%w: %s", ErrStockOpnameInvalid, message)
}
