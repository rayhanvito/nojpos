package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Seed transfer: satu toko (storeA) dengan outlet default A1 + cabang A2,
// satu toko lain (storeB) dengan outlet B1. Produk polos P1 berstok
// 10 di A1 dan 3 di A2; produk bervarian P2 dengan varian V berstok 5
// di A1 (baris stok varian di A2 sengaja belum ada).
type transferSeed struct {
	db       *gorm.DB
	svc      *Service
	storeA   uint
	outletA1 uint
	outletA2 uint
	outletB1 uint
	plain    uint
	withVar  uint
	variant  uint
}

func newTransferSeed(t *testing.T) transferSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "transfer.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &outlet.Outlet{}, &catalog.Product{}, &catalog.ProductVariant{}, &catalog.PriceTier{},
		&ProductStock{}, &Movement{}, &StockTransfer{}, &StockTransferItem{},
	); err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	storeA := &outlet.Store{Name: "Toko A"}
	must(db.Create(storeA).Error)
	storeB := &outlet.Store{Name: "Toko B"}
	must(db.Create(storeB).Error)
	a1 := &outlet.Outlet{StoreID: storeA.ID, Name: "Pusat", IsDefault: true}
	must(db.Create(a1).Error)
	a2 := &outlet.Outlet{StoreID: storeA.ID, Name: "Cabang"}
	must(db.Create(a2).Error)
	b1 := &outlet.Outlet{StoreID: storeB.ID, Name: "Toko Lain", IsDefault: true}
	must(db.Create(b1).Error)

	plain := &catalog.Product{StoreID: storeA.ID, OutletID: &a1.ID, Name: "Kopi Sachet", SKU: "KP-1", Active: true, Stock: 10}
	must(db.Create(plain).Error)
	withVar := &catalog.Product{StoreID: storeA.ID, OutletID: &a1.ID, Name: "Teh Botol", SKU: "TB-1", Active: true}
	must(db.Create(withVar).Error)
	variant := &catalog.ProductVariant{StoreID: storeA.ID, OutletID: &a1.ID, ProductID: withVar.ID, Name: "250ml", SKU: "TB-1-250", Stock: 5}
	must(db.Create(variant).Error)

	must(db.Create(&ProductStock{ProductID: plain.ID, OutletID: a1.ID, Qty: 10}).Error)
	must(db.Create(&ProductStock{ProductID: plain.ID, OutletID: a2.ID, Qty: 3}).Error)
	must(db.Create(&ProductStock{ProductID: withVar.ID, VariantID: &variant.ID, OutletID: a1.ID, Qty: 5}).Error)

	svc := NewService(catalog.NewProductRepo(db), NewMovementRepo(db), nil)
	svc.SetStockEngine(NewEngine(db), database.NewTransactor(db))
	svc.SetTransferRepo(NewTransferRepo(db))

	return transferSeed{
		db: db, svc: svc, storeA: storeA.ID,
		outletA1: a1.ID, outletA2: a2.ID, outletB1: b1.ID,
		plain: plain.ID, withVar: withVar.ID, variant: variant.ID,
	}
}

func (s transferSeed) stock(t *testing.T, outletID, productID uint, variantID *uint) int {
	t.Helper()
	qty, err := GetOutletStock(s.db, outletID, productID, variantID)
	if err != nil {
		t.Fatal(err)
	}
	return qty
}

