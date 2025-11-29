package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/cart/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/cart/domain/repository"

	"log/slog"
)

type CartService struct {
	repo   repository.CartRepository
	logger *slog.Logger
}

func NewCartService(repo repository.CartRepository, logger *slog.Logger) *CartService {
	return &CartService{
		repo:   repo,
		logger: logger,
	}
}

// GetCart 获取购物车
func (s *CartService) GetCart(ctx context.Context, userID uint64) (*entity.Cart, error) {
	cart, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		cart = entity.NewCart(userID)
		if err := s.repo.Save(ctx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to create cart", "user_id", userID, "error", err)
			return nil, err
		}
		s.logger.InfoContext(ctx, "cart created successfully", "user_id", userID, "cart_id", cart.ID)
	}
	return cart, nil
}

// AddItem 添加商品到购物车
func (s *CartService) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AddItem(productID, skuID, productName, skuName, price, quantity, imageURL)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to add item to cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item added to cart successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// UpdateItemQuantity 更新购物车项数量
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.UpdateItemQuantity(skuID, quantity)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to update item quantity", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item quantity updated successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// RemoveItem 移除购物车项
func (s *CartService) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.RemoveItem(skuID)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove item from cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item removed from cart successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, userID uint64) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.Clear()
	if err := s.repo.Clear(ctx, uint64(cart.ID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear cart", "user_id", userID, "cart_id", cart.ID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "cart cleared successfully", "user_id", userID, "cart_id", cart.ID)
	return nil
}
