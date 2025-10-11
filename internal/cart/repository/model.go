package data

import (
	"gorm.io/gorm"
)

// Cart represents a user's shopping cart.
type Cart struct {
	gorm.Model
	UserID uint64 `gorm:"uniqueIndex;not null;comment:用户ID" json:"userId"`
	// Other cart-level information can be added here if needed, e.g., total price, total quantity
	// Items []CartItem `gorm:"foreignKey:CartID"` // GORM will handle this relationship
}

// CartItem represents an item in a user's shopping cart.
type CartItem struct {
	gorm.Model
	CartID   uint64 `gorm:"index;not null;comment:购物车ID" json:"cartId"`
	UserID   uint64 `gorm:"index;not null;comment:用户ID" json:"userId"` // Redundant but useful for direct queries
	SkuID    uint64 `gorm:"uniqueIndex:idx_user_sku;not null;comment:商品SKU ID" json:"skuId"`
	Quantity uint32 `gorm:"not null;default:1;comment:商品数量" json:"quantity"`
	Checked  bool   `gorm:"not null;default:true;comment:是否选中" json:"checked"`
	// Product details (title, image, price) will be fetched from product service and not stored here
}

// TableName specifies the table name for Cart.
func (Cart) TableName() string {
	return "carts"
}

// TableName specifies the table name for CartItem.
func (CartItem) TableName() string {
	return "cart_items"
}
