package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

// Repository adalah kontrak penyimpanan yang dipakai Service dan
// Dispatcher (diimplementasikan Repo).
type Repository interface {
	Create(ctx context.Context, w *Webhook) error
	List(ctx context.Context, storeID uint) ([]Webhook, error)
	Delete(ctx context.Context, storeID, id uint) error
	ListSubscribed(ctx context.Context, storeID uint, event string) ([]Webhook, error)
	CreateDelivery(ctx context.Context, d *Delivery) error
	UpdateDeliveryResult(ctx context.Context, id uint, status *int, success bool, errMsg string) error
}

// Service melayani manajemen langganan webhook (CRUD admin).
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, storeID uint) ([]Metadata, error) {
	rows, err := s.repo.List(ctx, storeID)
	if err != nil {
		return nil, err
	}
	out := make([]Metadata, 0, len(rows))
	for _, w := range rows {
		out = append(out, Metadata{
			ID:        w.ID,
			URL:       w.URL,
			Events:    parseEvents(w.Events),
			CreatedAt: w.CreatedAt,
		})
	}
	return out, nil
}

// Create memvalidasi URL + events, membuat secret baru, dan menyimpan
// langganan. Secret dikembalikan SEKALI di sini (untuk respons 201);
// tidak ada jalur baca lain yang pernah memuatnya.
func (s *Service) Create(ctx context.Context, storeID uint, rawURL string, events []string) (Metadata, string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if err := validateEvents(events); err != nil {
		return Metadata{}, "", err
	}
	if err := validateURL(ctx, rawURL); err != nil {
		return Metadata{}, "", err
	}
	var random [20]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Metadata{}, "", err
	}
	secret := "whsec_" + hex.EncodeToString(random[:])
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return Metadata{}, "", err
	}
	w := &Webhook{StoreID: storeID, URL: rawURL, Events: string(eventsJSON), Secret: secret}
	if err := s.repo.Create(ctx, w); err != nil {
		return Metadata{}, "", err
	}
	return Metadata{ID: w.ID, URL: w.URL, Events: events, CreatedAt: w.CreatedAt}, secret, nil
}

func (s *Service) Delete(ctx context.Context, storeID, id uint) error {
	return s.repo.Delete(ctx, storeID, id)
}

func validateEvents(events []string) error {
	if len(events) == 0 {
		return ErrEventsEmpty
	}
	for _, e := range events {
		if !ValidEvent(e) {
			return ErrEventUnknown
		}
	}
	return nil
}

// lookupIPAddr adalah jahitan resolusi DNS agar validasi URL dapat
// diuji tanpa jaringan.
var lookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

// validateURL memastikan URL http/https yang menunjuk alamat publik:
// host literal IP diperiksa langsung; nama host di-resolve dan SEMUA
// hasil resolusinya harus publik. Ini pertahanan SSRF — tanpa ini admin
// (atau akun admin yang dibajak) dapat membuat server menembak layanan
// internal (metadata cloud, admin panel LAN, dsb.).
func validateURL(ctx context.Context, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ErrInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidURL
	}
	host := u.Hostname()
	if host == "" {
		return ErrInvalidURL
	}
	if strings.EqualFold(host, "localhost") {
		return ErrPrivateURL
	}
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return ErrPrivateURL
		}
		return nil
	}
	addrs, err := lookupIPAddr(ctx, host)
	if err != nil || len(addrs) == 0 {
		return ErrInvalidURL
	}
	for _, a := range addrs {
		if isBlockedIP(a.IP) {
			return ErrPrivateURL
		}
	}
	return nil
}

// isBlockedIP menolak loopback, privat (RFC 1918 + alamat lokal unik +
// ruang bersama RFC 6598 via net.IP.IsPrivate), link-local, multicast
// link-local, dan jaringan 0.0.0.0/8.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil && v4[0] == 0 {
		return true
	}
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified()
}
