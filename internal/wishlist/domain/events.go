package domain

import "time"

const (
	WishlistItemAddedEventType   = "wishlist.item.added"
	WishlistItemRemovedEventType = "wishlist.item.removed"
	WishlistClearedEventType     = "wishlist.cleared"
)

// WishlistItemAddedEvent 收藏夹新增事件。
type WishlistItemAddedEvent struct {
	UserID    uint64    `json:"user_id"`
	SkuID     uint64    `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}

// WishlistItemRemovedEvent 收藏夹移除事件。
type WishlistItemRemovedEvent struct {
	UserID    uint64    `json:"user_id"`
	SkuID     uint64    `json:"sku_id"`
	Timestamp time.Time `json:"timestamp"`
}

// WishlistClearedEvent 收藏夹清空事件。
type WishlistClearedEvent struct {
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
