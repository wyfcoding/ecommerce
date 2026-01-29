package domain

import (
	"time"
)

// ProductCreatedEvent 商品创建事件
type ProductCreatedEvent struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	Stock     int32     `json:"stock"`
	Timestamp time.Time `json:"timestamp"`
}

// ProductUpdatedEvent 商品更新事件
type ProductUpdatedEvent struct {
	ID        uint      `json:"id"`
	Status    int       `json:"status"` // ProductStatus
	Timestamp time.Time `json:"timestamp"`
}

// ProductDeletedEvent 商品删除事件
type ProductDeletedEvent struct {
	ID        uint      `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// SKUAddedEvent SKU 新增事件
type SKUAddedEvent struct {
	ProductID uint      `json:"product_id"`
	SKUID     uint      `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SKUUpdatedEvent SKU 更新事件
type SKUUpdatedEvent struct {
	ProductID uint      `json:"product_id"`
	SKUID     uint      `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SKUDeletedEvent SKU 删除事件
type SKUDeletedEvent struct {
	ProductID uint      `json:"product_id"`
	SKUID     uint      `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}
