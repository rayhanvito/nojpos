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
	ErrStoreMismatch    = errors.New("akun tidak ditemukan di toko Anda")
	ErrNotEditable      = errors.New("hanya akun kasir yang dapat dinonaktifkan")
	ErrNotRenamable     = errors.New("hanya akun kasir yang dapat diubah")
	ErrCashierOpenShift = errors.New("kasir memiliki shift terbuka dan tidak dapat dipindahkan outlet")
)

type UserService struct {
	users    UserRepository
	cashiers CashierRepository
	outlets  OutletRepository
	shifts   ShiftRepository
}

func NewUserService(users UserRepository, cashiers CashierRepository, outlets ...OutletRepository) *UserService {
	s := &UserService{users: users, cashiers: cashiers}
	if len(outlets) > 0 {
		s.outlets = outlets[0]
	}
	return s
}

func (s *UserService) SetShiftRepository(shifts ShiftRepository) { s.shifts = shifts }

func (s *UserService) List(ctx context.Context, storeID uint) ([]model.PublicUser, error) {
	owners, err := s.users.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	cashiers, err := s.cashiers.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}

	out := make([]model.PublicUser, 0, len(owners)+len(cashiers))
	for _, u := range owners {
		out = append(out, u.Public())
	}
	storeName := ""
	if len(owners) > 0 {
		storeName = owners[0].StoreName
	}
	for _, c := range cashiers {
		out = append(out, c.Public(storeName))
	}
	return out, nil
}

func (s *UserService) CreateCashier(ctx context.Context, storeID, defaultOutletID uint, name string, requestedOutletID *uint) (*model.PublicUser, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("nama wajib diisi")
	}
	outletID := defaultOutletID
	if requestedOutletID != nil {
		outletID = *requestedOutletID
	}
	if outletID == 0 && s.outlets != nil {
		outlet, err := s.outlets.GetDefault(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = outlet.ID
	}
	if s.outlets != nil {
		if _, err := s.outlets.GetByID(ctx, storeID, outletID); err != nil {
			return nil, ErrOutletNotFound
		}
	}

	c, err := s.cashiers.Create(ctx, storeID, outletID, name)
	if err != nil {
		return nil, err
	}
	owners, _ := s.users.ListByStore(ctx, storeID)
	storeName := ""
	if len(owners) > 0 {
		storeName = owners[0].StoreName
	}
	pub := c.Public(storeName)
	return &pub, nil
}

func (s *UserService) SetActive(ctx context.Context, storeID, targetID uint, active bool) error {
	c, err := s.cashiers.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrStoreMismatch
		}
		return err
	}
	if c.StoreID != storeID {
		return ErrStoreMismatch
	}
	return s.cashiers.SetActive(ctx, targetID, active)
}

func (s *UserService) Delete(ctx context.Context, storeID, targetID uint) error {
	c, err := s.cashiers.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrStoreMismatch
		}
		return err
	}
	if c.StoreID != storeID {
		return ErrStoreMismatch
	}
	return s.cashiers.Delete(ctx, targetID)
}

func (s *UserService) Rename(ctx context.Context, storeID, targetID uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("Nama kasir wajib diisi.")
	}
	c, err := s.cashiers.GetByID(ctx, targetID)
	if err == nil {
		if c.StoreID != storeID {
			return ErrStoreMismatch
		}
		return s.cashiers.UpdateName(ctx, targetID, name)
	}
	if !errors.Is(err, repo.ErrNotFound) {
		return err
	}
	// Cashier not found: check if ID belongs to admin user in same store (ID collision case MEMORY.md §10)
	u, uErr := s.users.GetByID(ctx, targetID)
	if uErr == nil && u.StoreID == storeID {
		return ErrNotRenamable
	}
	return ErrStoreMismatch
}

func (s *UserService) UpdateCashier(ctx context.Context, storeID, targetID uint, rawName string, outletID *uint) error {
	name := strings.TrimSpace(rawName)
	var nameUpdate *string
	if name != "" {
		nameUpdate = &name
	}
	if nameUpdate == nil && outletID == nil {
		return fmt.Errorf("nama atau outlet harus diisi")
	}
	cashier, err := s.cashiers.GetByID(ctx, targetID)
	if err != nil || cashier.StoreID != storeID {
		return ErrStoreMismatch
	}
	if outletID != nil {
		if s.outlets == nil {
			return ErrOutletNotFound
		}
		if _, err := s.outlets.GetByID(ctx, storeID, *outletID); err != nil {
			return ErrOutletNotFound
		}
		if cashier.OutletID == nil || *outletID != *cashier.OutletID {
			if s.shifts == nil || cashier.OutletID == nil {
				return errors.New("outlet kasir atau repositori shift tidak tersedia")
			}
			if _, err := s.shifts.GetActiveShift(ctx, storeID, *cashier.OutletID, targetID); err == nil {
				return ErrCashierOpenShift
			} else if !errors.Is(err, repo.ErrNotFound) {
				return err
			}
		}
	}
	if err := s.cashiers.UpdateProfile(ctx, storeID, targetID, nameUpdate, outletID); err != nil {
		if errors.Is(err, repo.ErrCashierOpenShift) {
			return ErrCashierOpenShift
		}
		return err
	}
	return nil
}
