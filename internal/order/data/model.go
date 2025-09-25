package data

import (
	"time"

	"gorm.io/gorm"
)

// Order 对应数据库中的 `orders` 表。
type Order struct {
	gorm.Model
	UserID          uint64 `gorm:"index;not null" json:"userId"`
	TotalAmount     uint64 `gorm:"not null" json:"totalAmount"`
	PaymentAmount   uint64 `gorm:"not null" json:"paymentAmount"`
	ShippingFee     uint64 `gorm:"not null" json:"shippingFee"`
	Status          int8   `gorm:"index;not null" json:"status"`
	ShippingAddress []byte `gorm:"type:json" json:"shippingAddress"` // 直接用 []byte 存储 JSON
}

// OrderItem 对应数据库中的 `order_items` 表。
type OrderItem struct {
	gorm.Model
	OrderID      uint64 `gorm:"index;not null" json:"orderId"`
	SkuID        uint64 `gorm:"index;not null" json:"skuId"`
	SpuID        uint64 `gorm:"index" json:"spuId"`
	ProductTitle string `gorm:"type:varchar(255)" json:"productTitle"`
	ProductImage string `gorm:"type:varchar(255)" json:"productImage"`
	Price        uint64 `gorm:"not null" json:"price"`
	Quantity     uint32 `gorm:"not null" json:"quantity"`
	SubTotal     uint64 `gorm:"not null" json:"subTotal"`
}

func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}
