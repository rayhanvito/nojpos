package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/user"
)

// fakeStockEngine adalah StockEngine palsu tanpa efek samping.
type fakeStockEngine struct{}

func (fakeStockEngine) GetStock(ctx context.Context, outletID, productID uint, variantID *uint) (int, error) {
	return 0, nil
}

func (fakeStockEngine) SetStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, qty int) error {
	return nil
}

func (fakeStockEngine) RecordMovement(ctx context.Context, storeID, outletID, productID uint, variantID *uint, movementType string, qty int, reason, actor string) error {
	return nil
}

func (fakeStockEngine) IsDefaultOutlet(ctx context.Context, storeID, outletID uint) (bool, error) {
	return true, nil
}

// fakeTransactor menjalankan fn langsung tanpa transaksi sungguhan.
type fakeTransactor struct{}

func (fakeTransactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// fakeCats adalah CategoryRepository palsu: hanya metode yang dilalui
// jalur uji yang berperilaku; sisanya mengembalikan nilai nol.
type fakeCats struct {
	createErr              error
	getErr                 error
	deleteErr              error
	renameErr              error
	createCall             int
	renameCall             int
	renameName             string
	items                  []*Category
	productCounts          map[uint]int64
	productCountByCategory map[uint]int64
	productCountsCalls     int
	productCountsStoreID   uint
	productCountCalls      int
	productCountStoreID    uint
	productCountCategoryID uint
}

func (f *fakeCats) ListByStore(ctx context.Context, storeID uint) ([]*Category, error) {
	return f.items, nil
}

func (f *fakeCats) GetByID(ctx context.Context, storeID, id uint) (*Category, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return &Category{Name: "Minuman"}, nil
}

func (f *fakeCats) Create(ctx context.Context, storeID uint, name string) (*Category, error) {
	f.createCall++
	if f.createErr != nil {
		return nil, f.createErr
	}
	c := &Category{StoreID: storeID, Name: name, Active: true}
	c.ID = 3
	return c, nil
}

func (f *fakeCats) Rename(ctx context.Context, storeID, id uint, name string) (*Category, error) {
	f.renameCall++
	f.renameName = name
	if f.renameErr != nil {
		return nil, f.renameErr
	}
	c := &Category{StoreID: storeID, Name: name}
	c.ID = id
	return c, nil
}

func (f *fakeCats) ProductCountsByStore(ctx context.Context, storeID uint) (map[uint]int64, error) {
	f.productCountsCalls++
	f.productCountsStoreID = storeID
	return f.productCounts, nil
}

func (f *fakeCats) ProductCountByCategory(ctx context.Context, storeID, categoryID uint) (int64, error) {
	f.productCountCalls++
	f.productCountStoreID = storeID
	f.productCountCategoryID = categoryID
	return f.productCountByCategory[categoryID], nil
}

func (f *fakeCats) Delete(ctx context.Context, storeID, id uint) (bool, error) {
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	return false, nil
}

// fakeProds adalah ProductRepository palsu: yang dilalui jalur uji
// dikendalikan lewat field, sisanya nilai nol.
type fakeProds struct {
	skuTaken        bool
	variantSkuTaken bool
	createOut       *Product
	returned        *Product
	createCalls     int
}

func (f *fakeProds) DistinctUnits(ctx context.Context, storeID uint) ([]string, error) {
	return []string{}, nil
}

func (f *fakeProds) GetByIDForStore(ctx context.Context, storeID, id uint) (*Product, error) {
	return nil, ErrNotFound
}

func (f *fakeProds) GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*Product, error) {
	if f.returned != nil {
		return f.returned, nil
	}
	return nil, ErrNotFound
}

func (f *fakeProds) ListForStore(ctx context.Context, storeID uint, flt ProductFilter) (*ProductPage, error) {
	return &ProductPage{}, nil
}

func (f *fakeProds) ListForOutlet(ctx context.Context, storeID, outletID uint, flt ProductFilter) (*ProductPage, error) {
	return &ProductPage{}, nil
}

func (f *fakeProds) ListForExport(ctx context.Context, storeID uint) ([]*Product, error) {
	return nil, nil
}

func (f *fakeProds) ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*Product, error) {
	return nil, nil
}

