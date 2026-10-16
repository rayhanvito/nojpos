package service

import (
	"context"
	"errors"
	"strings"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

var (
	ErrOutletNotFound  = errors.New("outlet tidak ditemukan di toko Anda")
	ErrOutletForbidden = errors.New("kasir hanya dapat menggunakan outlet yang ditetapkan")
)

type OutletService struct {
	outlets OutletRepository
}

func NewOutletService(outlets OutletRepository) *OutletService {
	return &OutletService{outlets: outlets}
}

func (s *OutletService) List(ctx context.Context, storeID uint) ([]model.Outlet, error) {
	return s.outlets.List(ctx, storeID)
}

func (s *OutletService) Create(ctx context.Context, storeID uint, in model.Outlet) (*model.Outlet, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	if in.Name == "" {
		return nil, errors.New("nama outlet wajib diisi")
	}
	return s.outlets.Create(ctx, storeID, &in)
}

func (s *OutletService) Update(ctx context.Context, storeID, id uint, in model.Outlet, isDefault *bool) (*model.Outlet, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	if in.Name == "" {
		return nil, errors.New("nama outlet wajib diisi")
	}
	outlet, err := s.outlets.Update(ctx, storeID, id, in, isDefault)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrOutletNotFound
	}
	return outlet, err
}

func (s *OutletService) Get(ctx context.Context, storeID, id uint) (*model.Outlet, error) {
	outlet, err := s.outlets.GetByID(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrOutletNotFound
	}
	return outlet, err
}
