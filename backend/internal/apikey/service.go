package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

type Repository interface {
	Create(ctx context.Context, key *APIKey) error
	List(ctx context.Context, storeID uint) ([]Metadata, error)
	GetByHash(ctx context.Context, hash string) (*APIKey, error)
	MarkUsed(ctx context.Context, keyID, storeID uint, at time.Time) error
	Delete(ctx context.Context, storeID, id uint) error
}

type DefaultOutletReader interface {
	GetDefault(ctx context.Context, storeID uint) (*outlet.Outlet, error)
}

type Service struct {
	keys    Repository
	outlets DefaultOutletReader
}

func NewService(keys Repository, outlets DefaultOutletReader) *Service {
	return &Service{keys: keys, outlets: outlets}
}

func (s *Service) Create(ctx context.Context, storeID, createdBy uint, name string) (Metadata, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Metadata{}, "", ErrInvalidName
	}

	var random [20]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Metadata{}, "", err
	}
	fullKey := "njp_" + hex.EncodeToString(random[:])
	sum := sha256.Sum256([]byte(fullKey))
	key := &APIKey{
		StoreID: storeID, Name: name, KeyPrefix: fullKey[:11],
		KeyHash: hex.EncodeToString(sum[:]), CreatedBy: createdBy,
	}
	if err := s.keys.Create(ctx, key); err != nil {
		return Metadata{}, "", err
	}
	return Metadata{ID: key.ID, Name: key.Name, Prefix: key.KeyPrefix, CreatedAt: key.CreatedAt}, fullKey, nil
}

func (s *Service) List(ctx context.Context, storeID uint) ([]Metadata, error) {
	return s.keys.List(ctx, storeID)
}

func (s *Service) Delete(ctx context.Context, storeID, id uint) error {
	return s.keys.Delete(ctx, storeID, id)
}

func (s *Service) Authenticate(ctx context.Context, rawKey string) (*middleware.Claims, error) {
	if len(rawKey) != 44 || !strings.HasPrefix(rawKey, "njp_") {
		return nil, middleware.ErrInvalidAPIKey
	}
	decoded, err := hex.DecodeString(rawKey[4:])
	if err != nil || len(decoded) != 20 {
		return nil, middleware.ErrInvalidAPIKey
	}
	sum := sha256.Sum256([]byte(rawKey))
	key, err := s.keys.GetByHash(ctx, hex.EncodeToString(sum[:]))
	if errors.Is(err, ErrNotFound) {
		return nil, middleware.ErrInvalidAPIKey
	}
	if err != nil {
		return nil, err
	}
	defaultOutlet, err := s.outlets.GetDefault(ctx, key.StoreID)
	if err != nil {
		return nil, err
	}
	// last_used_at ditulis paling sering sekali per menit per kunci:
	// cukup untuk visibilitas pemakaian, tanpa menulis DB di setiap
	// permintaan terautentikasi.
	now := time.Now().UTC()
	if key.LastUsedAt == nil || !key.LastUsedAt.After(now.Add(-time.Minute)) {
		if err := s.keys.MarkUsed(ctx, key.ID, key.StoreID, now); err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, middleware.ErrInvalidAPIKey
			}
			return nil, err
		}
	}
	return &middleware.Claims{
		UserID: key.CreatedBy, StoreID: key.StoreID, OutletID: defaultOutlet.ID,
		Role: middleware.RoleAdmin, APIKey: true,
	}, nil
}
