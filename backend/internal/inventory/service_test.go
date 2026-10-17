package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
)

// fakeProducts adalah ProductRepository palsu: hanya metode yang
// dilalui jalur uji yang berperilaku; sisanya mengembalikan nilai nol.
type fakeProducts struct {
	adjustErr        error
	adjustVariantErr error
	product          *catalog.Product
	variant          *catalog.ProductVariant
}

func (f *fakeProducts) GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*catalog.Product, error) {
	if f.product != nil {
		return f.product, nil
	}
	return &catalog.Product{}, nil
}

func (f *fakeProducts) GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*catalog.ProductVariant, error) {
	if f.variant != nil {
		return f.variant, nil
	}
	return &catalog.ProductVariant{}, nil
}

func (f *fakeProducts) AdjustStockAtOutlet(ctx context.Context, storeID, outletID, productID uint, delta int, reason, actor string) (*catalog.Product, error) {
	if f.adjustErr != nil {
		return nil, f.adjustErr
	}
	if f.product != nil {
		return f.product, nil
	}
	return &catalog.Product{}, nil
}

func (f *fakeProducts) AdjustVariantStockAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint, delta int, reason, actor string) (*catalog.ProductVariant, error) {
	if f.adjustVariantErr != nil {
		return nil, f.adjustVariantErr
	}
	if f.variant != nil {
		return f.variant, nil
	}
	return &catalog.ProductVariant{}, nil
}

// fakeMovements adalah MovementRepository palsu yang menangkap filter
// yang diterimanya supaya terjemahan DTO -> filter repo bisa diperiksa.
type fakeMovements struct {
	gotStoreID  uint
	gotOutletID uint
	gotFilter   MovementFilter
	page        *MovementPage
}

func (f *fakeMovements) ListForStore(ctx context.Context, storeID uint, filter MovementFilter) (*MovementPage, error) {
	f.gotStoreID = storeID
	f.gotFilter = filter
	return f.page, nil
}

func (f *fakeMovements) ListForOutlet(ctx context.Context, storeID, outletID uint, filter MovementFilter) (*MovementPage, error) {
	f.gotStoreID = storeID
	f.gotOutletID = outletID
	f.gotFilter = filter
	return f.page, nil
}

func TestAdjustStockInvalidDirection(t *testing.T) {
	svc := NewService(&fakeProducts{}, &fakeMovements{}, nil)
	_, err := svc.AdjustStock(context.Background(), 1, 1, 10, nil, "kasir", "sideways", 3, "koreksi")
	if !errors.Is(err, ErrBadDirection) {
		t.Fatalf("arah tidak valid harus ErrBadDirection, dapat %v", err)
	}
}

func TestListMovementsTranslatesFilter(t *testing.T) {
	wantPage := &MovementPage{Items: []*Movement{}, Total: 0, Page: 2, Limit: 10}
	movs := &fakeMovements{page: wantPage}
	svc := NewService(&fakeProducts{}, movs, nil)
	got, err := svc.ListMovements(context.Background(), 5, 9, MovementListFilter{
		Type: "sale", ProductID: 7, Q: "kopi", Page: 2, Limit: 10,
	})
	if err != nil {
		t.Fatalf("ListMovements: %v", err)
	}
	if got != wantPage {
		t.Fatalf("halaman harus diteruskan dari repo, dapat %+v", got)
	}
	if movs.gotStoreID != 5 || movs.gotOutletID != 9 {
		t.Fatalf("store/outlet salah terjemah: store=%d outlet=%d", movs.gotStoreID, movs.gotOutletID)
	}
	wantFilter := MovementFilter{Type: "sale", ProductID: 7, Q: "kopi", Page: 2, Limit: 10}
	if movs.gotFilter != wantFilter {
		t.Fatalf("filter salah terjemah: dapat %+v, mau %+v", movs.gotFilter, wantFilter)
	}
}
