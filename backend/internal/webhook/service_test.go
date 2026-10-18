package webhook

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"testing"
)

// fakeRepo adalah Repository ingatan untuk uji Service: Delete
// meniru semantik repo asli (store-scoped, ErrNotFound untuk toko lain)
// dalam skala kecil; jalur lain dicatat seadanya.
type fakeRepo struct {
	rows   []Webhook
	nextID uint
}

func (f *fakeRepo) Create(ctx context.Context, w *Webhook) error {
	f.nextID++
	w.ID = f.nextID
	f.rows = append(f.rows, *w)
	return nil
}

func (f *fakeRepo) List(ctx context.Context, storeID uint) ([]Webhook, error) {
	out := make([]Webhook, 0)
	for _, w := range f.rows {
		if w.StoreID == storeID {
			out = append(out, w)
		}
	}
	return out, nil
}

func (f *fakeRepo) Delete(ctx context.Context, storeID, id uint) error {
	for i, w := range f.rows {
		if w.ID == id {
			if w.StoreID != storeID {
				return ErrNotFound
			}
			f.rows = append(f.rows[:i], f.rows[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (f *fakeRepo) ListSubscribed(ctx context.Context, storeID uint, event string) ([]Webhook, error) {
	return nil, nil
}

func (f *fakeRepo) CreateDelivery(ctx context.Context, d *Delivery) error { return nil }

func (f *fakeRepo) UpdateDeliveryResult(ctx context.Context, id uint, status *int, success bool, errMsg string) error {
	return nil
}

const publicURL = "https://93.184.216.34/hook" // IP literal publik: tanpa DNS

func TestCreateRejectsInvalidEvents(t *testing.T) {
	svc := NewService(&fakeRepo{})
	ctx := context.Background()

	if _, _, err := svc.Create(ctx, 1, publicURL, nil); !errors.Is(err, ErrEventsEmpty) {
		t.Fatalf("events nil: mau ErrEventsEmpty, dapat %v", err)
	}
	if _, _, err := svc.Create(ctx, 1, publicURL, []string{}); !errors.Is(err, ErrEventsEmpty) {
		t.Fatalf("events kosong: mau ErrEventsEmpty, dapat %v", err)
	}
	if _, _, err := svc.Create(ctx, 1, publicURL, []string{"stok.berubah"}); !errors.Is(err, ErrEventUnknown) {
		t.Fatalf("event asing: mau ErrEventUnknown, dapat %v", err)
	}
	if _, _, err := svc.Create(ctx, 1, publicURL, []string{EventTransactionCreated, "bogus"}); !errors.Is(err, ErrEventUnknown) {
		t.Fatalf("satu event asing: mau ErrEventUnknown, dapat %v", err)
	}
}

func TestCreateRejectsPrivateAndBadURLs(t *testing.T) {
	svc := NewService(&fakeRepo{})
	ctx := context.Background()
	events := []string{EventTransactionCreated}

	for _, raw := range []string{
		"http://127.0.0.1:8080/hook",
		"http://127.5.6.7/hook",
		"http://10.0.0.8/hook",
		"http://172.16.0.9/hook",
		"http://172.31.255.1/hook",
		"http://192.168.1.10/hook",
		"http://169.254.169.254/latest/meta-data",
		"http://[::1]/hook",
		"http://localhost/hook",
		"http://LOCALHOST:3000/hook",
		"http://0.0.0.0/hook",
	} {
		if _, _, err := svc.Create(ctx, 1, raw, events); !errors.Is(err, ErrPrivateURL) {
			t.Errorf("%s: mau ErrPrivateURL, dapat %v", raw, err)
		}
	}
	for _, raw := range []string{"", "bukan-url", "ftp://example.com/hook", "https://", "//example.com/hook"} {
		if _, _, err := svc.Create(ctx, 1, raw, events); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("%q: mau ErrInvalidURL, dapat %v", raw, err)
		}
	}
}

func TestCreateRejectsHostnameResolvingToPrivateIP(t *testing.T) {
	original := lookupIPAddr
	t.Cleanup(func() { lookupIPAddr = original })
	ctx := context.Background()
	events := []string{EventStockUpdated}
	svc := NewService(&fakeRepo{})

	lookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}, {IP: net.ParseIP("192.168.7.7")}}, nil
	}
	if _, _, err := svc.Create(ctx, 1, "https://hooks.contoh.id/in", events); !errors.Is(err, ErrPrivateURL) {
		t.Fatalf("satu hasil resolusi privat harus menolak: dapat %v", err)
	}

	lookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}}, nil
	}
	if _, _, err := svc.Create(ctx, 1, "https://hooks.contoh.id/in", events); err != nil {
		t.Fatalf("resolusi publik harus lolos: dapat %v", err)
	}
}

func TestCreateSuccessSecretFormatAndListHidesSecret(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	ctx := context.Background()

	meta, secret, err := svc.Create(ctx, 42, publicURL, []string{EventTransactionCreated, EventStockUpdated})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(secret, "whsec_") || len(secret) != len("whsec_")+40 {
		t.Fatalf("format secret salah: %q", secret)
	}
	if _, err := hex.DecodeString(secret[len("whsec_"):]); err != nil {
		t.Fatalf("secret bukan hex: %v", err)
	}
	if meta.ID == 0 || meta.URL != publicURL || len(meta.Events) != 2 {
		t.Fatalf("metadata tidak sesuai: %+v", meta)
	}

	items, err := svc.List(ctx, 42)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("list: mau 1 item, dapat %d", len(items))
	}
	raw, _ := json.Marshal(items)
	if strings.Contains(string(raw), secret) || strings.Contains(string(raw), "secret") {
		t.Fatalf("daftar webhook membocorkan secret: %s", raw)
	}

	other, err := svc.List(ctx, 43)
	if err != nil || len(other) != 0 {
		t.Fatalf("list toko lain harus kosong: %v %v", other, err)
	}
}

func TestDeleteScopesToStore(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)
	ctx := context.Background()

	meta, _, err := svc.Create(ctx, 7, publicURL, []string{EventTransactionRefunded})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.Delete(ctx, 8, meta.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("hapus milik toko lain: mau ErrNotFound, dapat %v", err)
	}
	if err := svc.Delete(ctx, 7, meta.ID); err != nil {
		t.Fatalf("hapus milik sendiri: %v", err)
	}
	if err := svc.Delete(ctx, 7, meta.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("hapus dua kali: mau ErrNotFound, dapat %v", err)
	}
}
