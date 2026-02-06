package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartQueryService 处理购物车的读操作。
type CartQueryService struct {
	repo       domain.CartRepository
	readRepo   domain.CartReadRepository
	searchRepo domain.CartSearchRepository
	logger     *slog.Logger
}

// NewCartQueryService 负责处理购物车相关的读操作和查询逻辑。
func NewCartQueryService(repo domain.CartRepository, readRepo domain.CartReadRepository, searchRepo domain.CartSearchRepository, logger *slog.Logger) *CartQueryService {
	return &CartQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// GetCart 获取用户的购物车，如果不存在则创建。
func (s *CartQueryService) GetCart(ctx context.Context, userID uint64) (*domain.Cart, error) {
	if s.readRepo != nil {
		if cart, err := s.readRepo.GetByUserID(ctx, userID); err == nil && cart != nil {
			return cart, nil
		}
	}

	if s.searchRepo != nil {
		if cart, err := s.searchRepo.Search(ctx, userID); err == nil && cart != nil {
			return cart, nil
		}
	}

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

	if cart != nil && s.readRepo != nil {
		if err := s.readRepo.Save(ctx, cart); err != nil {
			s.logger.WarnContext(ctx, "failed to warm cart cache", "user_id", userID, "error", err)
		}
	}
	return cart, nil
}
