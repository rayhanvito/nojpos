package outlet

import (
	"context"
	"errors"
	"strings"
)

// Repository adalah kontrak penyimpanan outlet yang dipakai Service
// (metode-metode OutletRepository lama di service/interfaces.go).
type Repository interface {
	List(ctx context.Context, storeID uint) ([]Outlet, error)
	GetByID(ctx context.Context, storeID, id uint) (*Outlet, error)
	GetDefault(ctx context.Context, storeID uint) (*Outlet, error)
	Create(ctx context.Context, storeID uint, outlet *Outlet) (*Outlet, error)
	Update(ctx context.Context, storeID, id uint, outlet Outlet, isDefault *bool) (*Outlet, error)
}

// Service adalah OutletService lama di service/outlet.go, dipindahkan
// verbatim.
type Service struct {
	outlets Repository
}

func NewService(outlets Repository) *Service {
	return &Service{outlets: outlets}
}

func (s *Service) List(ctx context.Context, storeID uint) ([]Outlet, error) {
	return s.outlets.List(ctx, storeID)
}

func (s *Service) Create(ctx context.Context, storeID uint, in Outlet) (*Outlet, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	if in.Name == "" {
		return nil, errors.New("nama outlet wajib diisi")
	}
	return s.outlets.Create(ctx, storeID, &in)
}

func (s *Service) Update(ctx context.Context, storeID, id uint, in Outlet, isDefault *bool) (*Outlet, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Address = strings.TrimSpace(in.Address)
	if in.Name == "" {
		return nil, errors.New("nama outlet wajib diisi")
	}
	outlet, err := s.outlets.Update(ctx, storeID, id, in, isDefault)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrOutletNotFound
	}
	return outlet, err
}

func (s *Service) Get(ctx context.Context, storeID, id uint) (*Outlet, error) {
	outlet, err := s.outlets.GetByID(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrOutletNotFound
	}
	return outlet, err
}