func (s transferSeed) count(t *testing.T, model any) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(model).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCreateTransferMovesStockAndRecordsMovements(t *testing.T) {
	s := newTransferSeed(t)
	ctx := context.Background()

	tr, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", CreateTransferCmd{
		FromOutletID: s.outletA1, ToOutletID: s.outletA2, Note: "restok cabang",
		Items: []TransferItemCmd{{ProductID: s.plain, Qty: 4}},
	})
	if err != nil {
		t.Fatalf("CreateTransfer: %v", err)
	}
	if tr.ID == 0 || tr.FromOutletID != s.outletA1 || tr.ToOutletID != s.outletA2 || tr.CreatedBy != 77 || tr.CreatedByName != "Admin A" {
		t.Fatalf("header transfer salah: %+v", tr)
	}
	if len(tr.Items) != 1 || tr.Items[0].ProductID != s.plain || tr.Items[0].Qty != 4 || tr.Items[0].TransferID != tr.ID {
		t.Fatalf("item transfer salah: %+v", tr.Items)
	}
	if got := s.stock(t, s.outletA1, s.plain, nil); got != 6 {
		t.Fatalf("stok asal harus 6, dapat %d", got)
	}
	if got := s.stock(t, s.outletA2, s.plain, nil); got != 7 {
		t.Fatalf("stok tujuan harus 7, dapat %d", got)
	}

	// Cermin stok produk (outlet default) ikut bergerak.
	var p catalog.Product
	if err := s.db.First(&p, s.plain).Error; err != nil {
		t.Fatal(err)
	}
	if p.Stock != 6 {
		t.Fatalf("cermin stok produk harus 6, dapat %d", p.Stock)
	}

	var movs []Movement
	if err := s.db.Order("id ASC").Find(&movs).Error; err != nil {
		t.Fatal(err)
	}
	if len(movs) != 2 {
		t.Fatalf("harus ada 2 movement, dapat %d", len(movs))
	}
	wantReason := fmt.Sprintf("Transfer #%d: restok cabang", tr.ID)
	out, in := movs[0], movs[1]
	if out.Type != MovementTransferOut || out.Qty != -4 || out.OutletID == nil || *out.OutletID != s.outletA1 || out.Reason != wantReason || out.Actor != "Admin A" {
		t.Fatalf("movement keluar salah: %+v", out)
	}
	if in.Type != MovementTransferIn || in.Qty != 4 || in.OutletID == nil || *in.OutletID != s.outletA2 || in.Reason != wantReason {
		t.Fatalf("movement masuk salah: %+v", in)
	}
}

func TestCreateTransferVariantCreatesDestinationRow(t *testing.T) {
	s := newTransferSeed(t)
	ctx := context.Background()

	_, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", CreateTransferCmd{
		FromOutletID: s.outletA1, ToOutletID: s.outletA2,
		Items: []TransferItemCmd{{ProductID: s.withVar, VariantID: &s.variant, Qty: 2}},
	})
	if err != nil {
		t.Fatalf("CreateTransfer varian: %v", err)
	}
	if got := s.stock(t, s.outletA1, s.withVar, &s.variant); got != 3 {
		t.Fatalf("stok varian asal harus 3, dapat %d", got)
	}
	if got := s.stock(t, s.outletA2, s.withVar, &s.variant); got != 2 {
		t.Fatalf("stok varian tujuan harus 2 (baris baru), dapat %d", got)
	}
	if n := s.count(t, &Movement{}); n != 2 {
		t.Fatalf("harus ada 2 movement varian, dapat %d", n)
	}
}

func TestCreateTransferInsufficientStockRollsBackEverything(t *testing.T) {
	s := newTransferSeed(t)
	ctx := context.Background()

	// Item pertama sah (2), item kedua melebihi stok (50 > sisa) —
	// seluruh transfer harus batal: bukti rollback penuh.
	_, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", CreateTransferCmd{
		FromOutletID: s.outletA1, ToOutletID: s.outletA2,
		Items: []TransferItemCmd{
			{ProductID: s.plain, Qty: 2},
			{ProductID: s.plain, Qty: 50},
		},
	})
	if !errors.Is(err, catalog.ErrStockInsufficient) {
		t.Fatalf("harus ErrStockInsufficient, dapat %v", err)
	}
	if got := s.stock(t, s.outletA1, s.plain, nil); got != 10 {
		t.Fatalf("rollback: stok asal harus tetap 10, dapat %d", got)
	}
	if got := s.stock(t, s.outletA2, s.plain, nil); got != 3 {
		t.Fatalf("rollback: stok tujuan harus tetap 3, dapat %d", got)
	}
	if n := s.count(t, &StockTransfer{}); n != 0 {
		t.Fatalf("rollback: header transfer harus 0, dapat %d", n)
	}
	if n := s.count(t, &StockTransferItem{}); n != 0 {
		t.Fatalf("rollback: item transfer harus 0, dapat %d", n)
	}
	if n := s.count(t, &Movement{}); n != 0 {
		t.Fatalf("rollback: movement harus 0, dapat %d", n)
	}
}

