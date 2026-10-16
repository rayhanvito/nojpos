package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrCategoryInvalid = errors.New("kategori tidak ditemukan di toko Anda")
	ErrCategoryTaken   = errors.New("kategori dengan nama itu sudah ada")
	ErrSkuTaken        = errors.New("SKU sudah digunakan di toko ini")
)

type CatalogService struct {
	cats  CategoryRepository
	prods ProductRepository
	movs  MovementRepository
	notif *NotificationService
}

func NewCatalogService(cats CategoryRepository, prods ProductRepository, movs MovementRepository, notif *NotificationService) *CatalogService {
	return &CatalogService{cats: cats, prods: prods, movs: movs, notif: notif}
}

func (s *CatalogService) ListCategories(ctx context.Context, storeID uint) ([]*model.Category, error) {
	return s.cats.ListByStore(ctx, storeID)
}

func (s *CatalogService) CreateCategory(ctx context.Context, storeID uint, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("nama kategori wajib diisi")
	}
	c, err := s.cats.Create(ctx, storeID, name)
	if err != nil {
		if errors.Is(err, repo.ErrDuplicate) {
			return nil, ErrCategoryTaken
		}
		return nil, err
	}
	return c, nil
}

func (s *CatalogService) DeleteCategory(ctx context.Context, storeID, id uint) (soft bool, err error) {
	soft, err = s.cats.Delete(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return false, ErrStoreMismatch
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

func (s *CatalogService) CreateProduct(ctx context.Context, storeID, outletID uint, actorName string, in ProductInput) (*model.Product, error) {
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
	p := &model.Product{
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
	out, err := s.prods.CreateWithInitialMovementAtOutlet(ctx, storeID, outletID, p, actorName)
	if err != nil {
		if errors.Is(err, repo.ErrDuplicate) {
			return nil, ErrSkuTaken
		}
		return nil, err
	}
	return out, nil
}

func (s *CatalogService) UpdateProduct(ctx context.Context, storeID, outletID, id uint, in ProductInput) (*model.Product, error) {
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
	p := &model.Product{
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
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrStoreMismatch
		}
		if errors.Is(err, repo.ErrDuplicate) {
			return nil, ErrSkuTaken
		}
		return nil, err
	}
	return s.prods.GetByIDAtOutlet(ctx, storeID, outletID, out.ID)
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

func (s *CatalogService) ListVariants(ctx context.Context, storeID, outletID, productID uint) ([]model.ProductVariant, error) {
	variants, err := s.prods.ListVariantsAtOutlet(ctx, storeID, outletID, productID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	return variants, err
}

func (s *CatalogService) CreateVariant(ctx context.Context, storeID, outletID, productID uint, actor string, in VariantInput) (*model.ProductVariant, error) {
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
	variant := &model.ProductVariant{Name: in.Name, SKU: in.SKU, BuyPrice: in.BuyPrice, SellPrice: in.SellPrice, Stock: in.Stock, MinStock: in.MinStock}
	created, err := s.prods.CreateVariantAtOutlet(ctx, storeID, outletID, productID, variant, actor)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrSkuTaken
	}
	return created, err
}

func (s *CatalogService) UpdateVariant(ctx context.Context, storeID, outletID, productID, variantID uint, actor string, in VariantInput) (*model.ProductVariant, error) {
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
	variant := &model.ProductVariant{Name: in.Name, SKU: in.SKU, BuyPrice: in.BuyPrice, SellPrice: in.SellPrice, Stock: in.Stock, MinStock: in.MinStock}
	updated, err := s.prods.UpdateVariantAtOutlet(ctx, storeID, outletID, productID, variantID, variant, actor)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	if errors.Is(err, repo.ErrDuplicate) {
		return nil, ErrSkuTaken
	}
	if err == nil && s.notif != nil {
		if product, getErr := s.prods.GetByIDAtOutlet(ctx, storeID, outletID, productID); getErr == nil {
			_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, updated.ID, product.Name, updated.Name, updated.Stock, updated.MinStock, product.Unit, "variant")
		}
	}
	return updated, err
}

func (s *CatalogService) DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error {
	err := s.prods.DeleteVariant(ctx, storeID, productID, variantID)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrStoreMismatch
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

func (s *CatalogService) ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]model.PriceTier, error) {
	tiers, err := s.prods.ListPriceTiers(ctx, storeID, productID, variantID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	return tiers, err
}

func (s *CatalogService) CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, in PriceTierInput) (*model.PriceTier, error) {
	if err := validatePriceTier(in); err != nil {
		return nil, err
	}
	tier := &model.PriceTier{MinQty: in.MinQty, SellPrice: in.SellPrice}
	created, err := s.prods.CreatePriceTier(ctx, storeID, productID, variantID, tier)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	return created, err
}

func (s *CatalogService) UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, in PriceTierInput) (*model.PriceTier, error) {
	if err := validatePriceTier(in); err != nil {
		return nil, err
	}
	tier := &model.PriceTier{MinQty: in.MinQty, SellPrice: in.SellPrice}
	updated, err := s.prods.UpdatePriceTier(ctx, storeID, productID, variantID, tierID, tier)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	return updated, err
}

