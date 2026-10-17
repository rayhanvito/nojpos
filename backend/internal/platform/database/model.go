package database

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel is gorm.Model with the API's JSON names. gorm.Model itself
// carries no JSON tags and would serialize as "ID"/"CreatedAt", breaking
// every client expecting "id". Behavior is identical: auto-increment uint
// PK, auto timestamps, soft delete via DeletedAt (hidden from JSON).
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
