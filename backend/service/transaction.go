package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

// ErrShiftRequired: checkout ditolak karena kasir pelaku belum punya shift
// aktif di outlet ini (keputusan Rayhan 2026-10-16: tidak ada transaksi
// tanpa shift — menggantikan gerbang lunak F2.7).
var ErrShiftRequired = errors.New("belum ada shift aktif")

type TrxService struct {
	trx      TrxRepository
	cashiers CashierRepository
	prods    ProductRepository
	shifts   ShiftRepository
	notif    *NotificationService
}

func NewTrxService(trx TrxRepository, cashiers CashierRepository, prods ProductRepository, shifts ShiftRepository, notif *NotificationService) *TrxService {
	return &TrxService{trx: trx, cashiers: cashiers, prods: prods, shifts: shifts, notif: notif}
}

type CheckoutCmd struct {
	Items        []CheckoutItemCmd
	Discount     int64
	VoucherCode  string
	Method       string
	Paid         int64
	Customer     string
	CustomerID   *uint
	PointsRedeem int64
	ClientTxID   string
}

type CheckoutItemCmd struct {
	ProductID uint
	VariantID *uint
	Qty       int
}

type PromoPreviewCmd struct {
	Items       []CheckoutItemCmd
	VoucherCode string
}

func (s *TrxService) Checkout(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, cmd CheckoutCmd) (*model.Trx, bool, error) {
	// RBAC: transaksi selalu pakai identitas dari JWT (claims), bukan dari payload.
	// fallbackName = claims.Name (JWT), actingAsCashierID = claims.ActingAsCashierID.
	// Admin tanpa switch mencatat transaksi atas namanya sendiri (cashier_id=0),
	// TANPA membuat baris cashiers kloningan. Kasir (actingAs) tetap memakai
	// baris cashiers miliknya agar scope data kasir bekerja.
	var cashierID uint
	cashierName := fallbackName

	if actingAsCashierID != nil {
		cashierID = *actingAsCashierID
		c, err := s.cashiers.GetByID(ctx, cashierID)
		if err == nil {
			cashierName = c.Name
		}
	}

	// GERBANG SHIFT: pelaku wajib punya shift aktif di outlet transaksi.
	// Berlaku juga untuk sinkronisasi offline (clientTxId): antrean FE hanya
	// bisa terbentuk setelah checkout lolos gerbang yang sama di perangkat.
	if s.shifts != nil {
		active, aerr := s.shifts.GetActiveShift(ctx, storeID, outletID, cashierID)
		if aerr != nil && !errors.Is(aerr, repo.ErrNotFound) {
			return nil, false, aerr
		}
		if active == nil {
			return nil, false, ErrShiftRequired
		}
	}

	items := make([]repo.CheckoutItem, len(cmd.Items))
	for i, it := range cmd.Items {
		items[i] = repo.CheckoutItem{ProductID: it.ProductID, VariantID: it.VariantID, Qty: it.Qty}
	}
	resTrx, created, err := s.trx.Checkout(ctx, repo.CheckoutInput{
		StoreID: storeID, OutletID: outletID, CashierID: cashierID, CashierName: cashierName,
		Items: items, Discount: cmd.Discount, Method: cmd.Method,
		VoucherCode: cmd.VoucherCode, Paid: cmd.Paid, Customer: cmd.Customer,
		CustomerID: cmd.CustomerID, PointsRedeem: cmd.PointsRedeem, ClientTxID: cmd.ClientTxID,
	})
	if err == nil && created && s.notif != nil {
		// Kekurangan offline selalu perlu diketahui admin, terlepas dari pelaku checkout.
		if resTrx.Status == model.TrxCompletedOfflineShortage {
			refIDStr := fmt.Sprintf("%d", resTrx.ID)
			title := "Transaksi offline kekurangan stok"
			message := fmt.Sprintf("Transaksi #TRX-%05d diterima, tetapi stok satu atau lebih barang tidak mencukupi.", resTrx.ID)
			_ = s.notif.Emit(ctx, storeID, title, message, model.CategoryTransaksi, "transaction_offline_shortage", fmt.Sprintf("%d", cashierID), cashierName, "transaction", refIDStr)
		} else if actingAsCashierID != nil {
			refIDStr := fmt.Sprintf("%d", resTrx.ID)
			title := "Transaksi baru"
			message := fmt.Sprintf("Rp%d · %s", resTrx.Total, resTrx.Method)
			_ = s.notif.Emit(ctx, storeID, title, message, model.CategoryTransaksi, "transaction_created", fmt.Sprintf("%d", cashierID), cashierName, "transaction", refIDStr)
		}

		s.checkLowStock(ctx, storeID, outletID, resTrx)
	}
	return resTrx, created, err
}

func (s *TrxService) PreviewPromo(ctx context.Context, storeID, outletID uint, cmd PromoPreviewCmd) (*model.PromoPreview, error) {
	items := make([]repo.CheckoutItem, len(cmd.Items))
	for i, item := range cmd.Items {
		items[i] = repo.CheckoutItem{ProductID: item.ProductID, VariantID: item.VariantID, Qty: item.Qty}
	}
	return s.trx.PreviewPromo(ctx, repo.PromoPreviewInput{StoreID: storeID, OutletID: outletID, Items: items, VoucherCode: cmd.VoucherCode})
}

func (s *TrxService) List(ctx context.Context, storeID uint, outletID *uint, cashierID uint, q, method, date, from, to string, status string, cashierFilter *uint, page, limit int) ([]*model.Trx, int, error) {
	return s.trx.List(ctx, storeID, outletID, cashierID, q, method, date, from, to, status, cashierFilter, page, limit)
}

func (s *TrxService) Get(ctx context.Context, storeID, id uint) (*model.Trx, error) {
	return s.trx.GetByID(ctx, storeID, id)
}

func (s *TrxService) Refund(ctx context.Context, storeID, outletID, trxID uint, items []repo.RefundItem, reason, byName string) (*model.Trx, error) {
	trx, err := s.trx.RefundAtOutlet(ctx, storeID, outletID, trxID, items, reason, byName)
	if err == nil && s.notif != nil {
		s.checkLowStock(ctx, storeID, outletID, trx)
		refIDStr := fmt.Sprintf("%d", trxID)
		title := "Refund"
		message := fmt.Sprintf("Invoice #TRX-%05d oleh %s", trxID, byName)
		_ = s.notif.Emit(ctx, storeID, title, message, model.CategoryTransaksi, "refund_created", "", byName, "transaction", refIDStr)
	}
	return trx, err
}

func (s *TrxService) checkLowStock(ctx context.Context, storeID, outletID uint, trx *model.Trx) {
	if trx == nil || s.notif == nil || s.prods == nil {
		return
	}
	for _, item := range trx.Items {
		product, err := s.prods.GetByIDAtOutlet(ctx, storeID, outletID, item.ProductID)
		if err != nil || product == nil {
			continue
		}
		if item.VariantID == nil {
			_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, product.ID, product.Name, "", product.Stock, product.MinStock, product.Unit, "product")
			continue
		}
		for _, variant := range product.Variants {
			if variant.ID == *item.VariantID {
				_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, variant.ID, product.Name, variant.Name, variant.Stock, variant.MinStock, product.Unit, "variant")
				break
			}
		}
	}
}
