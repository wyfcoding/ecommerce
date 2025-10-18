package model

import (
	"time"
)

// WishlistItem 心愿单项目模型
// 每个记录代表一个用户收藏的一个商品
type WishlistItem struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_product" json:"user_id"` // 用户ID
	ProductID uint      `gorm:"not null;uniqueIndex:idx_user_product" json:"product_id"` // 商品ID
	AddedAt   time.Time `gorm:"not null" json:"added_at"` // 添加时间

	// 为了方便前端展示，可以冗余一些商品信息
	// 在实际查询时通过 Preload 或 JOIN 获取
	// Product   product.Product `gorm:"-" json:"product,omitempty"`
}

// TableName 自定义表名
func (WishlistItem) TableName() string {
	// 使用复合唯一索引 (idx_user_product) 来确保一个用户对一个商品只能收藏一次
	return "wishlist_items"
}
