// 生成摘要：新增购物车读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 user_id 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartProjectionService 负责将事件转换为读模型更新。
type CartProjectionService struct {
	repo       domain.CartRepository
	readRepo   domain.CartReadRepository
	searchRepo domain.CartSearchRepository
	logger     *slog.Logger
}

// NewCartProjectionService 创建购物车投影服务。
func NewCartProjectionService(repo domain.CartRepository, readRepo domain.CartReadRepository, searchRepo domain.CartSearchRepository, logger *slog.Logger) *CartProjectionService {
	return &CartProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnItemAdded 处理商品添加事件。
func (s *CartProjectionService) OnItemAdded(ctx context.Context, event *domain.CartItemAddedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

// OnItemUpdated 处理商品更新事件。
func (s *CartProjectionService) OnItemUpdated(ctx context.Context, event *domain.CartItemUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

// OnItemRemoved 处理商品移除事件。
func (s *CartProjectionService) OnItemRemoved(ctx context.Context, event *domain.CartItemRemovedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

// OnCleared 处理购物车清空事件。
func (s *CartProjectionService) OnCleared(ctx context.Context, event *domain.CartClearedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

// OnMerged 处理购物车合并事件。
func (s *CartProjectionService) OnMerged(ctx context.Context, event *domain.CartsMergedEvent) error {
	if event == nil {
		return nil
	}
	if err := s.refreshCart(ctx, event.TargetUserID); err != nil {
		return err
	}
	return s.refreshCart(ctx, event.SourceUserID)
}

// OnCouponApplied 处理优惠券应用事件。
func (s *CartProjectionService) OnCouponApplied(ctx context.Context, event *domain.CouponAppliedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

// OnCouponRemoved 处理优惠券移除事件。
func (s *CartProjectionService) OnCouponRemoved(ctx context.Context, event *domain.CouponRemovedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCart(ctx, event.UserID)
}

func (s *CartProjectionService) refreshCart(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return nil
	}
	cart, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load cart for projection", "user_id", userID, "error", err)
		return err
	}
	if cart == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, userID)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, userID)
		}
		return nil
	}
	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to save cart read model", "user_id", userID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to index cart", "user_id", userID, "error", err)
			return err
		}
	}
	return nil
}
