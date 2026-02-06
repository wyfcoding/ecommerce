package mysql

import (
	"github.com/wyfcoding/ecommerce/internal/cart/domain"
	"gorm.io/gorm"
)

// CartModel 购物车写模型。
type CartModel struct {
	gorm.Model
	UserID            uint64          `gorm:"not null;uniqueIndex;comment:用户ID"`
	AppliedCouponCode string          `gorm:"type:varchar(100);comment:已应用的优惠券码"`
	Items             []CartItemModel `gorm:"foreignKey:CartID"`
}

// CartItemModel 购物车项写模型。
type CartItemModel struct {
	gorm.Model
	CartID          uint64  `gorm:"not null;index;comment:购物车ID"`
	ProductID       string  `gorm:"not null;comment:商品ID"`
	SkuID           string  `gorm:"not null;index;comment:SKU ID"`
	ProductName     string  `gorm:"type:varchar(255);not null;comment:商品名称"`
	SkuName         string  `gorm:"type:varchar(255);not null;comment:SKU名称"`
	Price           float64 `gorm:"type:decimal(10,2);not null;comment:价格"`
	Quantity        int32   `gorm:"not null;comment:数量"`
	ProductImageURL string  `gorm:"type:varchar(255);comment:商品图片URL"`
}

func (CartModel) TableName() string {
	return "carts"
}

func (CartItemModel) TableName() string {
	return "cart_items"
}

func toCartModel(cart *domain.Cart) *CartModel {
	if cart == nil {
		return nil
	}
	model := &CartModel{
		Model: gorm.Model{
			ID:        uint(cart.ID),
			CreatedAt: cart.CreatedAt,
			UpdatedAt: cart.UpdatedAt,
		},
		UserID:            cart.UserID,
		AppliedCouponCode: cart.AppliedCouponCode,
	}
	if len(cart.Items) == 0 {
		return model
	}
	items := make([]CartItemModel, 0, len(cart.Items))
	for _, item := range cart.Items {
		if item == nil {
			continue
		}
		items = append(items, CartItemModel{
			Model: gorm.Model{
				ID:        uint(item.ID),
				CreatedAt: item.CreatedAt,
				UpdatedAt: item.UpdatedAt,
			},
			CartID:          cart.ID,
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			Price:           item.Price,
			Quantity:        item.Quantity,
			ProductImageURL: item.ProductImageURL,
		})
	}
	model.Items = items
	return model
}

func toCart(model *CartModel) *domain.Cart {
	if model == nil {
		return nil
	}
	cart := &domain.Cart{
		ID:                uint64(model.ID),
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		UserID:            model.UserID,
		AppliedCouponCode: model.AppliedCouponCode,
	}
	if len(model.Items) == 0 {
		return cart
	}
	items := make([]*domain.CartItem, 0, len(model.Items))
	for _, item := range model.Items {
		items = append(items, &domain.CartItem{
			ID:              uint64(item.ID),
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
			CartID:          item.CartID,
			ProductID:       item.ProductID,
			SkuID:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			Price:           item.Price,
			Quantity:        item.Quantity,
			ProductImageURL: item.ProductImageURL,
		})
	}
	cart.Items = items
	return cart
}