func (f *fakeProds) InsertProduct(ctx context.Context, storeID, outletID uint, p *Product) error {
	f.createCalls++
	if f.createOut != nil {
		f.returned = f.createOut
		return nil
	}
	p.ID = 9
	f.returned = p
	return nil
}

func (f *fakeProds) DefaultOutlet(ctx context.Context, storeID uint) (*outlet.Outlet, error) {
	o := &outlet.Outlet{StoreID: storeID, IsDefault: true}
	o.ID = 1
	return o, nil
}

func (f *fakeProds) Update(ctx context.Context, storeID, id uint, p *Product) (*Product, error) {
	return p, nil
}

func (f *fakeProds) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	return nil
}

func (f *fakeProds) IsSkuTaken(ctx context.Context, storeID uint, sku string, exceptProductID uint) (bool, error) {
	return f.skuTaken, nil
}

func (f *fakeProds) ListVariantsForStore(ctx context.Context, storeID, productID uint) ([]ProductVariant, error) {
	return nil, nil
}

func (f *fakeProds) ListVariantsForOutlet(ctx context.Context, storeID, outletID, productID uint) ([]ProductVariant, error) {
	return nil, nil
}

func (f *fakeProds) GetVariant(ctx context.Context, storeID, productID, variantID uint) (*ProductVariant, error) {
	return nil, ErrNotFound
}

func (f *fakeProds) GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*ProductVariant, error) {
	return nil, ErrNotFound
}

func (f *fakeProds) InsertVariant(ctx context.Context, storeID, outletID, productID uint, variant *ProductVariant) error {
	return nil
}

func (f *fakeProds) UpdateVariantFields(ctx context.Context, storeID, productID, variantID uint, variant *ProductVariant) error {
	return nil
}

func (f *fakeProds) DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error {
	return nil
}

func (f *fakeProds) IsVariantSkuTaken(ctx context.Context, storeID uint, sku string, exceptID uint) (bool, error) {
	return f.variantSkuTaken, nil
}

func (f *fakeProds) ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]PriceTier, error) {
	return nil, nil
}

func (f *fakeProds) CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tier *PriceTier) (*PriceTier, error) {
	return tier, nil
}

func (f *fakeProds) UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, tier *PriceTier) (*PriceTier, error) {
	return tier, nil
}

func (f *fakeProds) DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error {
	return nil
}

func (f *fakeProds) Delete(ctx context.Context, storeID, id uint) error { return nil }

type notifyCall struct {
	itemID        uint
	productName   string
	variantName   string
	currentStock  int
	minStock      int
	unit          string
	referenceType string
}

type fakeNotifier struct {
	calls []notifyCall
}

func (f *fakeNotifier) CheckAndNotifyLowStock(ctx context.Context, storeID, itemID uint, productName, variantName string, currentStock, minStock int, unit, referenceType string) error {
	f.calls = append(f.calls, notifyCall{
		itemID: itemID, productName: productName, variantName: variantName,
		currentStock: currentStock, minStock: minStock, unit: unit, referenceType: referenceType,
	})
	return nil
}

func TestCreateCategoryBlankNameRejected(t *testing.T) {
	cats := &fakeCats{}
	svc := NewService(cats, &fakeProds{}, &fakeNotifier{})
	if _, err := svc.CreateCategory(context.Background(), 1, "   "); err == nil {
		t.Fatal("nama kategori kosong harus ditolak")
	}
	if cats.createCall != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama kosong, createCall=%d", cats.createCall)
	}
}

func TestCreateCategoryDuplicateBecomesCategoryTaken(t *testing.T) {
	cats := &fakeCats{createErr: ErrDuplicate}
	svc := NewService(cats, &fakeProds{}, &fakeNotifier{})
	_, err := svc.CreateCategory(context.Background(), 1, "Minuman")
	if !errors.Is(err, ErrCategoryTaken) {
		t.Fatalf("err = %v, mau ErrCategoryTaken", err)
	}
}

func TestCreateCategoryProductCountStartsAtZero(t *testing.T) {
	svc := NewService(&fakeCats{}, &fakeProds{}, nil)
	got, err := svc.CreateCategory(context.Background(), 1, "Minuman")
	if err != nil {
		t.Fatal(err)
	}
	assertCategoryProductCount(t, got, 0)
}

