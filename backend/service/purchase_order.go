package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

var (
	ErrPurchaseOrderNotFound = errors.New("purchase order tidak ditemukan")
	ErrPurchaseOrderInvalid  = errors.New("data purchase order tidak valid")
	ErrPurchaseOrderLocked   = errors.New("barang pada PO yang sudah dipesan tidak dapat diubah")
)

type PurchaseOrderItemInput struct {
	ProductID uint
	VariantID *uint
	Qty       int
	BuyPrice  int64
}

type PurchaseOrderInput struct {
	SupplierID uint
	Items      []PurchaseOrderItemInput
}

type PurchaseOrderReceiptInput struct {
	ItemID uint
	Qty    int
}

type PurchaseOrderService struct {
	orders    PurchaseOrderRepository
	suppliers SupplierRepository
	products  ProductRepository
	stores    StoreRepository
}

func NewPurchaseOrderService(orders PurchaseOrderRepository, suppliers SupplierRepository, products ProductRepository, stores StoreRepository) *PurchaseOrderService {
	return &PurchaseOrderService{orders: orders, suppliers: suppliers, products: products, stores: stores}
}

func (s *PurchaseOrderService) List(ctx context.Context, storeID, outletID uint, f repo.PurchaseOrderFilter) ([]*model.PurchaseOrder, error) {
	return s.orders.List(ctx, storeID, outletID, f)
}

func (s *PurchaseOrderService) Get(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	po, err := s.orders.GetByID(ctx, storeID, outletID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrPurchaseOrderNotFound
	}
	return po, err
}

func (s *PurchaseOrderService) Create(ctx context.Context, storeID, outletID uint, in PurchaseOrderInput) (*model.PurchaseOrder, error) {
	items, err := s.validate(ctx, storeID, outletID, in)
	if err != nil {
		return nil, err
	}
	location := s.storeLocation(ctx, storeID)
	datePrefix := time.Now().In(location).Format("20060102")
	return s.orders.CreateDraft(ctx, storeID, outletID, in.SupplierID, items, datePrefix)
}

func (s *PurchaseOrderService) Update(ctx context.Context, storeID, outletID, id uint, in PurchaseOrderInput) (*model.PurchaseOrder, error) {
	items, err := s.validate(ctx, storeID, outletID, in)
	if err != nil {
		return nil, err
	}
	po, err := s.orders.UpdateDraft(ctx, storeID, outletID, id, in.SupplierID, items)
	return po, s.mapOrderError(err)
}

func (s *PurchaseOrderService) Delete(ctx context.Context, storeID, outletID, id uint) error {
	return s.mapOrderError(s.orders.DeleteDraft(ctx, storeID, outletID, id))
}

func (s *PurchaseOrderService) Order(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	po, err := s.orders.Order(ctx, storeID, outletID, id)
	return po, s.mapOrderError(err)
}

func (s *PurchaseOrderService) Receive(ctx context.Context, storeID, outletID, id uint, items []PurchaseOrderReceiptInput, actor string) (*model.PurchaseOrder, error) {
	if len(items) == 0 {
		return nil, invalidPurchaseOrder("Minimal satu barang penerimaan wajib diisi.")
	}
	received := make([]repo.PurchaseOrderReceiptItem, 0, len(items))
	seen := make(map[uint]bool, len(items))
	for _, item := range items {
		if item.ItemID == 0 || item.Qty <= 0 || seen[item.ItemID] {
			return nil, invalidPurchaseOrder("Item penerimaan harus unik dan qty harus lebih dari 0.")
		}
		seen[item.ItemID] = true
		received = append(received, repo.PurchaseOrderReceiptItem{ItemID: item.ItemID, Qty: item.Qty})
	}
	po, err := s.orders.Receive(ctx, storeID, outletID, id, received, actor)
	return po, s.mapOrderError(err)
}

func (s *PurchaseOrderService) Cancel(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	po, err := s.orders.Cancel(ctx, storeID, outletID, id)
	return po, s.mapOrderError(err)
}

func (s *PurchaseOrderService) validate(ctx context.Context, storeID, outletID uint, in PurchaseOrderInput) ([]model.PurchaseOrderItem, error) {
	if in.SupplierID == 0 {
		return nil, invalidPurchaseOrder("supplierId wajib diisi.")
	}
	if _, err := s.suppliers.GetByID(ctx, storeID, in.SupplierID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, invalidPurchaseOrder("Supplier tidak ditemukan di toko Anda.")
		}
		return nil, err
	}
	if len(in.Items) == 0 {
		return nil, invalidPurchaseOrder("Minimal satu barang wajib diisi.")
	}

	items := make([]model.PurchaseOrderItem, 0, len(in.Items))
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
		product, err := s.products.GetByIDAtOutlet(ctx, storeID, outletID, item.ProductID)
		if errors.Is(err, repo.ErrNotFound) {
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
		items = append(items, model.PurchaseOrderItem{
			ProductID: item.ProductID, VariantID: item.VariantID,
			QtyOrdered: item.Qty, BuyPrice: item.BuyPrice,
		})
	}
	return items, nil
}

func (s *PurchaseOrderService) storeLocation(ctx context.Context, storeID uint) *time.Location {
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

func (s *PurchaseOrderService) mapOrderError(err error) error {
	switch {
	case errors.Is(err, repo.ErrNotFound):
		return ErrPurchaseOrderNotFound
	case errors.Is(err, repo.ErrPurchaseOrderItemsLocked):
		return ErrPurchaseOrderLocked
	case errors.Is(err, repo.ErrPurchaseOrderReceiptItem):
		return invalidPurchaseOrder("Item penerimaan tidak sesuai dengan PO.")
	case errors.Is(err, repo.ErrPurchaseOrderReceiptOver):
		return invalidPurchaseOrder("Qty diterima tidak boleh melebihi sisa pesanan.")
	case errors.Is(err, repo.ErrPurchaseOrderReceiptTarget):
		return invalidPurchaseOrder("Produk atau varian tujuan stok sudah berubah atau dihapus; periksa katalog sebelum menerima barang.")
	default:
		return err
	}
}

func invalidPurchaseOrder(message string) error {
	return fmt.Errorf("%w: %s", ErrPurchaseOrderInvalid, message)
}