func TestCreateTransferRejections(t *testing.T) {
	s := newTransferSeed(t)
	ctx := context.Background()
	base := CreateTransferCmd{FromOutletID: s.outletA1, ToOutletID: s.outletA2}

	same := base
	same.ToOutletID = s.outletA1
	same.Items = []TransferItemCmd{{ProductID: s.plain, Qty: 1}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", same); !errors.Is(err, ErrTransferSameOutlet) {
		t.Fatalf("outlet sama harus ErrTransferSameOutlet, dapat %v", err)
	}

	empty := base
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", empty); !errors.Is(err, ErrTransferItemsEmpty) {
		t.Fatalf("items kosong harus ErrTransferItemsEmpty, dapat %v", err)
	}

	zero := base
	zero.Items = []TransferItemCmd{{ProductID: s.plain, Qty: 0}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", zero); !errors.Is(err, ErrTransferInvalidQty) {
		t.Fatalf("qty 0 harus ErrTransferInvalidQty, dapat %v", err)
	}

	fromOther := base
	fromOther.FromOutletID = s.outletB1
	fromOther.Items = []TransferItemCmd{{ProductID: s.plain, Qty: 1}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", fromOther); !errors.Is(err, ErrNotFound) {
		t.Fatalf("outlet asal toko lain harus ErrNotFound, dapat %v", err)
	}

	toOther := base
	toOther.ToOutletID = s.outletB1
	toOther.Items = []TransferItemCmd{{ProductID: s.plain, Qty: 1}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", toOther); !errors.Is(err, ErrNotFound) {
		t.Fatalf("outlet tujuan toko lain harus ErrNotFound, dapat %v", err)
	}

	ghost := base
	ghost.Items = []TransferItemCmd{{ProductID: 999999, Qty: 1}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", ghost); !errors.Is(err, ErrNotFound) {
		t.Fatalf("produk tak dikenal harus ErrNotFound, dapat %v", err)
	}

	noVariant := base
	noVariant.Items = []TransferItemCmd{{ProductID: s.withVar, Qty: 1}}
	if _, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", noVariant); !errors.Is(err, catalog.ErrVariantRequired) {
		t.Fatalf("produk bervarian tanpa varian harus ErrVariantRequired, dapat %v", err)
	}

	if n := s.count(t, &StockTransfer{}); n != 0 {
		t.Fatalf("semua penolakan tidak boleh menyisakan transfer, dapat %d", n)
	}
	if n := s.count(t, &Movement{}); n != 0 {
		t.Fatalf("semua penolakan tidak boleh menyisakan movement, dapat %d", n)
	}
}

