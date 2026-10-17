package sales

import (
	"context"
	"errors"
	"fmt"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/notification"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/promo"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// Transactions adalah kontrak penyimpanan transaksi yang dipakai
// Service (TrxRepository lama di service/interfaces.go, bertipe slice).
type Transactions interface {
	Checkout(ctx context.Context, in CheckoutInput) (*Trx, bool, error)
	PreviewPromo(ctx context.Context, in PromoPreviewInput) (*promo.Preview, error)
	Refund(ctx context.Context, storeID, trxID uint, items []RefundItem, reason, byName string) (*Trx, error)
	RefundAtOutlet(ctx context.Context, storeID, outletID, trxID uint, items []RefundItem, reason, byName string) (*Trx, error)
	List(ctx context.Context, storeID uint, outletID *uint, cashierID uint, q, method, date, from, to string, status string, cashierFilter *uint, page, limit int) ([]*Trx, int, error)
	ListForExport(ctx context.Context, storeID uint, period, timezone, date, from, to string) ([]*Trx, error)
	GetByID(ctx context.Context, storeID, id uint) (*Trx, error)
}

// Cashiers adalah kontrak kasir yang dipakai Service — hanya
// GetByIDUnscoped dari CashierRepository lama yang benar-benar
// dipakai (tipe staff.Cashier).
type Cashiers interface {
	GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error)
}

// Products adalah kontrak katalog yang dipakai Service — hanya
// GetByIDForOutlet dari ProductRepository lama yang benar-benar
// dipakai (tipe catalog.Product).
type Products interface {
	GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*catalog.Product, error)
}

// Shifts adalah kontrak shift yang dipakai Service — hanya
// GetActiveShift dari ShiftRepository lama yang benar-benar dipakai
// (tipe shift.CashierShift).
type Shifts interface {
	GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*shift.CashierShift, error)
}

// Notifier adalah potongan notification.Service yang dipakai Service
// (struktural): Emit untuk notifikasi transaksi/refund dan
// CheckAndNotifyLowStock untuk stok rendah sesudah checkout/refund.
type Notifier interface {
	Emit(ctx context.Context, storeID uint, title, message string, category notification.NotificationCategory, notifType string, actorID, actorName, refType, refID string) error
	CheckAndNotifyLowStock(ctx context.Context, storeID, itemID uint, productName, variantName string, currentStock, minStock int, unit, referenceType string) error
}

// Service adalah TrxService lama di service/transaction.go, dipindah
// verbatim; orkestrasi transaksi checkout/refund kini di sini —
// pemanggilan repo dibungkus tx.WithinTx dan repo ikut bergabung
// (JOIN) ke transaksi itu.
type Service struct {
	transactions Transactions
	cashiers     Cashiers
	prods        Products
	shifts       Shifts
	notif        Notifier
	tx           *database.Transactor
}

func NewService(transactions Transactions, cashiers Cashiers, prods Products, shifts Shifts, notif Notifier, tx *database.Transactor) *Service {
	return &Service{transactions: transactions, cashiers: cashiers, prods: prods, shifts: shifts, notif: notif, tx: tx}
}

