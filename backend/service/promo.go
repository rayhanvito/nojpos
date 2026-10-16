package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

var (
	ErrPromoInvalid   = errors.New("data promo tidak valid")
	ErrPromoCodeTaken = errors.New("kode voucher sudah digunakan di toko ini")
)

type PromoInput struct {
	Name        string
	Type        string
	Value       float64
	AmountValue int64
	Scope       string
	CategoryID  *uint
	ProductID   *uint
	VoucherCode *string
	StartsAt    time.Time
	EndsAt      time.Time
	HappyStart  *string
	HappyEnd    *string
	MinPurchase int64
	MaxDiscount int64
	Combinable  bool
}

type PromoService struct {
	promos PromoRepository
	cats   CategoryRepository
	prods  ProductRepository
}

func NewPromoService(promos PromoRepository, cats CategoryRepository, prods ProductRepository) *PromoService {
	return &PromoService{promos: promos, cats: cats, prods: prods}
}

func (s *PromoService) List(ctx context.Context, storeID uint) ([]*model.Promo, error) {
	return s.promos.List(ctx, storeID)
}

func (s *PromoService) Get(ctx context.Context, storeID, id uint) (*model.Promo, error) {
	return s.promos.GetByID(ctx, storeID, id)
}

func (s *PromoService) Create(ctx context.Context, storeID uint, in PromoInput) (*model.Promo, error) {
	promo, err := s.validate(ctx, storeID, in)
	if err != nil {
		return nil, err
	}
	promo.Active = true
	created, err := s.promos.Create(ctx, storeID, promo)
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrPromoCodeTaken
	}
	return created, err
}

func (s *PromoService) Update(ctx context.Context, storeID, id uint, in PromoInput) (*model.Promo, error) {
	promo, err := s.validate(ctx, storeID, in)
	if err != nil {
		return nil, err
	}
	updated, err := s.promos.Update(ctx, storeID, id, promo)
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrPromoCodeTaken
	}
	return updated, err
}

func (s *PromoService) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	return s.promos.SetActive(ctx, storeID, id, active)
}

func (s *PromoService) Delete(ctx context.Context, storeID, id uint) error {
	return s.promos.Delete(ctx, storeID, id)
}

func (s *PromoService) validate(ctx context.Context, storeID uint, in PromoInput) (*model.Promo, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, fmt.Errorf("%w: nama wajib diisi", ErrPromoInvalid)
	}
	switch in.Type {
	case "percent":
		if math.IsNaN(in.Value) || math.IsInf(in.Value, 0) || in.Value <= 0 || in.Value > 100 {
			return nil, fmt.Errorf("%w: nilai persen harus lebih dari 0 dan maksimal 100", ErrPromoInvalid)
		}
	case "amount":
		if in.AmountValue < 0 {
			return nil, fmt.Errorf("%w: potongan nominal harus berupa Rupiah bulat dan tidak negatif", ErrPromoInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: tipe harus percent atau amount", ErrPromoInvalid)
	}
	if in.Scope != "all" && in.Scope != "category" && in.Scope != "product" {
		return nil, fmt.Errorf("%w: cakupan promo tidak valid", ErrPromoInvalid)
	}
	if in.StartsAt.IsZero() || in.EndsAt.IsZero() || !in.StartsAt.Before(in.EndsAt) {
		return nil, fmt.Errorf("%w: tanggal mulai harus sebelum tanggal akhir", ErrPromoInvalid)
	}
	if in.MinPurchase < 0 || in.MaxDiscount < 0 {
		return nil, fmt.Errorf("%w: minimum belanja dan batas potongan tidak boleh negatif", ErrPromoInvalid)
	}
	if (in.HappyStart == nil) != (in.HappyEnd == nil) || in.HappyStart != nil && (!validHourMinute(*in.HappyStart) || !validHourMinute(*in.HappyEnd)) {
		return nil, fmt.Errorf("%w: jam promo harus lengkap dengan format HH:MM", ErrPromoInvalid)
	}
	if in.Scope == "all" {
		if in.CategoryID != nil || in.ProductID != nil {
			return nil, fmt.Errorf("%w: cakupan semua tidak menerima ID kategori atau produk", ErrPromoInvalid)
		}
	} else if in.Scope == "category" {
		if in.CategoryID == nil || *in.CategoryID == 0 || in.ProductID != nil {
			return nil, fmt.Errorf("%w: ID kategori wajib dan ID produk harus kosong", ErrPromoInvalid)
		}
		if _, err := s.cats.GetByID(ctx, storeID, *in.CategoryID); err != nil {
			return nil, fmt.Errorf("%w: kategori tidak ditemukan di toko Anda", ErrPromoInvalid)
		}
	} else {
		if in.ProductID == nil || *in.ProductID == 0 || in.CategoryID != nil {
			return nil, fmt.Errorf("%w: ID produk wajib dan ID kategori harus kosong", ErrPromoInvalid)
		}
		if _, err := s.prods.GetByID(ctx, storeID, *in.ProductID); err != nil {
			return nil, fmt.Errorf("%w: produk tidak ditemukan di toko Anda", ErrPromoInvalid)
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

	return &model.Promo{
		Name: in.Name, Type: in.Type, Value: in.Value, AmountValue: in.AmountValue, Scope: in.Scope,
		CategoryID: in.CategoryID, ProductID: in.ProductID, VoucherCode: in.VoucherCode,
		StartsAt: in.StartsAt, EndsAt: in.EndsAt, HappyStart: in.HappyStart, HappyEnd: in.HappyEnd,
		MinPurchase: in.MinPurchase, MaxDiscount: in.MaxDiscount, Combinable: in.Combinable,
	}, nil
}
