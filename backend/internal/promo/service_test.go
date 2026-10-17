package promo

import (
	"context"
	"errors"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"testing"
	"time"
)

type fakeRepo struct {
	createErr   error
	createCalls int
	created     *Promo
}

func (f *fakeRepo) List(ctx context.Context, storeID uint) ([]*Promo, error) {
	return nil, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, storeID, id uint) (*Promo, error) {
	return nil, ErrNotFound
}

func (f *fakeRepo) Create(ctx context.Context, storeID uint, promo *Promo) (*Promo, error) {
	f.createCalls++
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.created = promo
	return promo, nil
}

func (f *fakeRepo) Update(ctx context.Context, storeID, id uint, promo *Promo) (*Promo, error) {
	return promo, nil
}

func (f *fakeRepo) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	return nil
}

func (f *fakeRepo) Delete(ctx context.Context, storeID, id uint) error { return nil }

type fakeCats struct{ err error }

func (f fakeCats) GetByID(ctx context.Context, storeID, id uint) (*catalog.Category, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &catalog.Category{}, nil
}

type fakeProds struct{ err error }

func (f fakeProds) GetByIDForStore(ctx context.Context, storeID, id uint) (*catalog.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &catalog.Product{}, nil
}

func validPromoInput() PromoInput {
	return PromoInput{
		Name:     "Promo Uji",
		Type:     "percent",
		Value:    10,
		Scope:    "all",
		StartsAt: time.Now().Add(-time.Hour),
		EndsAt:   time.Now().Add(24 * time.Hour),
	}
}

func newTestService(repo *fakeRepo, cats fakeCats, prods fakeProds) *Service {
	return NewService(repo, cats, prods)
}

func TestCreatePercentValidForcesActive(t *testing.T) {
	repo := &fakeRepo{}
	svc := newTestService(repo, fakeCats{}, fakeProds{})
	created, err := svc.Create(context.Background(), 1, validPromoInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("repo Create harus terpanggil 1x, dapat %d", repo.createCalls)
	}
	if created == nil || !created.Active {
		t.Fatalf("promo baru harus Active=true: %+v", created)
	}
}

func TestCreatePercentZeroInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	in.Value = 0
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("percent 0 harus ErrInvalid, dapat %v", err)
	}
}

func TestCreatePercentOver100Invalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	in.Value = 100.5
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("percent >100 harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateAmountNegativeInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	in.Type = "amount"
	in.AmountValue = -1000
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("amount negatif harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateUnknownTypeInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	in.Type = "gratis"
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("tipe tak dikenal harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateScopeCategoryWithoutIDInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	in.Scope = "category"
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("scope category tanpa CategoryID harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateScopeCategoryNotFoundInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{err: errors.New("tidak ada")}, fakeProds{})
	in := validPromoInput()
	in.Scope = "category"
	catID := uint(7)
	in.CategoryID = &catID
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("kategori tak ditemukan harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateScopeProductNotFoundInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{err: errors.New("tidak ada")})
	in := validPromoInput()
	in.Scope = "product"
	prodID := uint(9)
	in.ProductID = &prodID
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("produk tak ditemukan harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateScopeAllWithCategoryIDInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	catID := uint(3)
	in.CategoryID = &catID
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("scope all dengan CategoryID harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateHappyHoursUnpairedInvalid(t *testing.T) {
	svc := newTestService(&fakeRepo{}, fakeCats{}, fakeProds{})
	in := validPromoInput()
	start := "10:00"
	in.HappyStart = &start
	if _, err := svc.Create(context.Background(), 1, in); !errors.Is(err, ErrInvalid) {
		t.Fatalf("happy hours tak berpasangan harus ErrInvalid, dapat %v", err)
	}
}

func TestCreateVoucherCodeTrimmed(t *testing.T) {
	repo := &fakeRepo{}
	svc := newTestService(repo, fakeCats{}, fakeProds{})
	in := validPromoInput()
	code := "  HEMAT  "
	in.VoucherCode = &code
	created, err := svc.Create(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.VoucherCode == nil || *created.VoucherCode != "HEMAT" {
		t.Fatalf("kode harus ter-trim menjadi HEMAT, dapat %v", created.VoucherCode)
	}
}

func TestCreateVoucherCodeBlankBecomesNil(t *testing.T) {
	repo := &fakeRepo{}
	svc := newTestService(repo, fakeCats{}, fakeProds{})
	in := validPromoInput()
	code := "   "
	in.VoucherCode = &code
	created, err := svc.Create(context.Background(), 1, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.VoucherCode != nil {
		t.Fatalf("kode kosong harus menjadi nil, dapat %q", *created.VoucherCode)
	}
}

func TestCreateDuplicateMapsToCodeTaken(t *testing.T) {
	repo := &fakeRepo{createErr: ErrDuplicate}
	svc := newTestService(repo, fakeCats{}, fakeProds{})
	if _, err := svc.Create(context.Background(), 1, validPromoInput()); !errors.Is(err, ErrCodeTaken) {
		t.Fatalf("ErrDuplicate dari repo harus menjadi ErrCodeTaken, dapat %v", err)
	}
}
