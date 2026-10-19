package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/user"
)

// CategoryRepository adalah kontrak penyimpanan kategori yang dipakai
// Service (tanda tangan sama dengan interface lama di service/interfaces.go,
// tipe kini milik slice ini).
type CategoryRepository interface {
	ListByStore(ctx context.Context, storeID uint) ([]*Category, error)
	ProductCountsByStore(ctx context.Context, storeID uint) (map[uint]int64, error)
	ProductCountByCategory(ctx context.Context, storeID, categoryID uint) (int64, error)
	GetByID(ctx context.Context, storeID, id uint) (*Category, error)
	Create(ctx context.Context, storeID uint, name string) (*Category, error)
	Rename(ctx context.Context, storeID, id uint, name string) (*Category, error)
	Delete(ctx context.Context, storeID, id uint) (softDeleted bool, err error)
}

// ProductRepository adalah kontrak penyimpanan produk yang dipakai
// Service (tanda tangan sama dengan interface lama di service/interfaces.go,
// tipe kini milik slice ini).
type ProductRepository interface {
	GetByIDForStore(ctx context.Context, storeID, id uint) (*Product, error)
	GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*Product, error)
	ListForStore(ctx context.Context, storeID uint, f ProductFilter) (*ProductPage, error)
	ListForOutlet(ctx context.Context, storeID, outletID uint, f ProductFilter) (*ProductPage, error)
	ListForExport(ctx context.Context, storeID uint) ([]*Product, error)
	ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*Product, error)
	DistinctUnits(ctx context.Context, storeID uint) ([]string, error)
	InsertProduct(ctx context.Context, storeID, outletID uint, p *Product) error
	DefaultOutlet(ctx context.Context, storeID uint) (*outlet.Outlet, error)
	Update(ctx context.Context, storeID, id uint, p *Product) (*Product, error)
	SetActive(ctx context.Context, storeID, id uint, active bool) error
	IsSkuTaken(ctx context.Context, storeID uint, sku string, exceptProductID uint) (bool, error)
	ListVariantsForStore(ctx context.Context, storeID, productID uint) ([]ProductVariant, error)
	ListVariantsForOutlet(ctx context.Context, storeID, outletID, productID uint) ([]ProductVariant, error)
	GetVariant(ctx context.Context, storeID, productID, variantID uint) (*ProductVariant, error)
	GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*ProductVariant, error)
	InsertVariant(ctx context.Context, storeID, outletID, productID uint, variant *ProductVariant) error
	UpdateVariantFields(ctx context.Context, storeID, productID, variantID uint, variant *ProductVariant) error
	DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error
	IsVariantSkuTaken(ctx context.Context, storeID uint, sku string, exceptID uint) (bool, error)
	ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]PriceTier, error)
	CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tier *PriceTier) (*PriceTier, error)
	UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, tier *PriceTier) (*PriceTier, error)
	DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error
	Delete(ctx context.Context, storeID, id uint) error
}

// Notifier adalah potongan NotificationService lama yang dipakai Service:
// pemberitahuan stok rendah (kesalahan dibuang pemanggil, seperti semula).
type Notifier interface {
	CheckAndNotifyLowStock(ctx context.Context, storeID, itemID uint, productName, variantName string, currentStock, minStock int, unit, referenceType string) error
}

// StockAdjuster adalah potongan InventoryService lama yang dipakai alur
// impor CSV untuk mencatat stok awal produk impor.
type StockAdjuster interface {
	AdjustStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, actor, direction string, qty int64, reason string) (*Product, error)
}

type Service struct {
	cats  CategoryRepository
	prods ProductRepository
	notif Notifier
	inv   StockAdjuster
	stock StockEngine
	tx    Transactor
}

func NewService(cats CategoryRepository, prods ProductRepository, notif Notifier) *Service {
	return &Service{cats: cats, prods: prods, notif: notif}
}

// SetInventory memasang StockAdjuster yang dipakai alur impor CSV untuk
// mencatat stok awal (pola setter yang sama dengan SetOutlets di AuthService).
func (s *Service) SetInventory(inv StockAdjuster) { s.inv = inv }

// ListProductUnits: satuan yang sedang dipakai produk toko ini —
// sumber daftar satuan yang mengelola dirinya sendiri di form produk.
func (s *Service) ListProductUnits(ctx context.Context, storeID uint) ([]string, error) {
	return s.prods.DistinctUnits(ctx, storeID)
}

func (s *Service) ListCategories(ctx context.Context, storeID uint) ([]*Category, error) {
	categories, err := s.cats.ListByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	counts, err := s.cats.ProductCountsByStore(ctx, storeID)
	if err != nil {
		return nil, err
	}
	for _, category := range categories {
		category.ProductCount = int(counts[category.ID])
	}
	return categories, nil
}

