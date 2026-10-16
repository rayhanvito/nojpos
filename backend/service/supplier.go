package service

import (
	"context"
	"errors"
	"strings"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrSupplierNotFound     = errors.New("supplier tidak ditemukan")
	ErrSupplierNameRequired = errors.New("nama supplier wajib diisi")
)

type SupplierInput struct {
	Name    string
	Contact string
	Address string
}

type SupplierService struct {
	suppliers SupplierRepository
}

func NewSupplierService(suppliers SupplierRepository) *SupplierService {
	return &SupplierService{suppliers: suppliers}
}

func (s *SupplierService) List(ctx context.Context, storeID uint) ([]*model.Supplier, error) {
	return s.suppliers.List(ctx, storeID)
}

func (s *SupplierService) Get(ctx context.Context, storeID, id uint) (*model.Supplier, error) {
	supplier, err := s.suppliers.GetByID(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrSupplierNotFound
	}
	return supplier, err
}

func (s *SupplierService) Create(ctx context.Context, storeID uint, in SupplierInput) (*model.Supplier, error) {
	supplier, err := normalizeSupplier(in)
	if err != nil {
		return nil, err
	}
	return s.suppliers.Create(ctx, storeID, supplier)
}

func (s *SupplierService) Update(ctx context.Context, storeID, id uint, in SupplierInput) (*model.Supplier, error) {
	supplier, err := normalizeSupplier(in)
	if err != nil {
		return nil, err
	}
	updated, err := s.suppliers.Update(ctx, storeID, id, supplier)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrSupplierNotFound
	}
	return updated, err
}

func (s *SupplierService) Delete(ctx context.Context, storeID, id uint) error {
	err := s.suppliers.Delete(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrSupplierNotFound
	}
	return err
}

func normalizeSupplier(in SupplierInput) (*model.Supplier, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, ErrSupplierNameRequired
	}
	return &model.Supplier{
		Name:    in.Name,
		Contact: strings.TrimSpace(in.Contact),
		Address: strings.TrimSpace(in.Address),
	}, nil
}
