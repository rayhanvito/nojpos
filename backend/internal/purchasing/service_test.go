package purchasing

import (
	"context"
	"errors"
	"testing"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
)

// Tes unit validasi yang benar-benar ada di service purchasing:
// validate() PurchaseOrderService, gerbang awal Receive, mapOrderError,
// dan normalizeSupplier SupplierService.

// fakeSuppliers adalah Suppliers palsu: hanya GetByID yang
// dikonfigurasi tes; sisanya nilai nol.
type fakeSuppliers struct {
	getErr error
}

func (f *fakeSuppliers) List(ctx context.Context, storeID uint) ([]*Supplier, error) {
	return []*Supplier{}, nil
}

func (f *fakeSuppliers) GetByID(ctx context.Context, storeID, id uint) (*Supplier, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &Supplier{Name: "Supplier Uji"}, nil
}

func (f *fakeSuppliers) Create(ctx context.Context, storeID uint, supplier *Supplier) (*Supplier, error) {
	return supplier, nil
}

func (f *fakeSuppliers) Update(ctx context.Context, storeID, id uint, supplier *Supplier) (*Supplier, error) {
	return supplier, nil
}

func (f *fakeSuppliers) Delete(ctx context.Context, storeID, id uint) error { return nil }

// fakeProducts adalah Products palsu dengan produk terkonfigurasi.
type fakeProducts struct {
	product *catalog.Product
	err     error
}

func (f *fakeProducts) GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*catalog.Product, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.product, nil
}

type fakeStores struct{}

func (fakeStores) GetTimezone(ctx context.Context, storeID uint) (string, error) {
	return "Asia/Jakarta", nil
}

func newValidateService(suppliers *fakeSuppliers, products *fakeProducts) *Service {
	return NewService(nil, suppliers, products, fakeStores{}, nil)
}

func TestValidateRejectsBadInputs(t *testing.T) {
	plain := &catalog.Product{Name: "Produk Polos"}
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{product: plain})
	ctx := context.Background()

	cases := []PurchaseOrderInput{
		{SupplierID: 0, Items: []PurchaseOrderItemInput{{ProductID: 1, Qty: 1}}},
		{SupplierID: 1, Items: nil},
		{SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 0, Qty: 1}}},
		{SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 1, Qty: 0}}},
		{SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 1, Qty: 1, BuyPrice: -1}}},
	}
	for i, in := range cases {
		if _, err := svc.validate(ctx, 1, 1, in); !errors.Is(err, ErrPurchaseOrderInvalid) {
			t.Fatalf("kasus %d harus ErrPurchaseOrderInvalid, dapat %v", i, err)
		}
	}
}

func TestValidateSupplierNotFound(t *testing.T) {
	svc := newValidateService(&fakeSuppliers{getErr: ErrNotFound}, &fakeProducts{product: &catalog.Product{}})
	_, err := svc.validate(context.Background(), 1, 1, PurchaseOrderInput{
		SupplierID: 9, Items: []PurchaseOrderItemInput{{ProductID: 1, Qty: 1}},
	})
	if !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("supplier hilang harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
}

func TestValidateProductNotFound(t *testing.T) {
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{err: ErrNotFound})
	_, err := svc.validate(context.Background(), 1, 1, PurchaseOrderInput{
		SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 9, Qty: 1}},
	})
	if !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("produk hilang harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
}

func TestValidateVariantRules(t *testing.T) {
	variant := catalog.ProductVariant{Name: "Besar"}
	variant.ID = 31
	withVariants := &catalog.Product{Name: "Produk Varian", Variants: []catalog.ProductVariant{variant}}
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{product: withVariants})
	ctx := context.Background()

	// Produk bervarian tanpa VariantID ditolak.
	if _, err := svc.validate(ctx, 1, 1, PurchaseOrderInput{
		SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 1, Qty: 2, BuyPrice: 3000}},
	}); !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("tanpa varian harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
	// VariantID yang tidak ada di produk ditolak.
	wrong := uint(99)
	if _, err := svc.validate(ctx, 1, 1, PurchaseOrderInput{
		SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 1, VariantID: &wrong, Qty: 2, BuyPrice: 3000}},
	}); !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("varian asing harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
	// VariantID yang benar lolos dan item terbangun verbatim.
	items, err := svc.validate(ctx, 1, 1, PurchaseOrderInput{
		SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 1, VariantID: &variant.ID, Qty: 2, BuyPrice: 3000}},
	})
	if err != nil {
		t.Fatalf("validate valid: %v", err)
	}
	if len(items) != 1 || items[0].QtyOrdered != 2 || items[0].BuyPrice != 3000 || items[0].VariantID == nil || *items[0].VariantID != 31 {
		t.Fatalf("item hasil validate salah: %+v", items)
	}
}

