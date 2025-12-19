package domain

import (
	"gorm.io/gorm" // 导入GORM库。
)

// Wishlist 实体是收藏夹模块的聚合根。
// 它代表用户收藏的一个商品（SKU）。
type Wishlist struct {
	gorm.Model         // 嵌入gorm.Model。
	UserID      uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:用户ID" json:"user_id"`  // 收藏商品的用户ID，与SKUID共同构成唯一索引。
	ProductID   uint64 `gorm:"index;not null;comment:商品ID" json:"product_id"`                  // 收藏的商品ID，索引字段。
	SkuID       uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:SKU ID" json:"sku_id"` // 收藏的SKU ID，与UserID共同构成唯一索引。
	ProductName string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`    // 商品名称。
	SkuName     string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`       // SKU名称。
	Price       uint64 `gorm:"not null;comment:价格(分)" json:"price"`                            // 收藏时的商品价格（单位：分）。
	ImageURL    string `gorm:"type:varchar(255);comment:图片URL" json:"image_url"`               // 商品图片URL。
}
