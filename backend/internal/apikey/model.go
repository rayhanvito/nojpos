package apikey

import (
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

type APIKey struct {
	database.BaseModel
	StoreID    uint       `gorm:"not null;index" json:"-"`
	Name       string     `gorm:"not null" json:"-"`
	KeyPrefix  string     `gorm:"not null" json:"-"`
	KeyHash    string     `gorm:"not null;uniqueIndex" json:"-"`
	CreatedBy  uint       `gorm:"not null" json:"-"`
	LastUsedAt *time.Time `json:"-"`
}

func (APIKey) TableName() string { return "api_keys" }

type Metadata struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `gorm:"column:key_prefix" json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}
