package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Repo-repo di file ini dipindahkan dari repo/{user,refresh,otp}.go
// (BE-F1c): embed database.BaseRepo, akses DB lewat DB(ctx), badan
// metode verbatim.

// mapDBErr memetakan error database ke sentinel slice, setara mapDBErr
// lama di repo/base.go: record kosong -> ErrNotFound; pelanggaran unik
// (Postgres 23505 / SQLite) -> ErrDuplicate.
func mapDBErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "23505") ||
			strings.Contains(msg, "duplicate key") ||
			strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "unique constraint") {
			return ErrDuplicate
		}
	}
	return err
}

// UserRepo adalah repositori user (UserRepo lama di repo/user.go).
type UserRepo struct {
	database.BaseRepo
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	var store outlet.Store

	err := r.DB(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	if err := r.DB(ctx).First(&store, "id = ?", u.StoreID).Error; err == nil {
		u.StoreName = store.Name
	}
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uint) (*User, error) {
	var u User
	var store outlet.Store

	err := r.DB(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	if err := r.DB(ctx).First(&store, "id = ?", u.StoreID).Error; err == nil {
		u.StoreName = store.Name
	}
	return &u, nil
}

func (r *UserRepo) ListByStore(ctx context.Context, storeID uint) ([]*User, error) {
	users := make([]*User, 0)
	var store outlet.Store
	if err := r.DB(ctx).First(&store, "id = ?", storeID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	storeName := store.Name

	err := r.DB(ctx).Where("store_id = ?", storeID).Order("created_at ASC").Find(&users).Error
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		u.StoreName = storeName
	}
	return users, nil
}

func (r *UserRepo) SetActive(ctx context.Context, id uint, active bool) error {
	res := r.DB(ctx).Model(&User{}).Where("id = ?", id).Update("active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) SetPasscode(ctx context.Context, id uint, hash *string) error {
	res := r.DB(ctx).Model(&User{}).Where("id = ?", id).Update("passcode_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) RegisterTx(ctx context.Context, storeName, email, name, passwordHash string) (*User, error) {
	var user *User
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		store := outlet.Store{Name: storeName}
		if err := tx.Create(&store).Error; err != nil {
			return mapDBErr(err)
		}
		outlet := outlet.Outlet{StoreID: store.ID, Name: storeName, Address: store.Address, IsDefault: true}
		if err := tx.Create(&outlet).Error; err != nil {
			return err
		}

		now := time.Now()
		u := User{
			StoreID:         store.ID,
			OutletID:        &outlet.ID,
			Email:           email,
			Name:            name,
			PasswordHash:    passwordHash,
			Role:            middleware.RoleAdmin,
			Active:          true,
			EmailVerifiedAt: &now,
		}
		if err := tx.Create(&u).Error; err != nil {
			return mapDBErr(err)
		}
		u.StoreName = store.Name
		user = &u
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, email string, passwordHash string) error {
	res := r.DB(ctx).Model(&User{}).Where("email = ?", email).Update("password_hash", passwordHash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdatePasswordByID(ctx context.Context, id uint, passwordHash string) error {
	res := r.DB(ctx).Model(&User{}).Where("id = ?", id).Update("password_hash", passwordHash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdateLastSeenAt(ctx context.Context, userID uint, lastSeenAt *time.Time) error {
	res := r.DB(ctx).Model(&User{}).Where("id = ?", userID).Update("last_seen_at", lastSeenAt)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RefreshRepo adalah repositori refresh token (repo/refresh.go lama).
type RefreshRepo struct {
	database.BaseRepo
}

func NewRefreshRepo(db *gorm.DB) *RefreshRepo {
	return &RefreshRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *RefreshRepo) Create(ctx context.Context, userID uint, actingAsCashierID, outletID *uint, tokenHash string, expiresAt time.Time) error {
	rt := RefreshToken{
		UserID:            userID,
		ActingAsCashierID: actingAsCashierID,
		OutletID:          outletID,
		TokenHash:         tokenHash,
		ExpiresAt:         expiresAt,
		Revoked:           false,
	}
	return r.DB(ctx).Create(&rt).Error
}

func (r *RefreshRepo) GetActiveByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.DB(ctx).Where("token_hash = ?", tokenHash).First(&rt).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &rt, nil
}

func (r *RefreshRepo) Revoke(ctx context.Context, tokenHash string) error {
	return r.DB(ctx).Model(&RefreshToken{}).Where("token_hash = ?", tokenHash).Update("revoked", true).Error
}

func (r *RefreshRepo) RevokeAllForUser(ctx context.Context, userID uint) error {
	return r.DB(ctx).Model(&RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}

// OtpRepo adalah repositori OTP email (repo/otp.go lama).
type OtpRepo struct {
	database.BaseRepo
}

func NewOtpRepo(db *gorm.DB) *OtpRepo {
	return &OtpRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *OtpRepo) UpsertOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) error {
	otp := EmailOtp{
		Email:      email,
		CodeHash:   codeHash,
		ExpiresAt:  expiresAt,
		Attempts:   0,
		LastSentAt: time.Now(),
		VerifiedAt: nil,
		CreatedAt:  time.Now(),
	}
	return r.DB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "email"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"code_hash":    codeHash,
			"expires_at":   expiresAt,
			"attempts":     0,
			"last_sent_at": time.Now(),
			"verified_at":  nil,
			"created_at":   time.Now(),
		}),
	}).Create(&otp).Error
}

func (r *OtpRepo) GetOTP(ctx context.Context, email string) (*EmailOtp, error) {
	var o EmailOtp
	err := r.DB(ctx).Where("email = ?", email).First(&o).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &o, nil
}

func (r *OtpRepo) IncrementAttempts(ctx context.Context, email string) (int, error) {
	var o EmailOtp
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("email = ?", email).First(&o).Error; err != nil {
			return err
		}
		o.Attempts++
		return tx.Save(&o).Error
	})
	if err != nil {
		return 0, mapDBErr(err)
	}
	return o.Attempts, nil
}

func (r *OtpRepo) MarkVerified(ctx context.Context, email string) error {
	now := time.Now()
	res := r.DB(ctx).Model(&EmailOtp{}).Where("email = ?", email).Update("verified_at", &now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *OtpRepo) IsEmailVerified(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&EmailOtp{}).Where("email = ? AND verified_at IS NOT NULL", email).Count(&count).Error
	return count > 0, err
}

func (r *OtpRepo) RevokeOTP(ctx context.Context, email string) error {
	return r.DB(ctx).Model(&EmailOtp{}).Where("email = ?", email).Updates(map[string]interface{}{
		"attempts":    3,
		"verified_at": nil,
	}).Error
}
