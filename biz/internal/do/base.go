package do

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"update_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