func (s *Service) CreateCategory(ctx context.Context, storeID uint, name string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("nama kategori wajib diisi")
	}
	c, err := s.cats.Create(ctx, storeID, name)
	if err != nil {
		if errors.Is(err, ErrDuplicate) {
			return nil, ErrCategoryTaken
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) RenameCategory(ctx context.Context, storeID, id uint, name string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("nama kategori wajib diisi")
	}
	category, err := s.cats.Rename(ctx, storeID, id, name)
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrCategoryTaken
	}
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	if err != nil {
		return nil, err
	}
	count, err := s.cats.ProductCountByCategory(ctx, storeID, id)
	if err != nil {
		return nil, err
	}
	category.ProductCount = int(count)
	return category, nil
}

func (s *Service) DeleteCategory(ctx context.Context, storeID, id uint) (soft bool, err error) {
	soft, err = s.cats.Delete(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return false, user.ErrStoreMismatch
	}
	return soft, err
}

type ProductInput struct {
	Name       string
	SKU        string
	Barcode    string
	CategoryID *uint
	BuyPrice   int64
	SellPrice  int64
	Stock      int
	MinStock   *int
	Unit       string
}

func normalizeProductInput(in *ProductInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.TrimSpace(in.SKU)
	in.Barcode = strings.TrimSpace(in.Barcode)
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Unit == "" {
		in.Unit = "pcs"
	}
	if in.Name == "" {
		return fmt.Errorf("nama produk wajib diisi")
	}
	if in.SKU == "" {
		return fmt.Errorf("SKU wajib diisi")
	}
	if in.BuyPrice < 0 || in.SellPrice < 0 {
		return fmt.Errorf("harga tidak boleh negatif")
	}
	if in.Stock < 0 {
		return fmt.Errorf("stok tidak boleh negatif")
	}
	if in.MinStock != nil && *in.MinStock < 0 {
		return fmt.Errorf("batas stok minimum tidak boleh negatif")
	}
	return nil
}

func (s *Service) CreateProduct(ctx context.Context, storeID, outletID uint, actorName string, in ProductInput) (*Product, error) {
	if err := normalizeProductInput(&in); err != nil {
		return nil, err
	}
	if in.CategoryID != nil && *in.CategoryID != 0 {
		if _, err := s.cats.GetByID(ctx, storeID, *in.CategoryID); err != nil {
			return nil, ErrCategoryInvalid
		}
	} else {
		in.CategoryID = nil
	}
	taken, err := s.prods.IsSkuTaken(ctx, storeID, in.SKU, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrSkuTaken
	}

	minStock := 0
	if in.MinStock != nil {
		minStock = *in.MinStock
	}
	p := &Product{
		CategoryID: in.CategoryID,
		Name:       in.Name,
		SKU:        in.SKU,
		Barcode:    in.Barcode,
		BuyPrice:   in.BuyPrice,
		SellPrice:  in.SellPrice,
		Stock:      in.Stock,
		MinStock:   minStock,
		Unit:       in.Unit,
		Active:     true,
	}
	out, err := s.createProductWithStock(ctx, storeID, outletID, p, actorName)
	if err != nil {
		if errors.Is(err, ErrDuplicate) {
			return nil, ErrSkuTaken
		}
		return nil, err
	}
	if s.notif != nil {
		_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, out.ID, out.Name, "", out.Stock, out.MinStock, out.Unit, "product")
	}
	return out, nil
}

func (s *Service) UpdateProduct(ctx context.Context, storeID, outletID, id uint, in ProductInput) (*Product, error) {
	if err := normalizeProductInput(&in); err != nil {
		return nil, err
	}
	if in.CategoryID != nil && *in.CategoryID != 0 {
		if _, err := s.cats.GetByID(ctx, storeID, *in.CategoryID); err != nil {
			return nil, ErrCategoryInvalid
		}
	} else {
		in.CategoryID = nil
	}

	taken, err := s.prods.IsSkuTaken(ctx, storeID, in.SKU, id)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrSkuTaken
	}

	minStock := -1
	if in.MinStock != nil {
		minStock = *in.MinStock
	}
	p := &Product{
		CategoryID: in.CategoryID,
		Name:       in.Name,
		SKU:        in.SKU,
		Barcode:    in.Barcode,
		BuyPrice:   in.BuyPrice,
		SellPrice:  in.SellPrice,
		MinStock:   minStock,
		Unit:       in.Unit,
	}
	out, err := s.prods.Update(ctx, storeID, id, p)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, user.ErrStoreMismatch
		}
		if errors.Is(err, ErrDuplicate) {
			return nil, ErrSkuTaken
		}
		return nil, err
	}
	updated, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, out.ID)
	if err != nil {
		return nil, err
	}
	if s.notif != nil {
		_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, updated.ID, updated.Name, "", updated.Stock, updated.MinStock, updated.Unit, "product")
	}
	return updated, nil
}

type VariantInput struct {
	Name      string
	SKU       string
	BuyPrice  int64
	SellPrice int64
	Stock     int
	MinStock  int
}

