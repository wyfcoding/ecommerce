package mysql

import (
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
	"gorm.io/gorm"
)

// WishlistModel 收藏夹写模型（持久化专用）。
type WishlistModel struct {
	gorm.Model
	UserID      uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:用户ID"`
	ProductID   uint64 `gorm:"index;not null;comment:商品ID"`
	SkuID       uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:SKU ID"`
	ProductName string `gorm:"type:varchar(255);not null;comment:商品名称"`
	SkuName     string `gorm:"type:varchar(255);not null;comment:SKU名称"`
	Price       uint64 `gorm:"not null;comment:价格(分)"`
	ImageURL    string `gorm:"type:varchar(255);comment:图片URL"`
}

func (WishlistModel) TableName() string {
	return "wishlists"
}

func toWishlistModel(item *domain.Wishlist) *WishlistModel {
	if item == nil {
		return nil
	}
	return &WishlistModel{
		Model: gorm.Model{
			ID:        uint(item.ID),
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		},
		UserID:      item.UserID,
		ProductID:   item.ProductID,
		SkuID:       item.SkuID,
		ProductName: item.ProductName,
		SkuName:     item.SkuName,
		Price:       item.Price,
		ImageURL:    item.ImageURL,
	}
}

func toWishlist(item *WishlistModel) *domain.Wishlist {
	if item == nil {
		return nil
	}
	return &domain.Wishlist{
		ID:          uint64(item.ID),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		UserID:      item.UserID,
		ProductID:   item.ProductID,
		SkuID:       item.SkuID,
		ProductName: item.ProductName,
		SkuName:     item.SkuName,
		Price:       item.Price,
		ImageURL:    item.ImageURL,
	}
}
