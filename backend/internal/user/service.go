package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// UserRepository adalah kontrak penyimpanan user yang dipakai
// Service — hanya metode yang benar-benar dipakai (tipe auth.User).
type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*auth.User, error)
	ListByStore(ctx context.Context, storeID uint) ([]*auth.User, error)
	SetPasscode(ctx context.Context, id uint, hash *string) error
}

// CashierRepository adalah kontrak penyimpanan kasir yang dipakai
// Service — hanya metode yang benar-benar dipakai (tipe staff.Cashier).
type CashierRepository interface {
	GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error)
	ListByStore(ctx context.Context, storeID uint) ([]*staff.Cashier, error)
	Create(ctx context.Context, storeID, outletID uint, name string) (*staff.Cashier, error)
	UpdateProfile(ctx context.Context, storeID, id uint, name *string, outletID *uint) error
	SetActive(ctx context.Context, id uint, active bool) error
	SetPasscode(ctx context.Context, id uint, hash *string) error
	Delete(ctx context.Context, id uint) error
	UpdateName(ctx context.Context, id uint, name string) error
}

// OutletRepository adalah kontrak baca outlet yang dipakai Service.
type OutletRepository interface {
	GetByID(ctx context.Context, storeID, id uint) (*outlet.Outlet, error)
	GetDefault(ctx context.Context, storeID uint) (*outlet.Outlet, error)
}

// ShiftRepository adalah kontrak shift yang dipakai Service — hanya
// GetActiveShift dari ShiftRepository lama yang benar-benar dipakai.
type ShiftRepository interface {
	GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*shift.CashierShift, error)
}

// ActivityLogger adalah kontrak pencatat aktivitas untuk handler
// (diimplementasikan activity.Service; paket ini tidak
// mengimpor paket itu).
type ActivityLogger interface {
	Log(ctx context.Context, storeID uint, actorID, actorName, action, detail, refType, refID string)
}

// Service adalah UserService lama di service/user.go (BE-F1d), plus
// SetPasscode dari SettingsService lama — badan metode verbatim.

type Service struct {
	users    UserRepository
	cashiers CashierRepository
	outlets  OutletRepository
	shifts   ShiftRepository
}

func NewService(users UserRepository, cashiers CashierRepository, outlets ...OutletRepository) *Service {
	s := &Service{users: users, cashiers: cashiers}
	if len(outlets) > 0 {
		s.outlets = outlets[0]
	}
	return s
}

func (s *Service) SetShiftRepository(shifts ShiftRepository) { s.shifts = shifts }

func (s *Service) List(ctx context.Context, storeID uint) ([]auth.PublicUser, error) {
	owners, err := s.users.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	cashiers, err := s.cashiers.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}

	out := make([]auth.PublicUser, 0, len(owners)+len(cashiers))
	for _, u := range owners {
		out = append(out, u.Public())
	}
	storeName := ""
	if len(owners) > 0 {
		storeName = owners[0].StoreName
	}
	for _, c := range cashiers {
		out = append(out, auth.PublicUserFromCashier(c, storeName))
	}
	return out, nil
}

func (s *Service) CreateCashier(ctx context.Context, storeID, defaultOutletID uint, name string, requestedOutletID *uint) (*auth.PublicUser, error) {
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
	pub := auth.PublicUserFromCashier(c, storeName)
	return &pub, nil
}

func (s *Service) SetActive(ctx context.Context, storeID, targetID uint, active bool) error {
	c, err := s.cashiers.GetByIDUnscoped(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrStoreMismatch
		}
		return err
	}
	if c.StoreID != storeID {
		return ErrStoreMismatch
	}
	return s.cashiers.SetActive(ctx, targetID, active)
}

func (s *Service) Delete(ctx context.Context, storeID, targetID uint) error {
	c, err := s.cashiers.GetByIDUnscoped(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrStoreMismatch
		}
		return err
	}
	if c.StoreID != storeID {
		return ErrStoreMismatch
	}
	return s.cashiers.Delete(ctx, targetID)
}

func (s *Service) UpdateCashier(ctx context.Context, storeID, targetID uint, rawName string, outletID *uint) error {
	name := strings.TrimSpace(rawName)
	var nameUpdate *string
	if name != "" {
		nameUpdate = &name
	}
	if nameUpdate == nil && outletID == nil {
		return fmt.Errorf("nama atau outlet harus diisi")
	}
	cashier, err := s.cashiers.GetByIDUnscoped(ctx, targetID)
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
			} else if !errors.Is(err, ErrNotFound) {
				return err
			}
		}
	}
	if err := s.cashiers.UpdateProfile(ctx, storeID, targetID, nameUpdate, outletID); err != nil {
		if errors.Is(err, staff.ErrCashierOpenShift) {
			return ErrCashierOpenShift
		}
		return err
	}
	return nil
}

// SetPasscode dipindahkan verbatim dari SettingsService lama di
// service/settings.go (BE-F1d): kini milik Service di paket ini.
// SetPasscode sets the 5-digit PIN for an owner or cashier. roleHint comes
// from the /users list entry the admin clicked ("admin"/"cashier"/"") —
// users and cashiers have independent numeric sequences, so one number can
// exist in both tables. Empty hint keeps the legacy cashier-first order.
func (s *Service) SetPasscode(ctx context.Context, storeID, targetID uint, passcode, roleHint string) error {
	passcode = strings.TrimSpace(passcode)
	var hash *string
	if passcode != "" {
		if len(passcode) != 5 {
			return fmt.Errorf("passcode harus 5 angka")
		}
		for _, ch := range passcode {
			if ch < '0' || ch > '9' {
				return fmt.Errorf("passcode harus 5 angka")
			}
		}
		h, err := bcrypt.GenerateFromPassword([]byte(passcode), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		hs := string(h)
		hash = &hs
	}

	tryCashier := roleHint == "" || roleHint == string(middleware.RoleCashier)
	tryOwner := roleHint == "" || roleHint == string(middleware.RoleAdmin)

	if tryCashier {
		if c, err := s.cashiers.GetByIDUnscoped(ctx, targetID); err == nil && c.StoreID == storeID {
			return s.cashiers.SetPasscode(ctx, targetID, hash)
		}
	}
	if tryOwner {
		if owner, err := s.users.GetByID(ctx, targetID); err == nil && owner.StoreID == storeID {
			return s.users.SetPasscode(ctx, targetID, hash)
		}
	}

	return ErrStoreMismatch
}