func normalizeVariantInput(in *VariantInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.SKU = strings.TrimSpace(in.SKU)
	if in.Name == "" || in.SKU == "" {
		return fmt.Errorf("nama varian dan SKU wajib diisi")
	}
	if in.BuyPrice < 0 || in.SellPrice < 0 || in.Stock < 0 || in.MinStock < 0 {
		return fmt.Errorf("harga dan stok varian tidak boleh negatif")
	}
	return nil
}

func (s *Service) ListVariants(ctx context.Context, storeID, outletID, productID uint) ([]ProductVariant, error) {
	variants, err := s.prods.ListVariantsForOutlet(ctx, storeID, outletID, productID)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	return variants, err
}

func (s *Service) CreateVariant(ctx context.Context, storeID, outletID, productID uint, actor string, in VariantInput) (*ProductVariant, error) {
	if err := normalizeVariantInput(&in); err != nil {
		return nil, err
	}
	taken, err := s.prods.IsVariantSkuTaken(ctx, storeID, in.SKU, 0)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrSkuTaken
	}
	variant := &ProductVariant{Name: in.Name, SKU: in.SKU, BuyPrice: in.BuyPrice, SellPrice: in.SellPrice, Stock: in.Stock, MinStock: in.MinStock}
	created, err := s.createVariantWithStock(ctx, storeID, outletID, productID, variant, actor)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrSkuTaken
	}
	return created, err
}

func (s *Service) UpdateVariant(ctx context.Context, storeID, outletID, productID, variantID uint, actor string, in VariantInput) (*ProductVariant, error) {
	if err := normalizeVariantInput(&in); err != nil {
		return nil, err
	}
	taken, err := s.prods.IsVariantSkuTaken(ctx, storeID, in.SKU, variantID)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrSkuTaken
	}
	variant := &ProductVariant{Name: in.Name, SKU: in.SKU, BuyPrice: in.BuyPrice, SellPrice: in.SellPrice, Stock: in.Stock, MinStock: in.MinStock}
	updated, err := s.updateVariantWithStock(ctx, storeID, outletID, productID, variantID, variant, actor)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	if errors.Is(err, ErrDuplicate) {
		return nil, ErrSkuTaken
	}
	if err == nil && s.notif != nil {
		if product, getErr := s.prods.GetByIDForOutlet(ctx, storeID, outletID, productID); getErr == nil {
			_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, updated.ID, product.Name, updated.Name, updated.Stock, updated.MinStock, product.Unit, "variant")
		}
	}
	return updated, err
}

func (s *Service) DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error {
	err := s.prods.DeleteVariant(ctx, storeID, productID, variantID)
	if errors.Is(err, ErrNotFound) {
		return user.ErrStoreMismatch
	}
	return err
}

type PriceTierInput struct {
	MinQty    int
	SellPrice int64
}

func validatePriceTier(in PriceTierInput) error {
	if in.MinQty < 1 || in.SellPrice < 0 {
		return fmt.Errorf("jumlah minimum harus minimal 1 dan harga tidak boleh negatif")
	}
	return nil
}

func (s *Service) ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]PriceTier, error) {
	tiers, err := s.prods.ListPriceTiers(ctx, storeID, productID, variantID)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	return tiers, err
}

func (s *Service) CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, in PriceTierInput) (*PriceTier, error) {
	if err := validatePriceTier(in); err != nil {
		return nil, err
	}
	tier := &PriceTier{MinQty: in.MinQty, SellPrice: in.SellPrice}
	created, err := s.prods.CreatePriceTier(ctx, storeID, productID, variantID, tier)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	return created, err
}

func (s *Service) UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, in PriceTierInput) (*PriceTier, error) {
	if err := validatePriceTier(in); err != nil {
		return nil, err
	}
	tier := &PriceTier{MinQty: in.MinQty, SellPrice: in.SellPrice}
	updated, err := s.prods.UpdatePriceTier(ctx, storeID, productID, variantID, tierID, tier)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	return updated, err
}

func (s *Service) DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error {
	err := s.prods.DeletePriceTier(ctx, storeID, productID, variantID, tierID)
	if errors.Is(err, ErrNotFound) {
		return user.ErrStoreMismatch
	}
	return err
}

func (s *Service) ListProducts(ctx context.Context, storeID, outletID uint, f ProductFilter) (*ProductPage, error) {
	return s.prods.ListForOutlet(ctx, storeID, outletID, f)
}

func (s *Service) GetProduct(ctx context.Context, storeID, outletID, id uint) (*Product, error) {
	p, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, user.ErrStoreMismatch
	}
	return p, err
}

func (s *Service) SetProductActive(ctx context.Context, storeID, id uint, active bool) error {
	err := s.prods.SetActive(ctx, storeID, id, active)
	if errors.Is(err, ErrNotFound) {
		return user.ErrStoreMismatch
	}
	return err
}

func (s *Service) DeleteProduct(ctx context.Context, storeID, id uint) error {
	err := s.prods.Delete(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return user.ErrStoreMismatch
	}
	return err
}
