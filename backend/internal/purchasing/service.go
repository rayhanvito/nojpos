package purchasing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Orders adalah kontrak penyimpanan purchase order yang dipakai
// Service (PurchaseOrderRepository lama di service/interfaces.go,
// bertipe slice).
type Orders interface {
	List(ctx context.Context, storeID, outletID uint, f PurchaseOrderFilter) ([]*PurchaseOrder, error)
	GetByID(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error)
	CreateDraft(ctx context.Context, storeID, outletID, supplierID uint, items []PurchaseOrderItem, datePrefix string) (*PurchaseOrder, error)
	UpdateDraft(ctx context.Context, storeID, outletID, id, supplierID uint, items []PurchaseOrderItem) (*PurchaseOrder, error)
	DeleteDraft(ctx context.Context, storeID, outletID, id uint) error
	Order(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error)
	Receive(ctx context.Context, storeID, outletID, id uint, items []PurchaseOrderReceiptItem, actor string) (*PurchaseOrder, error)
	Cancel(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error)
}

// Suppliers adalah kontrak penyimpanan supplier yang dipakai kedua
// service slice ini (SupplierRepository lama di service/interfaces.go,
// bertipe slice).
type Suppliers interface {
	List(ctx context.Context, storeID uint) ([]*Supplier, error)
	GetByID(ctx context.Context, storeID, id uint) (*Supplier, error)
	Create(ctx context.Context, storeID uint, supplier *Supplier) (*Supplier, error)
	Update(ctx context.Context, storeID, id uint, supplier *Supplier) (*Supplier, error)
	Delete(ctx context.Context, storeID, id uint) error
}

// Products adalah kontrak katalog yang dipakai Service — hanya
// GetByIDForOutlet dari ProductRepository lama yang benar-benar
// dipakai (tipe catalog.Product).
type Products interface {
	GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*catalog.Product, error)
}

// Stores adalah kontrak toko yang dipakai Service — hanya GetTimezone
// dari StoreRepository lama yang benar-benar dipakai
// (diimplementasikan outlet.StoreRepo).
type Stores interface {
	GetTimezone(ctx context.Context, storeID uint) (string, error)
}

// SupplierService adalah SupplierService lama di service/supplier.go,
// dipindah verbatim.
type SupplierService struct {
	suppliers Suppliers
}

func NewSupplierService(suppliers Suppliers) *SupplierService {
	return &SupplierService{suppliers: suppliers}
}

func (s *SupplierService) List(ctx context.Context, storeID uint) ([]*Supplier, error) {
	return s.suppliers.List(ctx, storeID)
}

func (s *SupplierService) Get(ctx context.Context, storeID, id uint) (*Supplier, error) {
	supplier, err := s.suppliers.GetByID(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrSupplierNotFound
	}
	return supplier, err
}

func (s *SupplierService) Create(ctx context.Context, storeID uint, in SupplierInput) (*Supplier, error) {
	supplier, err := normalizeSupplier(in)
	if err != nil {
		return nil, err
	}
	return s.suppliers.Create(ctx, storeID, supplier)
}

func (s *SupplierService) Update(ctx context.Context, storeID, id uint, in SupplierInput) (*Supplier, error) {
	supplier, err := normalizeSupplier(in)
	if err != nil {
		return nil, err
	}
	updated, err := s.suppliers.Update(ctx, storeID, id, supplier)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrSupplierNotFound
	}
	return updated, err
}

func (s *SupplierService) Delete(ctx context.Context, storeID, id uint) error {
	err := s.suppliers.Delete(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return ErrSupplierNotFound
	}
	return err
}

func normalizeSupplier(in SupplierInput) (*Supplier, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return nil, ErrSupplierNameRequired
	}
	return &Supplier{
		Name:    in.Name,
		Contact: strings.TrimSpace(in.Contact),
		Address: strings.TrimSpace(in.Address),
	}, nil
}

// Service adalah PurchaseOrderService lama di
// service/purchase_order.go, dipindah verbatim.
type Service struct {
	orders    Orders
	suppliers Suppliers
	products  Products
	stores    Stores
	tx        *database.Transactor
}

func NewService(orders Orders, suppliers Suppliers, products Products, stores Stores, tx *database.Transactor) *Service {
	return &Service{orders: orders, suppliers: suppliers, products: products, stores: stores, tx: tx}
}

func (s *Service) List(ctx context.Context, storeID, outletID uint, f PurchaseOrderFilter) ([]*PurchaseOrder, error) {
	return s.orders.List(ctx, storeID, outletID, f)
}

func (s *Service) Get(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	po, err := s.orders.GetByID(ctx, storeID, outletID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrPurchaseOrderNotFound
	}
	return po, err
}

func (s *Service) Create(ctx context.Context, storeID, outletID uint, in PurchaseOrderInput) (*PurchaseOrder, error) {
	items, err := s.validate(ctx, storeID, outletID, in)
	if err != nil {
		return nil, err
	}
	location := s.storeLocation(ctx, storeID)
	datePrefix := time.Now().In(location).Format("20060102")
	return s.orders.CreateDraft(ctx, storeID, outletID, in.SupplierID, items, datePrefix)
}