func (s *CatalogService) DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error {
	err := s.prods.DeletePriceTier(ctx, storeID, productID, variantID, tierID)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrStoreMismatch
	}
	return err
}

func (s *CatalogService) ListProducts(ctx context.Context, storeID, outletID uint, f repo.ProductFilter) (*repo.ProductPage, error) {
	return s.prods.ListAtOutlet(ctx, storeID, outletID, f)
}

func (s *CatalogService) GetProduct(ctx context.Context, storeID, outletID, id uint) (*model.Product, error) {
	p, err := s.prods.GetByIDAtOutlet(ctx, storeID, outletID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrStoreMismatch
	}
	return p, err
}

func (s *CatalogService) SetProductActive(ctx context.Context, storeID, id uint, active bool) error {
	err := s.prods.SetActive(ctx, storeID, id, active)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrStoreMismatch
	}
	return err
}

var ErrBadDirection = errors.New("arah penyesuaian tidak valid")

func (s *CatalogService) AdjustStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, actor, direction string, qty int64, reason string) (*model.Product, error) {
	reason = strings.TrimSpace(reason)
	if direction != "plus" && direction != "minus" {
		return nil, ErrBadDirection
	}
	if qty < 1 {
		return nil, fmt.Errorf("jumlah harus minimal 1")
	}
	if reason == "" {
		return nil, fmt.Errorf("alasan penyesuaian wajib diisi")
	}
	delta := int(qty)
	if direction == "minus" {
		delta = -int(qty)
	}
	if variantID != nil {
		variant, err := s.prods.AdjustVariantStockAtOutlet(ctx, storeID, outletID, productID, *variantID, delta, reason, actor)
		if err != nil {
			switch {
			case errors.Is(err, repo.ErrNotFound):
				return nil, ErrStoreMismatch
			case errors.Is(err, repo.ErrStockInsufficient):
				return nil, fmt.Errorf("stok tidak boleh negatif")
			}
			return nil, err
		}
		product, err := s.prods.GetByIDAtOutlet(ctx, storeID, outletID, productID)
		if err != nil {
			return nil, err
		}
		if s.notif != nil {
			_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, variant.ID, product.Name, variant.Name, variant.Stock, variant.MinStock, product.Unit, "variant")
		}
		return product, nil
	}

	p, err := s.prods.AdjustStockAtOutlet(ctx, storeID, outletID, productID, delta, reason, actor)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			return nil, ErrStoreMismatch
		case errors.Is(err, repo.ErrStockInsufficient):
			return nil, fmt.Errorf("stok tidak boleh negatif")
		}
		return nil, err
	}
	if p != nil && s.notif != nil {
		_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, p.ID, p.Name, "", p.Stock, p.MinStock, p.Unit, "product")
	}
	return p, nil
}

func (s *CatalogService) ListMovements(ctx context.Context, storeID, outletID uint, f repo.MovementFilter) (*repo.MovementPage, error) {
	return s.movs.ListAtOutlet(ctx, storeID, outletID, f)
}

func (s *CatalogService) DeleteProduct(ctx context.Context, storeID, id uint) error {
	err := s.prods.Delete(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrStoreMismatch
	}
	return err
}
