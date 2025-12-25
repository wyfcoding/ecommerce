package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// Cart 是购物车应用服务的门面。
type Cart struct {
	Manager *CartManager
	Query   *CartQuery
}

// NewCart 创建购物车服务实例。
func NewCart(manager *CartManager, query *CartQuery) *Cart {
	return &Cart{
		Manager: manager,
		Query:   query,
	}
}

// GetCart 获取购物车。
func (s *Cart) GetCart(ctx context.Context, userID uint64) (*domain.Cart, error) {
	return s.Query.GetCart(ctx, userID)
}

// AddItem 添加商品到购物车。
func (s *Cart) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	return s.Manager.AddItem(ctx, userID, productID, skuID, productName, skuName, price, quantity, imageURL)
}

// UpdateItemQuantity 更新购物车商品数量。
func (s *Cart) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	return s.Manager.UpdateItemQuantity(ctx, userID, skuID, quantity)
}

// RemoveItem 从购物车移除商品。
func (s *Cart) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	return s.Manager.RemoveItem(ctx, userID, skuID)
}

// ClearCart 清空购物车。
func (s *Cart) ClearCart(ctx context.Context, userID uint64) error {
	return s.Manager.ClearCart(ctx, userID)
}

// MergeCarts 合并购物车。
func (s *Cart) MergeCarts(ctx context.Context, sourceUserID, targetUserID uint64) error {
	return s.Manager.MergeCarts(ctx, sourceUserID, targetUserID)
}

// ApplyCoupon 应用优惠券。
func (s *Cart) ApplyCoupon(ctx context.Context, userID uint64, couponCode string) error {
	return s.Manager.ApplyCoupon(ctx, userID, couponCode)
}

// RemoveCoupon 移除优惠券。
func (s *Cart) RemoveCoupon(ctx context.Context, userID uint64) error {
	return s.Manager.RemoveCoupon(ctx, userID)
}
