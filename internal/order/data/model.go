package data

import (
	"time"

	"gorm.io/gorm"
)

// Order 对应数据库中的 `orders` 表。
type Order struct {
	ID              uint64 `gorm:"primarykey"`
	UserID          uint64 `gorm:"index;not null"`
	TotalAmount     uint64 `gorm:"not null"`
	PaymentAmount   uint64 `gorm:"not null"`
	ShippingFee     uint64 `gorm:"not null"`
	Status          int8   `gorm:"not null;index"`
	ShippingAddress []byte `gorm:"type:json"` // 直接用 []byte 存储 JSON
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

// OrderItem 对应数据库中的 `order_items` 表。
type OrderItem struct {
	ID           uint64 `gorm:"primarykey"`
	OrderID      uint64 `gorm:"index;not null"`
	SkuID        uint64 `gorm:"index;not null"`
	SpuID        uint64 `gorm:"index"`
	ProductTitle string `gorm:"type:varchar(255)"`
	ProductImage string `gorm:"type:varchar(255)"`
	Price        uint64 `gorm:"not null"`
	Quantity     uint32 `gorm:"not null"`
	SubTotal     uint64 `gorm:"not null"`
}

func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}
