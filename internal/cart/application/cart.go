package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartService 门面服务，整合 CommandService 和 Query。
type CartService struct {
	command *CartCommandService
	query   *CartQuery
}

// NewCartService 构造函数。
func NewCartService(command *CartCommandService, query *CartQuery) *CartService {
	return &CartService{
		command: command,
		query:   query,
	}
}

// --- Commands (Writes) ---

func (s *CartService) AddItem(ctx context.Context, userID uint64, productID, skuID string, productName, skuName string, price float64, quantity int32, imageURL string) error {
	return s.command.AddItem(ctx, userID, productID, skuID, productName, skuName, price, quantity, imageURL)
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID string, quantity int32) error {
	return s.command.UpdateItemQuantity(ctx, userID, skuID, quantity)
}

func (s *CartService) RemoveItem(ctx context.Context, userID uint64, skuID string) error {
	return s.command.RemoveItem(ctx, userID, skuID)
}

func (s *CartService) RemoveItems(ctx context.Context, userID uint64, skuIDs []string) error {
	return s.command.RemoveItems(ctx, userID, skuIDs)
}

func (s *CartService) ClearCart(ctx context.Context, userID uint64) error {
	return s.command.ClearCart(ctx, userID)
}

func (s *CartService) MergeCarts(ctx context.Context, sourceUserID, targetUserID uint64) error {
	return s.command.MergeCarts(ctx, sourceUserID, targetUserID)
}

func (s *CartService) ApplyCoupon(ctx context.Context, userID uint64, couponCode string) error {
	return s.command.ApplyCoupon(ctx, userID, couponCode)
}

func (s *CartService) RemoveCoupon(ctx context.Context, userID uint64) error {
	return s.command.RemoveCoupon(ctx, userID)
}

// --- Query (Reads) ---

func (s *CartService) GetCart(ctx context.Context, userID uint64) (*domain.Cart, error) {
	return s.query.GetCart(ctx, userID)
}
