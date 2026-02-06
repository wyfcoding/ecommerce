package domain

import "time"

// Wishlist 实体是收藏夹模块的聚合根。
// 它代表用户收藏的一个商品（SKU）。
type Wishlist struct {
	ID          uint64    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserID      uint64    `json:"user_id"`      // 收藏商品的用户ID。
	ProductID   uint64    `json:"product_id"`   // 收藏的商品ID。
	SkuID       uint64    `json:"sku_id"`       // 收藏的SKU ID。
	ProductName string    `json:"product_name"` // 商品名称。
	SkuName     string    `json:"sku_name"`     // SKU名称。
	Price       uint64    `json:"price"`        // 收藏时的商品价格（单位：分）。
	ImageURL    string    `json:"image_url"`    // 商品图片URL。
}