func TestListCategoriesSetsProductCountsFromOneStoreQuery(t *testing.T) {
	first := &Category{}
	first.ID = 1
	second := &Category{}
	second.ID = 2
	cats := &fakeCats{
		items:         []*Category{first, second},
		productCounts: map[uint]int64{1: 4},
	}
	svc := NewService(cats, &fakeProds{}, nil)
	got, err := svc.ListCategories(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if cats.productCountsCalls != 1 || cats.productCountsStoreID != 7 {
		t.Fatalf("ProductCountsByStore calls/store = %d/%d, mau 1/7", cats.productCountsCalls, cats.productCountsStoreID)
	}
	if len(got) != 2 {
		t.Fatalf("ListCategories items = %d, mau 2", len(got))
	}
	assertCategoryProductCount(t, got[0], 4)
	assertCategoryProductCount(t, got[1], 0)
}

func TestDeleteCategoryNotFoundBecomesStoreMismatch(t *testing.T) {
	cats := &fakeCats{deleteErr: ErrNotFound}
	svc := NewService(cats, &fakeProds{}, &fakeNotifier{})
	_, err := svc.DeleteCategory(context.Background(), 1, 99)
	if !errors.Is(err, user.ErrStoreMismatch) {
		t.Fatalf("err = %v, mau user.ErrStoreMismatch", err)
	}
}

func TestRenameCategoryTrimsNameBeforeRepo(t *testing.T) {
	cats := &fakeCats{}
	svc := NewService(cats, &fakeProds{}, nil)
	got, err := svc.RenameCategory(context.Background(), 1, 9, "  Minuman  ")
	if err != nil {
		t.Fatal(err)
	}
	if cats.renameCall != 1 || cats.renameName != "Minuman" || got.Name != "Minuman" {
		t.Fatalf("Rename harus menerima nama ter-trim sekali, call=%d name=%q hasil=%+v", cats.renameCall, cats.renameName, got)
	}
}

func TestRenameCategorySetsProductCount(t *testing.T) {
	cats := &fakeCats{productCountByCategory: map[uint]int64{9: 3}}
	svc := NewService(cats, &fakeProds{}, nil)
	got, err := svc.RenameCategory(context.Background(), 2, 9, "Minuman")
	if err != nil {
		t.Fatal(err)
	}
	if cats.productCountCalls != 1 || cats.productCountStoreID != 2 || cats.productCountCategoryID != 9 {
		t.Fatalf("ProductCountByCategory calls/store/category = %d/%d/%d, mau 1/2/9", cats.productCountCalls, cats.productCountStoreID, cats.productCountCategoryID)
	}
	assertCategoryProductCount(t, got, 3)
}

func assertCategoryProductCount(t *testing.T, category *Category, want int64) {
	t.Helper()
	payload, err := json.Marshal(category)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		ProductCount *int64 `json:"product_count"`
	}
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatal(err)
	}
	if got.ProductCount == nil || *got.ProductCount != want {
		t.Fatalf("product_count = %v, mau %d (JSON %s)", got.ProductCount, want, payload)
	}
}

func TestRenameCategoryBlankNameRejected(t *testing.T) {
	cats := &fakeCats{}
	svc := NewService(cats, &fakeProds{}, nil)
	if _, err := svc.RenameCategory(context.Background(), 1, 9, "  "); err == nil || err.Error() != "nama kategori wajib diisi" {
		t.Fatalf("nama kosong harus menghasilkan pesan validasi CreateCategory, dapat %v", err)
	}
	if cats.renameCall != 0 {
		t.Fatalf("repo tidak boleh dipanggil untuk nama kosong, call=%d", cats.renameCall)
	}
}

func TestRenameCategoryDuplicateBecomesCategoryTaken(t *testing.T) {
	cats := &fakeCats{renameErr: ErrDuplicate}
	svc := NewService(cats, &fakeProds{}, nil)
	if _, err := svc.RenameCategory(context.Background(), 1, 9, "Minuman"); !errors.Is(err, ErrCategoryTaken) {
		t.Fatalf("ErrDuplicate harus dipetakan ke ErrCategoryTaken, dapat %v", err)
	}
}

