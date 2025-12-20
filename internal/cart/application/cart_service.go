package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartService 是购物车应用服务的门面。
type CartService struct {
	Manager *CartManager
	Query   *CartQuery
}

// NewCartService 定义了 NewCart 相关的服务逻辑。
func NewCartService(manager *CartManager, query *CartQuery) *CartService {
	return &CartService{
		Manager: manager,
		Query:   query,
	}
}

func (s *CartService) GetCart(ctx context.Context, userID uint64) (*domain.Cart, error) {
	return s.Query.GetCart(ctx, userID)
}

func (s *CartService) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	return s.Manager.AddItem(ctx, userID, productID, skuID, productName, skuName, price, quantity, imageURL)
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	return s.Manager.UpdateItemQuantity(ctx, userID, skuID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	return s.Manager.RemoveItem(ctx, userID, skuID)
}

func (s *CartService) ClearCart(ctx context.Context, userID uint64) error {
	return s.Manager.ClearCart(ctx, userID)
}

func (s *CartService) MergeCarts(ctx context.Context, sourceUserID, targetUserID uint64) error {
	return s.Manager.MergeCarts(ctx, sourceUserID, targetUserID)
}

func (s *CartService) ApplyCoupon(ctx context.Context, userID uint64, couponCode string) error {
	return s.Manager.ApplyCoupon(ctx, userID, couponCode)
}

func (s *CartService) RemoveCoupon(ctx context.Context, userID uint64) error {
	return s.Manager.RemoveCoupon(ctx, userID)
}
