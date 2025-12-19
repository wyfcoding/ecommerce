package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartQuery 处理购物车的读操作。
type CartQuery struct {
	repo   domain.CartRepository
	logger *slog.Logger
}

func NewCartQuery(repo domain.CartRepository, logger *slog.Logger) *CartQuery {
	return &CartQuery{
		repo:   repo,
		logger: logger,
	}
}

// GetCart 获取用户的购物车，如果不存在则创建。
func (s *CartQuery) GetCart(ctx context.Context, userID uint64) (*domain.Cart, error) {
	cart, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		cart = domain.NewCart(userID)
		if err := s.repo.Save(ctx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to create cart", "user_id", userID, "error", err)
			return nil, err
		}
	}
	return cart, nil
}
