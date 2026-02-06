package domain

import (
	"time"
)

// Event types
const (
	ProductCreatedEventType  = "product.created"
	ProductUpdatedEventType  = "product.updated"
	ProductDeletedEventType  = "product.deleted"
	SKUAddedEventType        = "product.sku.added"
	SKUUpdatedEventType      = "product.sku.updated"
	SKUDeletedEventType      = "product.sku.deleted"
	BrandCreatedEventType    = "product.brand.created"
	BrandUpdatedEventType    = "product.brand.updated"
	BrandDeletedEventType    = "product.brand.deleted"
	CategoryCreatedEventType = "product.category.created"
	CategoryUpdatedEventType = "product.category.updated"
	CategoryDeletedEventType = "product.category.deleted"
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

// BrandCreatedEvent 品牌创建事件
type BrandCreatedEvent struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Logo      string    `json:"logo"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// BrandUpdatedEvent 品牌更新事件
type BrandUpdatedEvent struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Logo      string    `json:"logo"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// BrandDeletedEvent 品牌删除事件
type BrandDeletedEvent struct {
	ID        uint      `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

// CategoryCreatedEvent 分类创建事件
type CategoryCreatedEvent struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ParentID  uint      `json:"parent_id"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// CategoryUpdatedEvent 分类更新事件
type CategoryUpdatedEvent struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	ParentID  uint      `json:"parent_id"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// CategoryDeletedEvent 分类删除事件
type CategoryDeletedEvent struct {
	ID        uint      `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
