package promo

import (
	"context"
	"errors"
	"fmt"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"math"
	"strings"
)

// Repository adalah kontrak penyimpanan promo yang dipakai Service
// (6 metode Repo).
type Repository interface {
	List(ctx context.Context, storeID uint) ([]*Promo, error)
	GetByID(ctx context.Context, storeID, id uint) (*Promo, error)
	Create(ctx context.Context, storeID uint, promo *Promo) (*Promo, error)
	Update(ctx context.Context, storeID, id uint, promo *Promo) (*Promo, error)
	SetActive(ctx context.Context, storeID, id uint, active bool) error
	Delete(ctx context.Context, storeID, id uint) error
}

// CategoryLookup hanya dipakai validate untuk cek keberadaan kategori.
type CategoryLookup interface {
	GetByID(ctx context.Context, storeID, id uint) (*catalog.Category, error)
}

// ProductLookup hanya dipakai validate untuk cek keberadaan produk.
type ProductLookup interface {
	GetByIDForStore(ctx context.Context, storeID, id uint) (*catalog.Product, error)
}

type Service struct {
	promos Repository
	cats   CategoryLookup
	prods  ProductLookup
}

func NewService(promos Repository, cats CategoryLookup, prods ProductLookup) *Service {
	return &Service{promos: promos, cats: cats, prods: prods}
}

func (s *Service) List(ctx context.Context, storeID uint) ([]*Promo, error) {
	return s.promos.List(ctx, storeID)
}

func (s *Service) Get(ctx context.Context, storeID, id uint) (*Promo, error) {
	return s.promos.GetByID(ctx, storeID, id)
}

func (s *Service) Create(ctx context.Context, storeID uint, in PromoInput) (*Promo, error) {
	promo, err := s.validate(ctx, storeID, in)
	if err != nil {
		return nil, err
	}
	promo.Active = true
	created, err := s.promos.Create(ctx, storeID, promo)
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrCodeTaken
	}
	return created, err
}

func (s *Service) Update(ctx context.Context, storeID, id uint, in PromoInput) (*Promo, error) {
	promo, err := s.validate(ctx, storeID, in)
	if err != nil {
		return nil, err
	}
	updated, err := s.promos.Update(ctx, storeID, id, promo)
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrCodeTaken
	}
	return updated, err
}

func (s *Service) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	return s.promos.SetActive(ctx, storeID, id, active)
}

func (s *Service) Delete(ctx context.Context, storeID, id uint) error {
	return s.promos.Delete(ctx, storeID, id)
}

func (s *Service) validate(ctx context.Context, storeID uint, in PromoInput) (*Promo, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, fmt.Errorf("%w: nama wajib diisi", ErrInvalid)
	}
	switch in.Type {
	case "percent":
		if math.IsNaN(in.Value) || math.IsInf(in.Value, 0) || in.Value <= 0 || in.Value > 100 {
			return nil, fmt.Errorf("%w: nilai persen harus lebih dari 0 dan maksimal 100", ErrInvalid)
		}
	case "amount":
		if in.AmountValue < 0 {
			return nil, fmt.Errorf("%w: potongan nominal harus berupa Rupiah bulat dan tidak negatif", ErrInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: tipe harus percent atau amount", ErrInvalid)
	}
	if in.Scope != "all" && in.Scope != "category" && in.Scope != "product" {
		return nil, fmt.Errorf("%w: cakupan promo tidak valid", ErrInvalid)
	}
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() || !in.StartsAt.Before(in.EndsAt) {
		return nil, fmt.Errorf("%w: tanggal mulai harus sebelum tanggal akhir", ErrInvalid)
	}
	if in.MinPurchase < 0 || in.MaxDiscount < 0 {
		return nil, fmt.Errorf("%w: minimum belanja dan batas potongan tidak boleh negatif", ErrInvalid)
	}
	if (in.HappyStart == nil) != (in.HappyEnd == nil) || in.HappyStart != nil && (!validHourMinute(*in.HappyStart) || !validHourMinute(*in.HappyEnd)) {
		return nil, fmt.Errorf("%w: jam promo harus lengkap dengan format HH:MM", ErrInvalid)
	}
	if in.Scope == "all" {
		if in.CategoryID != nil || in.ProductID != nil {
			return nil, fmt.Errorf("%w: cakupan semua tidak menerima ID kategori atau produk", ErrInvalid)
		}
	} else if in.Scope == "category" {
		if in.CategoryID == nil || *in.CategoryID == 0 || in.ProductID != nil {
			return nil, fmt.Errorf("%w: ID kategori wajib dan ID produk harus kosong", ErrInvalid)
		}
		if _, err := s.cats.GetByID(ctx, storeID, *in.CategoryID); err != nil {
			return nil, fmt.Errorf("%w: kategori tidak ditemukan di toko Anda", ErrInvalid)
		}
	} else {
		if in.ProductID == nil || *in.ProductID == 0 || in.CategoryID != nil {
			return nil, fmt.Errorf("%w: ID produk wajib dan ID kategori harus kosong", ErrInvalid)
		}
		if _, err := s.prods.GetByIDForStore(ctx, storeID, *in.ProductID); err != nil {
			return nil, fmt.Errorf("%w: produk tidak ditemukan di toko Anda", ErrInvalid)
		}
	}
	if in.VoucherCode != nil {
		code := strings.TrimSpace(*in.VoucherCode)
		if code == "" {
			in.VoucherCode = nil
		} else {
			in.VoucherCode = &code
		}
	}

	return &Promo{
		Name: in.Name, Type: in.Type, Value: in.Value, AmountValue: in.AmountValue, Scope: in.Scope,
		CategoryID: in.CategoryID, ProductID: in.ProductID, VoucherCode: in.VoucherCode,
		StartsAt: in.StartsAt, EndsAt: in.EndsAt, HappyStart: in.HappyStart, HappyEnd: in.HappyEnd,
		MinPurchase: in.MinPurchase, MaxDiscount: in.MaxDiscount, Combinable: in.Combinable,
	}, nil
}

// validHourMinute adalah salinan privat slice dari helper di
// service/settings.go (settings tidak diubah).
func validHourMinute(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	for _, idx := range []int{0, 1, 3, 4} {
		if s[idx] < '0' || s[idx] > '9' {
			return false
		}
	}
	hh := int(s[0]-'0')*10 + int(s[1]-'0')
	mm := int(s[3]-'0')*10 + int(s[4]-'0')
	return hh < 24 && mm < 60
}
