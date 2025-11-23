package application

import (
	"context"
	"ecommerce/internal/cart/domain/entity"
	"ecommerce/internal/cart/domain/repository"

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
			return nil, err
		}
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
	return s.repo.Save(ctx, cart)
}

// UpdateItemQuantity 更新购物车项数量
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.UpdateItemQuantity(skuID, quantity)
	return s.repo.Save(ctx, cart)
}

// RemoveItem 移除购物车项
func (s *CartService) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.RemoveItem(skuID)
	return s.repo.Save(ctx, cart)
}

// ClearCart 清空购物车
func (s *CartService) ClearCart(ctx context.Context, userID uint64) error {
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.Clear()
	return s.repo.Clear(ctx, uint64(cart.ID))
}
