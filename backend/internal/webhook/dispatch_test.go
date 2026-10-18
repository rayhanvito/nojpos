package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

func newWebhookTestRepo(t *testing.T) (*Repo, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "webhook.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &Webhook{}, &Delivery{}); err != nil {
		t.Fatal(err)
	}
	return NewRepo(db), db
}

type captured struct {
	body      []byte
	event     string
	signature string
}

type receiver struct {
	ch     chan captured
	server *httptest.Server
}

func newReceiver(t *testing.T) *receiver {
	t.Helper()
	r := &receiver{ch: make(chan captured, 16)}
	r.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		r.ch <- captured{body: body, event: req.Header.Get("X-NOJPOS-Event"), signature: req.Header.Get("X-NOJPOS-Signature")}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(r.server.Close)
	return r
}

func (r *receiver) wait(t *testing.T) captured {
	t.Helper()
	select {
	case c := <-r.ch:
		return c
	case <-time.After(5 * time.Second):
		t.Fatal("penerima tidak menerima kiriman webhook dalam 5 detik")
		return captured{}
	}
}

func (r *receiver) expectNone(t *testing.T, within time.Duration) {
	t.Helper()
	select {
	case c := <-r.ch:
		t.Fatalf("penerima tidak boleh menerima kiriman, dapat event %q", c.event)
	case <-time.After(within):
	}
}

func seedWebhook(t *testing.T, repo *Repo, storeID uint, url string, events []string, secret string) Webhook {
	t.Helper()
	raw, _ := json.Marshal(events)
	w := Webhook{StoreID: storeID, URL: url, Events: string(raw), Secret: secret}
	if err := repo.Create(context.Background(), &w); err != nil {
		t.Fatal(err)
	}
	return w
}

func waitDelivery(t *testing.T, db *gorm.DB, webhookID uint, pred func(Delivery) bool) Delivery {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var rows []Delivery
		if err := db.Where("webhook_id = ?", webhookID).Order("id").Find(&rows).Error; err == nil {
			for _, row := range rows {
				if pred(row) {
					return row
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("baris delivery untuk webhook %d tidak mencapai keadaan yang ditunggu", webhookID)
	return Delivery{}
}

func TestDispatchSignaturePayloadFilterAndRecord(t *testing.T) {
	repo, db := newWebhookTestRepo(t)
	ctx := context.Background()
	const storeID = 7
	const secret = "whsec_0123456789abcdef0123456789abcdef01234567"

	trxRecv := newReceiver(t)
	stockRecv := newReceiver(t)
	trxHook := seedWebhook(t, repo, storeID, trxRecv.server.URL, []string{EventTransactionCreated}, secret)
	stockHook := seedWebhook(t, repo, storeID, stockRecv.server.URL, []string{EventStockUpdated}, secret)
	otherStoreHook := seedWebhook(t, repo, 99, trxRecv.server.URL, []string{EventTransactionCreated}, secret)

	d := NewDispatcher(repo)
	d.EmitTransactionCreated(ctx, storeID, TransactionData{
		TransactionID: 123, Number: "TRX-00123", Status: "completed",
		Total: 45000, Method: "QRIS", CashierName: "Kasir Uji",
	})

	got := trxRecv.wait(t)
	if got.event != EventTransactionCreated {
		t.Fatalf("header X-NOJPOS-Event = %q", got.event)
	}
	// Tanda tangan harus cocok dengan HMAC-SHA256(secret, body mentah
	// persis seperti diterima).
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(got.body)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if got.signature != want {
		t.Fatalf("tanda tangan = %q, mau %q", got.signature, want)
	}
	var env struct {
		ID        string `json:"id"`
		Event     string `json:"event"`
		CreatedAt string `json:"created_at"`
		StoreID   uint   `json:"store_id"`
		Data      struct {
			TransactionID uint   `json:"transaction_id"`
			Number        string `json:"number"`
			Total         int64  `json:"total"`
			Method        string `json:"method"`
		} `json:"data"`
	}
	if err := json.Unmarshal(got.body, &env); err != nil {
		t.Fatalf("payload bukan JSON sah: %v", err)
	}
	if env.ID == "" || env.Event != EventTransactionCreated || env.StoreID != storeID ||
		env.Data.TransactionID != 123 || env.Data.Number != "TRX-00123" || env.Data.Total != 45000 || env.Data.Method != "QRIS" {
		t.Fatalf("isi payload tidak sesuai: %s", got.body)
	}

	// Filter langganan: webhook stock.updated tidak menerima event ini,
	// dan webhook toko lain tidak menerima apa pun dari toko ini.
	stockRecv.expectNone(t, 400*time.Millisecond)

	row := waitDelivery(t, db, trxHook.ID, func(r Delivery) bool { return r.Success })
	if row.ResponseStatus == nil || *row.ResponseStatus != 200 {
		t.Fatalf("delivery tercatat status %v, mau 200", row.ResponseStatus)
	}
	if row.Payload != string(got.body) {
		t.Fatalf("payload tersimpan != body terkirim:\n%s\nvs\n%s", row.Payload, got.body)
	}
	if row.Event != EventTransactionCreated || row.StoreID != storeID {
		t.Fatalf("baris delivery salah: %+v", row)
	}
	var stray int64
	db.Model(&Delivery{}).Where("webhook_id IN ?", []uint{stockHook.ID, otherStoreHook.ID}).Count(&stray)
	if stray != 0 {
		t.Fatalf("ada %d delivery untuk webhook yang tidak berlangganan/toko lain", stray)
	}

	// Event stok hanya sampai ke pelanggan stock.updated.
	d.EmitStockUpdated(ctx, storeID, StockData{Source: "adjustment", ProductID: 5, QtyDelta: -2, OutletID: 1, Reason: "rusak"})
	gotStock := stockRecv.wait(t)
	if gotStock.event != EventStockUpdated {
		t.Fatalf("header event stok = %q", gotStock.event)
	}
	if !strings.Contains(string(gotStock.body), `"qty_delta":-2`) || !strings.Contains(string(gotStock.body), `"source":"adjustment"`) {
		t.Fatalf("payload stok tidak memuat data: %s", gotStock.body)
	}
	trxRecv.expectNone(t, 400*time.Millisecond)
	waitDelivery(t, db, stockHook.ID, func(r Delivery) bool { return r.Success })
}

func TestDispatchDeadReceiverRecordedFailed(t *testing.T) {
	repo, db := newWebhookTestRepo(t)
	ctx := context.Background()
	hook := seedWebhook(t, repo, 3, "http://127.0.0.1:1/dead", []string{EventTransactionCreated}, "whsec_dead")

	d := NewDispatcher(repo)
	start := time.Now()
	d.EmitTransactionCreated(ctx, 3, TransactionData{TransactionID: 1, Number: "TRX-00001", Status: "completed", Total: 1000, Method: "Cash"})
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Emit memblokir jalur bisnis selama %v", elapsed)
	}

	row := waitDelivery(t, db, hook.ID, func(r Delivery) bool { return r.Error != "" })
	if row.Success {
		t.Fatal("delivery ke penerima mati tercatat sukses")
	}
	if row.ResponseStatus != nil {
		t.Fatalf("response_status harus NULL untuk kegagalan jaringan, dapat %v", *row.ResponseStatus)
	}
}