func (s *Service) Update(ctx context.Context, storeID, outletID, id uint, in PurchaseOrderInput) (*PurchaseOrder, error) {
	items, err := s.validate(ctx, storeID, outletID, in)
	if err != nil {
		return nil, err
	}
	po, err := s.orders.UpdateDraft(ctx, storeID, outletID, id, in.SupplierID, items)
	return po, s.mapOrderError(err)
}

func (s *Service) Delete(ctx context.Context, storeID, outletID, id uint) error {
	return s.mapOrderError(s.orders.DeleteDraft(ctx, storeID, outletID, id))
}

func (s *Service) Order(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	po, err := s.orders.Order(ctx, storeID, outletID, id)
	return po, s.mapOrderError(err)
}

func (s *Service) Receive(ctx context.Context, storeID, outletID, id uint, items []PurchaseOrderReceiptInput, actor string) (*PurchaseOrder, error) {
	if len(items) == 0 {
		return nil, invalidPurchaseOrder("Minimal satu barang penerimaan wajib diisi.")
	}
	received := make([]PurchaseOrderReceiptItem, 0, len(items))
	seen := make(map[uint]bool, len(items))
	for _, item := range items {
		if item.ItemID == 0 || item.Qty <= 0 || seen[item.ItemID] {
			return nil, invalidPurchaseOrder("Item penerimaan harus unik dan qty harus lebih dari 0.")
		}
		seen[item.ItemID] = true
		received = append(received, PurchaseOrderReceiptItem{ItemID: item.ItemID, Qty: item.Qty})
	}
	var po *PurchaseOrder
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		po, err = s.orders.Receive(ctx, storeID, outletID, id, received, actor)
		return err
	})
	return po, s.mapOrderError(err)
}

func (s *Service) Cancel(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	po, err := s.orders.Cancel(ctx, storeID, outletID, id)
	return po, s.mapOrderError(err)
}

func (s *Service) validate(ctx context.Context, storeID, outletID uint, in PurchaseOrderInput) ([]PurchaseOrderItem, error) {
	if in.SupplierID == 0 {
		return nil, invalidPurchaseOrder("supplierId wajib diisi.")
	}
	if _, err := s.suppliers.GetByID(ctx, storeID, in.SupplierID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, invalidPurchaseOrder("Supplier tidak ditemukan di toko Anda.")
		}
		return nil, err
	}
	if len(in.Items) == 0 {
		return nil, invalidPurchaseOrder("Minimal satu barang wajib diisi.")
	}

	items := make([]PurchaseOrderItem, 0, len(in.Items))
	for _, item := range in.Items {
		if item.ProductID == 0 {
			return nil, invalidPurchaseOrder("productId wajib diisi.")
		}
		if item.Qty <= 0 {
			return nil, invalidPurchaseOrder("Qty barang harus lebih dari 0.")
		}
		if item.BuyPrice < 0 {
			return nil, invalidPurchaseOrder("Harga beli tidak boleh negatif.")
		}
		product, err := s.products.GetByIDForOutlet(ctx, storeID, outletID, item.ProductID)
		if errors.Is(err, ErrNotFound) {
			return nil, invalidPurchaseOrder("Produk tidak ditemukan di toko Anda.")
		}
		if err != nil {
			return nil, err
		}
		if item.VariantID == nil && len(product.Variants) > 0 {
			return nil, invalidPurchaseOrder("Pilih varian produk terlebih dahulu.")
		}
		if item.VariantID != nil {
			found := false
			for _, variant := range product.Variants {
				if variant.ID == *item.VariantID {
					found = true
					break
				}
			}
			if !found {
				return nil, invalidPurchaseOrder("Varian tidak ditemukan untuk produk tersebut di toko Anda.")
			}
		}
		items = append(items, PurchaseOrderItem{
			ProductID: item.ProductID, VariantID: item.VariantID,
			QtyOrdered: item.Qty, BuyPrice: item.BuyPrice,
		})
	}
	return items, nil
}

func (s *Service) storeLocation(ctx context.Context, storeID uint) *time.Location {
	timezone, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil || timezone == "" {
		timezone = "Asia/Makassar"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location, _ = time.LoadLocation("Asia/Makassar")
	}
	return location
}

func (s *Service) mapOrderError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return ErrPurchaseOrderNotFound
	case errors.Is(err, ErrPurchaseOrderItemsLocked):
		return ErrPurchaseOrderLocked
	case errors.Is(err, ErrPurchaseOrderReceiptItem):
		return invalidPurchaseOrder("Item penerimaan tidak sesuai dengan PO.")
	case errors.Is(err, ErrPurchaseOrderReceiptOver):
		return invalidPurchaseOrder("Qty diterima tidak boleh melebihi sisa pesanan.")
	case errors.Is(err, ErrPurchaseOrderReceiptTarget):
		return invalidPurchaseOrder("Produk atau varian tujuan stok sudah berubah atau dihapus; periksa katalog sebelum menerima barang.")
	default:
		return err
	}
}

func invalidPurchaseOrder(message string) error {
	return fmt.Errorf("%w: %s", ErrPurchaseOrderInvalid, message)
}