func (s *Service) Checkout(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, cmd CheckoutCmd) (*Trx, bool, error) {
	// RBAC: transaksi selalu pakai identitas dari JWT (claims), bukan dari payload.
	// fallbackName = claims.Name (JWT), actingAsCashierID = claims.ActingAsCashierID.
	// Admin tanpa switch mencatat transaksi atas namanya sendiri (cashier_id=0),
	// TANPA membuat baris cashiers kloningan. Kasir (actingAs) tetap memakai
	// baris cashiers miliknya agar scope data kasir bekerja.
	var cashierID uint
	cashierName := fallbackName

	if actingAsCashierID != nil {
		cashierID = *actingAsCashierID
		c, err := s.cashiers.GetByIDUnscoped(ctx, cashierID)
		if err == nil {
			cashierName = c.Name
		}
	}

	// GERBANG SHIFT: pelaku wajib punya shift aktif di outlet transaksi.
	// Berlaku juga untuk sinkronisasi offline (clientTxId): antrean FE hanya
	// bisa terbentuk setelah checkout lolos gerbang yang sama di perangkat.
	if s.shifts != nil {
		active, aerr := s.shifts.GetActiveShift(ctx, storeID, outletID, cashierID)
		if aerr != nil && !errors.Is(aerr, outlet.ErrNotFound) {
			return nil, false, aerr
		}
		if active == nil {
			return nil, false, ErrShiftRequired
		}
	}

	items := make([]CheckoutItem, len(cmd.Items))
	for i, it := range cmd.Items {
		items[i] = CheckoutItem{ProductID: it.ProductID, VariantID: it.VariantID, Qty: it.Qty}
	}
	var resTrx *Trx
	var created bool
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		resTrx, created, err = s.transactions.Checkout(ctx, CheckoutInput{
			StoreID: storeID, OutletID: outletID, CashierID: cashierID, CashierName: cashierName,
			Items: items, Discount: cmd.Discount, Method: cmd.Method,
			VoucherCode: cmd.VoucherCode, Paid: cmd.Paid, Customer: cmd.Customer,
			CustomerID: cmd.CustomerID, PointsRedeem: cmd.PointsRedeem, ClientTxID: cmd.ClientTxID,
		})
		return err
	})
	if err == nil && created && s.notif != nil {
		// Kekurangan offline selalu perlu diketahui admin, terlepas dari pelaku checkout.
		if resTrx.Status == TrxCompletedOfflineShortage {
			refIDStr := fmt.Sprintf("%d", resTrx.ID)
			title := "Transaksi offline kekurangan stok"
			message := fmt.Sprintf("Transaksi #TRX-%05d diterima, tetapi stok satu atau lebih barang tidak mencukupi.", resTrx.ID)
			_ = s.notif.Emit(ctx, storeID, title, message, notification.CategoryTransaksi, "transaction_offline_shortage", fmt.Sprintf("%d", cashierID), cashierName, "transaction", refIDStr)
		} else if actingAsCashierID != nil {
			refIDStr := fmt.Sprintf("%d", resTrx.ID)
			title := "Transaksi baru"
			message := fmt.Sprintf("Rp%d · %s", resTrx.Total, resTrx.Method)
			_ = s.notif.Emit(ctx, storeID, title, message, notification.CategoryTransaksi, "transaction_created", fmt.Sprintf("%d", cashierID), cashierName, "transaction", refIDStr)
		}

		s.checkLowStock(ctx, storeID, outletID, resTrx)
	}
	return resTrx, created, err
}

func (s *Service) PreviewPromo(ctx context.Context, storeID, outletID uint, cmd PromoPreviewCmd) (*promo.Preview, error) {
	items := make([]CheckoutItem, len(cmd.Items))
	for i, item := range cmd.Items {
		items[i] = CheckoutItem{ProductID: item.ProductID, VariantID: item.VariantID, Qty: item.Qty}
	}
	return s.transactions.PreviewPromo(ctx, PromoPreviewInput{StoreID: storeID, OutletID: outletID, Items: items, VoucherCode: cmd.VoucherCode})
}

func (s *Service) List(ctx context.Context, storeID uint, outletID *uint, cashierID uint, q, method, date, from, to string, status string, cashierFilter *uint, page, limit int) ([]*Trx, int, error) {
	return s.transactions.List(ctx, storeID, outletID, cashierID, q, method, date, from, to, status, cashierFilter, page, limit)
}

func (s *Service) Get(ctx context.Context, storeID, id uint) (*Trx, error) {
	return s.transactions.GetByID(ctx, storeID, id)
}

func (s *Service) Refund(ctx context.Context, storeID, outletID, trxID uint, items []RefundItem, reason, byName string) (*Trx, error) {
	var trx *Trx
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		trx, err = s.transactions.RefundAtOutlet(ctx, storeID, outletID, trxID, items, reason, byName)
		return err
	})
	if err == nil && s.notif != nil {
		s.checkLowStock(ctx, storeID, outletID, trx)
		refIDStr := fmt.Sprintf("%d", trxID)
		title := "Refund"
		message := fmt.Sprintf("Invoice #TRX-%05d oleh %s", trxID, byName)
		_ = s.notif.Emit(ctx, storeID, title, message, notification.CategoryTransaksi, "refund_created", "", byName, "transaction", refIDStr)
	}
	return trx, err
}

func (s *Service) checkLowStock(ctx context.Context, storeID, outletID uint, trx *Trx) {
	if trx == nil || s.notif == nil || s.prods == nil {
		return
	}
	for _, item := range trx.Items {
		product, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, item.ProductID)
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
