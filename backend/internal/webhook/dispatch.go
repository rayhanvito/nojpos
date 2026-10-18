package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TransactionData adalah ringkasan transaksi pada payload event
// transaction.created / transaction.refunded.
type TransactionData struct {
	TransactionID uint   `json:"transaction_id"`
	Number        string `json:"number"`
	Status        string `json:"status"`
	Total         int64  `json:"total"`
	Method        string `json:"method"`
	OutletID      *uint  `json:"outlet_id,omitempty"`
	CashierName   string `json:"cashier_name"`
}

// StockData adalah ringkasan perubahan stok pada payload event
// stock.updated. Kolom rujukan sumber (opname_session_id / transfer_id
// + from/to_outlet_id) hanya terisi untuk sumbernya masing-masing.
type StockData struct {
	Source          string `json:"source"` // adjustment | opname | transfer
	ProductID       uint   `json:"product_id"`
	VariantID       *uint  `json:"variant_id,omitempty"`
	QtyDelta        int    `json:"qty_delta"`
	OutletID        uint   `json:"outlet_id"`
	Reason          string `json:"reason,omitempty"`
	OpnameSessionID uint   `json:"opname_session_id,omitempty"`
	TransferID      uint   `json:"transfer_id,omitempty"`
	FromOutletID    uint   `json:"from_outlet_id,omitempty"`
	ToOutletID      uint   `json:"to_outlet_id,omitempty"`
}

// envelope adalah badan JSON yang dikirim ke penerima. Tanda tangan
// HMAC dihitung dari badan ini persis seperti terkirim.
type envelope struct {
	ID        string          `json:"id"`
	Event     string          `json:"event"`
	CreatedAt time.Time       `json:"created_at"`
	StoreID   uint            `json:"store_id"`
	Data      json.RawMessage `json:"data"`
}

// Dispatcher mengirim event ke webhook pelanggan secara ASINKRON:
// metode Emit* tidak pernah memblokir atau menggagalkan operasi bisnis
// pemanggilnya — pengiriman berjalan di goroutine dengan konteksnya
// sendiri dan setiap percobaan dicatat ke webhook_deliveries.
type Dispatcher struct {
	repo   Repository
	client *http.Client
}

func NewDispatcher(repo Repository) *Dispatcher {
	return &Dispatcher{repo: repo, client: &http.Client{Timeout: 5 * time.Second}}
}

func (d *Dispatcher) EmitTransactionCreated(ctx context.Context, storeID uint, data TransactionData) {
	d.emit(storeID, EventTransactionCreated, data)
}

func (d *Dispatcher) EmitTransactionRefunded(ctx context.Context, storeID uint, data TransactionData) {
	d.emit(storeID, EventTransactionRefunded, data)
}

func (d *Dispatcher) EmitStockUpdated(ctx context.Context, storeID uint, data StockData) {
	d.emit(storeID, EventStockUpdated, data)
}

func (d *Dispatcher) emit(storeID uint, event string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	// Konteks permintaan bisnis sengaja TIDAK diteruskan: ia dibatalkan
	// begitu respons bisnis selesai, sedangkan pengiriman harus tetap
	// berjalan. Dispatcher memakai konteks latarnya sendiri.
	go d.dispatch(storeID, event, raw)
}

func (d *Dispatcher) dispatch(storeID uint, event string, dataRaw json.RawMessage) {
	ctx := context.Background()
	hooks, err := d.repo.ListSubscribed(ctx, storeID, event)
	if err != nil {
		return
	}
	for _, h := range hooks {
		h := h
		go d.deliver(ctx, h, event, dataRaw)
	}
}

func (d *Dispatcher) deliver(ctx context.Context, h Webhook, event string, dataRaw json.RawMessage) {
	body, err := json.Marshal(envelope{
		ID:        newUUID(),
		Event:     event,
		CreatedAt: time.Now().UTC(),
		StoreID:   h.StoreID,
		Data:      dataRaw,
	})
	if err != nil {
		return
	}
	delivery := &Delivery{WebhookID: h.ID, StoreID: h.StoreID, Event: event, Payload: string(body)}
	if err := d.repo.CreateDelivery(ctx, delivery); err != nil {
		return
	}
	status, success, errMsg := d.send(ctx, h, event, body)
	_ = d.repo.UpdateDeliveryResult(ctx, delivery.ID, status, success, errMsg)
}

// send melakukan SATU percobaan POST (tanpa antrean ulang di v1) dan
// melaporkan hasilnya untuk dicatat. Tanda tangan dihitung dari body
// mentah yang sama persis dengan yang dikirim.
func (d *Dispatcher) send(ctx context.Context, h Webhook, event string, body []byte) (status *int, success bool, errMsg string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewReader(body))
	if err != nil {
		return nil, false, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-NOJPOS-Event", event)
	mac := hmac.New(sha256.New, []byte(h.Secret))
	mac.Write(body)
	req.Header.Set("X-NOJPOS-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, false, err.Error()
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	code := resp.StatusCode
	if code >= 200 && code < 300 {
		return &code, true, ""
	}
	return &code, false, fmt.Sprintf("status penerima %d", code)
}

// newUUID membuat UUID v4 dari crypto/rand (tanpa dependensi baru).
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