func TestListTransfersFiltersAndOrder(t *testing.T) {
	s := newTransferSeed(t)
	ctx := context.Background()
	mk := func(from, to uint, qty int) *StockTransfer {
		t.Helper()
		tr, err := s.svc.CreateTransfer(ctx, s.storeA, 77, "Admin A", CreateTransferCmd{
			FromOutletID: from, ToOutletID: to,
			Items: []TransferItemCmd{{ProductID: s.plain, Qty: qty}},
		})
		if err != nil {
			t.Fatalf("buat transfer: %v", err)
		}
		return tr
	}
	t1 := mk(s.outletA1, s.outletA2, 1)
	t2 := mk(s.outletA2, s.outletA1, 1)
	t3 := mk(s.outletA1, s.outletA2, 2)

	all, err := s.svc.ListTransfers(ctx, s.storeA, TransferListFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("daftar harus berisi 3 transfer, dapat %d", len(all))
	}
	if all[0].ID != t3.ID || all[1].ID != t2.ID || all[2].ID != t1.ID {
		t.Fatalf("urutan harus terbaru dulu: dapat %d,%d,%d", all[0].ID, all[1].ID, all[2].ID)
	}
	if len(all[0].Items) != 1 || all[0].Items[0].Qty != 2 {
		t.Fatalf("setiap baris harus menyertakan items: %+v", all[0].Items)
	}

	fromA1, err := s.svc.ListTransfers(ctx, s.storeA, TransferListFilter{FromOutletID: s.outletA1})
	if err != nil {
		t.Fatal(err)
	}
	if len(fromA1) != 2 {
		t.Fatalf("filter fromOutletId=A1 harus 2, dapat %d", len(fromA1))
	}

	toA1, err := s.svc.ListTransfers(ctx, s.storeA, TransferListFilter{ToOutletID: s.outletA1})
	if err != nil {
		t.Fatal(err)
	}
	if len(toA1) != 1 || toA1[0].ID != t2.ID {
		t.Fatalf("filter toOutletId=A1 harus hanya t2, dapat %+v", toA1)
	}

	today := time.Now().Format("2006-01-02")
	inRange, err := s.svc.ListTransfers(ctx, s.storeA, TransferListFilter{From: today, To: today})
	if err != nil {
		t.Fatal(err)
	}
	if len(inRange) != 3 {
		t.Fatalf("rentang hari ini harus 3, dapat %d", len(inRange))
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	outRange, err := s.svc.ListTransfers(ctx, s.storeA, TransferListFilter{From: yesterday, To: yesterday})
	if err != nil {
		t.Fatal(err)
	}
	if len(outRange) != 0 {
		t.Fatalf("rentang kemarin harus 0, dapat %d", len(outRange))
	}
}

func TestTransferHandlersCreateAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := newTransferSeed(t)
	h := NewHandler(s.svc)
	claims := &middleware.Claims{StoreID: s.storeA, UserID: 77, Name: "Admin A", Role: middleware.RoleAdmin}

	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/stock/transfers", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("claims", claims)
		h.CreateTransfer(c)
		return w
	}

	body := fmt.Sprintf(`{"fromOutletId":%d,"toOutletId":%d,"items":[{"productId":%d,"qty":4}],"note":"restok cabang"}`,
		s.outletA1, s.outletA2, s.plain)
	w := post(body)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST harus 201, dapat %d: %s", w.Code, w.Body.String())
	}
	var created StockTransfer
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.FromOutletID != s.outletA1 || created.ToOutletID != s.outletA2 || len(created.Items) != 1 || created.Items[0].Qty != 4 {
		t.Fatalf("respons POST salah: %s", w.Body.String())
	}
	t.Logf("POST /stock/transfers -> %d %s", w.Code, w.Body.String())

	sameBody := fmt.Sprintf(`{"fromOutletId":%d,"toOutletId":%d,"items":[{"productId":%d,"qty":1}]}`,
		s.outletA1, s.outletA1, s.plain)
	if w := post(sameBody); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "TRANSFER_SAME_OUTLET") {
		t.Fatalf("POST outlet sama harus 400 TRANSFER_SAME_OUTLET, dapat %d: %s", w.Code, w.Body.String())
	}

	get := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/stock/transfers"+query, nil)
		c.Set("claims", claims)
		h.ListTransfers(c)
		return w
	}

	w = get("?page=1&limit=1")
	if w.Code != http.StatusOK {
		t.Fatalf("GET harus 200, dapat %d: %s", w.Code, w.Body.String())
	}
	var page struct {
		Items []StockTransfer `json:"items"`
		Total int             `json:"total"`
		Page  int             `json:"page"`
		Limit int             `json:"limit"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Page != 1 || page.Limit != 1 || len(page.Items) != 1 {
		t.Fatalf("amplop paginasi salah: %s", w.Body.String())
	}
	t.Logf("GET /stock/transfers?page=1&limit=1 -> %d %s", w.Code, w.Body.String())

	w = get(fmt.Sprintf("?fromOutletId=%d", s.outletA2))
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Fatalf("filter fromOutletId=A2 harus total 0, dapat %s", w.Body.String())
	}
}
