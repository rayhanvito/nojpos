package auth

import (
	"time"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// Tipe-tipe di file ini dipindahkan verbatim dari model/models.go dan
// model/otp.go (BE-F1c): User, PublicUser, TokenPair, RefreshToken,
// EmailOtp.

type User struct {
	database.BaseModel
	StoreID         uint            `gorm:"not null;index" json:"store_id"`
	OutletID        *uint           `gorm:"index" json:"outlet_id,omitempty"`
	Email           string          `gorm:"not null;unique" json:"email"`
	Name            string          `gorm:"not null" json:"name"`
	PasswordHash    string          `gorm:"not null" json:"-"`
	PasscodeHash    *string         `json:"-"`
	Role            middleware.Role `gorm:"not null;default:'cashier'" json:"role"`
	Active          bool            `gorm:"not null;default:true" json:"active"`
	LastSeenAt      *time.Time      `gorm:"index" json:"last_seen_at,omitempty"`
	EmailVerifiedAt *time.Time      `json:"email_verified_at,omitempty"`
	StoreName       string          `gorm:"-" json:"store_name,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type PublicUser struct {
	ID          uint            `json:"id"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	Role        middleware.Role `json:"role"`
	Active      bool            `json:"active"`
	Online      bool            `json:"online"`
	LastSeenAt  *time.Time      `json:"last_seen_at,omitempty"`
	StoreID     uint            `json:"store_id"`
	StoreName   string          `json:"store_name,omitempty"`
	OutletID    *uint           `json:"outlet_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at,omitempty"`
	HasPasscode bool            `json:"has_passcode"`
}

func (u *User) Public() PublicUser {
	online := false
	if u.LastSeenAt != nil {
		online = time.Since(*u.LastSeenAt) < 90*time.Second
	}
	return PublicUser{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		Role:        u.Role,
		Active:      u.Active,
		Online:      online,
		LastSeenAt:  u.LastSeenAt,
		StoreID:     u.StoreID,
		StoreName:   u.StoreName,
		OutletID:    u.OutletID,
		CreatedAt:   u.CreatedAt,
		HasPasscode: u.PasscodeHash != nil && *u.PasscodeHash != "",
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshToken struct {
	database.BaseModel
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	OutletID  *uint     `gorm:"index" json:"outlet_id,omitempty"`
	TokenHash string    `gorm:"not null;unique" json:"token_hash"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Revoked   bool      `gorm:"not null;default:false" json:"revoked"`
	// Sesi kasir disimpan eksplisit: refresh wajib mempertahankan identitas
	// acting-as, kalau tidak sesi kasir berubah jadi admin tiap refresh.
	ActingAsCashierID *uint `gorm:"index" json:"-"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

type EmailOtp struct {
	Email      string     `gorm:"primaryKey" json:"email"`
	CodeHash   string     `gorm:"not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	Attempts   int        `gorm:"not null;default:0" json:"attempts"`
	LastSentAt time.Time  `gorm:"not null" json:"last_sent_at"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `gorm:"not null" json:"created_at"`
}

func (EmailOtp) TableName() string {
	return "email_otps"
}

func (o *EmailOtp) BeforeCreate(tx *gorm.DB) error {
	if o.LastSentAt.IsZero() {
		o.LastSentAt = time.Now()
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}
	return nil
}

// PublicUserFromCashier adalah converter PublicUser dari kasir
// (badan verbatim dari method Cashier.Public lama di internal/staff;
// dipindah ke sini karena auth yang memiliki tipe PublicUser dan
// staff tidak boleh mengimpor auth).
func PublicUserFromCashier(c *staff.Cashier, storeName string) PublicUser {
	online := false
	if c.LastSeenAt != nil {
		online = time.Since(*c.LastSeenAt) < 90*time.Second
	}
	return PublicUser{
		ID:          c.ID,
		Email:       "",
		Name:        c.Name,
		Role:        middleware.RoleCashier,
		Active:      c.Active,
		Online:      online,
		LastSeenAt:  c.LastSeenAt,
		StoreID:     c.StoreID,
		StoreName:   storeName,
		OutletID:    c.OutletID,
		CreatedAt:   c.CreatedAt,
		HasPasscode: c.PasscodeHash != nil && *c.PasscodeHash != "",
	}
}
