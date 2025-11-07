package repository

import (
	"time"

	"gorm.io/gorm"
)

// FlashSaleEvent is the database model for a flash sale event.
type FlashSaleEvent struct {
	gorm.Model
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	StartTime   time.Time `gorm:"not null"`
	EndTime     time.Time `gorm:"not null"`
	Status      string    `gorm:"type:varchar(50);not null"` // e.g., UPCOMING, ACTIVE, ENDED
}

// TableName specifies the table name for the FlashSaleEvent model.
func (FlashSaleEvent) TableName() string {
	return "flash_sale_events"
}

// FlashSaleProduct is the database model for a product within a flash sale event.
type FlashSaleProduct struct {
	gorm.Model
	EventID        uint    `gorm:"not null;index"` // Foreign key to FlashSaleEvent
	ProductID      string  `gorm:"type:varchar(100);not null;index"`
	FlashPrice     float64 `gorm:"not null"`
	TotalStock     int32   `gorm:"not null"`
	RemainingStock int32   `gorm:"not null"`
	MaxPerUser     int32   `gorm:"default:1"`
}

// TableName specifies the table name for the FlashSaleProduct model.
func (FlashSaleProduct) TableName() string {
	return "flash_sale_products"
}
