package entity

import (
	"gorm.io/gorm"
)

// Wishlist 收藏实体
type Wishlist struct {
	gorm.Model
	UserID      uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:用户ID" json:"user_id"`
	ProductID   uint64 `gorm:"index;not null;comment:商品ID" json:"product_id"`
	SkuID       uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:SKU ID" json:"sku_id"`
	ProductName string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`
	SkuName     string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`
	Price       uint64 `gorm:"not null;comment:价格(分)" json:"price"`
	ImageURL    string `gorm:"type:varchar(255);comment:图片URL" json:"image_url"`
}