func TestRenameCategoryNotFoundBecomesStoreMismatch(t *testing.T) {
	cats := &fakeCats{renameErr: ErrNotFound}
	svc := NewService(cats, &fakeProds{}, nil)
	if _, err := svc.RenameCategory(context.Background(), 1, 9, "Minuman"); !errors.Is(err, user.ErrStoreMismatch) {
		t.Fatalf("ErrNotFound harus dipetakan ke ErrStoreMismatch, dapat %v", err)
	}
}

func TestCreateProductUnknownCategoryBecomesCategoryInvalid(t *testing.T) {
	cats := &fakeCats{getErr: ErrNotFound}
	svc := NewService(cats, &fakeProds{}, &fakeNotifier{})
	catID := uint(7)
	_, err := svc.CreateProduct(context.Background(), 1, 1, "Kasir", ProductInput{
		Name: "Kopi", SKU: "KPI-1", CategoryID: &catID,
	})
	if !errors.Is(err, ErrCategoryInvalid) {
		t.Fatalf("err = %v, mau ErrCategoryInvalid", err)
	}
}

func TestCreateProductSkuTaken(t *testing.T) {
	prods := &fakeProds{skuTaken: true}
	svc := NewService(&fakeCats{}, prods, &fakeNotifier{})
	_, err := svc.CreateProduct(context.Background(), 1, 1, "Kasir", ProductInput{
		Name: "Kopi", SKU: "KPI-1",
	})
	if !errors.Is(err, ErrSkuTaken) {
		t.Fatalf("err = %v, mau ErrSkuTaken", err)
	}
	if prods.createCalls != 0 {
		t.Fatalf("repo create tidak boleh dipanggil bila SKU terpakai, createCalls=%d", prods.createCalls)
	}
}

func TestCreateProductValidCallsRepoAndNotifier(t *testing.T) {
	out := &Product{Name: "Kopi", SKU: "KPI-1", Stock: 2, MinStock: 5, Unit: "pcs"}
	out.ID = 9
	prods := &fakeProds{createOut: out}
	notif := &fakeNotifier{}
	svc := NewService(&fakeCats{}, prods, notif)
	svc.SetStockEngine(fakeStockEngine{}, fakeTransactor{})
	got, err := svc.CreateProduct(context.Background(), 1, 1, "Kasir", ProductInput{
		Name: "Kopi", SKU: "KPI-1", Stock: 2,
	})
	if err != nil {
		t.Fatalf("err tak terduga: %v", err)
	}
	if got != out {
		t.Fatalf("produk kembali bukan hasil repo: %+v", got)
	}
	if prods.createCalls != 1 {
		t.Fatalf("createCalls = %d, mau 1", prods.createCalls)
	}
	// Service memanggil notifier setiap produk berhasil dibuat (notifier
	// yang memutuskan ambang stok rendah); argumennya harus data produk.
	if len(notif.calls) != 1 {
		t.Fatalf("notifier calls = %d, mau 1", len(notif.calls))
	}
	call := notif.calls[0]
	if call.itemID != 9 || call.productName != "Kopi" || call.currentStock != 2 || call.minStock != 5 || call.unit != "pcs" || call.referenceType != "product" {
		t.Fatalf("argumen notifier salah: %+v", call)
	}
}

func TestCreateProductNilNotifierSkipsNotify(t *testing.T) {
	prods := &fakeProds{}
	svc := NewService(&fakeCats{}, prods, nil)
	svc.SetStockEngine(fakeStockEngine{}, fakeTransactor{})
	if _, err := svc.CreateProduct(context.Background(), 1, 1, "Kasir", ProductInput{
		Name: "Teh", SKU: "TEH-1",
	}); err != nil {
		t.Fatalf("err tak terduga: %v", err)
	}
	if prods.createCalls != 1 {
		t.Fatalf("createCalls = %d, mau 1", prods.createCalls)
	}
}

func TestCreateVariantSkuTaken(t *testing.T) {
	prods := &fakeProds{variantSkuTaken: true}
	svc := NewService(&fakeCats{}, prods, &fakeNotifier{})
	_, err := svc.CreateVariant(context.Background(), 1, 1, 9, "Kasir", VariantInput{
		Name: "Besar", SKU: "KPI-1-B",
	})
	if !errors.Is(err, ErrSkuTaken) {
		t.Fatalf("err = %v, mau ErrSkuTaken", err)
	}
}