func TestValidateHappyPathPlainProduct(t *testing.T) {
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{product: &catalog.Product{Name: "Polos"}})
	items, err := svc.validate(context.Background(), 1, 1, PurchaseOrderInput{
		SupplierID: 1, Items: []PurchaseOrderItemInput{{ProductID: 7, Qty: 3, BuyPrice: 1500}},
	})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(items) != 1 || items[0].ProductID != 7 || items[0].QtyOrdered != 3 || items[0].BuyPrice != 1500 {
		t.Fatalf("item salah: %+v", items)
	}
}

func TestReceivePreValidation(t *testing.T) {
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{})
	ctx := context.Background()
	if _, err := svc.Receive(ctx, 1, 1, 1, nil, "kasir"); !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("penerimaan kosong harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
	if _, err := svc.Receive(ctx, 1, 1, 1, []PurchaseOrderReceiptInput{{ItemID: 1, Qty: 1}, {ItemID: 1, Qty: 2}}, "kasir"); !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("item duplikat harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
	if _, err := svc.Receive(ctx, 1, 1, 1, []PurchaseOrderReceiptInput{{ItemID: 1, Qty: 0}}, "kasir"); !errors.Is(err, ErrPurchaseOrderInvalid) {
		t.Fatalf("qty 0 harus ErrPurchaseOrderInvalid, dapat %v", err)
	}
}

func TestMapOrderError(t *testing.T) {
	svc := newValidateService(&fakeSuppliers{}, &fakeProducts{})
	if err := svc.mapOrderError(ErrNotFound); !errors.Is(err, ErrPurchaseOrderNotFound) {
		t.Fatalf("ErrNotFound harus menjadi ErrPurchaseOrderNotFound, dapat %v", err)
	}
	if err := svc.mapOrderError(ErrPurchaseOrderItemsLocked); !errors.Is(err, ErrPurchaseOrderLocked) {
		t.Fatalf("ItemsLocked harus menjadi ErrPurchaseOrderLocked, dapat %v", err)
	}
	for _, sentinel := range []error{ErrPurchaseOrderReceiptItem, ErrPurchaseOrderReceiptOver, ErrPurchaseOrderReceiptTarget} {
		if err := svc.mapOrderError(sentinel); !errors.Is(err, ErrPurchaseOrderInvalid) {
			t.Fatalf("%v harus menjadi ErrPurchaseOrderInvalid, dapat %v", sentinel, err)
		}
	}
	other := errors.New("lain")
	if err := svc.mapOrderError(other); !errors.Is(err, other) {
		t.Fatalf("error lain harus diteruskan apa adanya, dapat %v", err)
	}
	if err := svc.mapOrderError(nil); err != nil {
		t.Fatalf("nil harus tetap nil, dapat %v", err)
	}
}

func TestSupplierNormalizeAndNotFound(t *testing.T) {
	svc := NewSupplierService(&fakeSuppliers{})
	if _, err := svc.Create(context.Background(), 1, SupplierInput{Name: "  "}); !errors.Is(err, ErrSupplierNameRequired) {
		t.Fatalf("nama kosong harus ErrSupplierNameRequired, dapat %v", err)
	}
	created, err := svc.Create(context.Background(), 1, SupplierInput{Name: "  PT Sumber  ", Contact: " 081 ", Address: " Jl. B "})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Name != "PT Sumber" || created.Contact != "081" || created.Address != "Jl. B" {
		t.Fatalf("field supplier harus ter-trim, dapat %+v", created)
	}

	missing := NewSupplierService(&fakeSuppliers{getErr: ErrNotFound})
	if _, err := missing.Get(context.Background(), 1, 9); !errors.Is(err, ErrSupplierNotFound) {
		t.Fatalf("Get harus ErrSupplierNotFound, dapat %v", err)
	}
}
