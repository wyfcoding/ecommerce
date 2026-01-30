package domain

import "time"

// WishlistItemAddedEvent 添加商品到收藏夹事件。
type WishlistItemAddedEvent struct {
	UserID    uint64    `json:"user_id"`
	ProductID uint64    `json:"product_id"`
	SKUID     uint64    `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}

// WishlistItemRemovedEvent 从收藏夹移除商品事件。
type WishlistItemRemovedEvent struct {
	UserID    uint64    `json:"user_id"`
	ProductID uint64    `json:"product_id"`
	SKUID     uint64    `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}
