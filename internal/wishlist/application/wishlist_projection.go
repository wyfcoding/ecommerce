// 生成摘要：新增收藏夹读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 user_id 为主键索引，写模型为最终一致性来源。
package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistProjectionService 负责将事件转换为读模型更新。
type WishlistProjectionService struct {
	repo       domain.WishlistRepository
	readRepo   domain.WishlistReadRepository
	searchRepo domain.WishlistSearchRepository
	logger     *slog.Logger
}

// NewWishlistProjectionService 创建收藏夹投影服务。
func NewWishlistProjectionService(repo domain.WishlistRepository, readRepo domain.WishlistReadRepository, searchRepo domain.WishlistSearchRepository, logger *slog.Logger) *WishlistProjectionService {
	return &WishlistProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnItemAdded 处理收藏夹新增事件。
func (s *WishlistProjectionService) OnItemAdded(ctx context.Context, event *domain.WishlistItemAddedEvent) error {
	if event == nil {
		return nil
	}
	item, err := s.repo.Get(ctx, event.UserID, event.SkuID)
	if err != nil {
		return err
	}
	if item == nil {
		return nil
	}
	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, item); err != nil {
			s.logger.ErrorContext(ctx, "failed to save wishlist read model", "user_id", event.UserID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, item); err != nil {
			s.logger.ErrorContext(ctx, "failed to index wishlist item", "user_id", event.UserID, "error", err)
			return err
		}
	}
	return nil
}

// OnItemRemoved 处理收藏夹移除事件。
func (s *WishlistProjectionService) OnItemRemoved(ctx context.Context, event *domain.WishlistItemRemovedEvent) error {
	if event == nil {
		return nil
	}
	if s.readRepo != nil {
		_ = s.readRepo.Delete(ctx, event.UserID, event.SkuID)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Delete(ctx, documentID(event.UserID, event.SkuID))
	}
	return nil
}

// OnCleared 处理收藏夹清空事件。
func (s *WishlistProjectionService) OnCleared(ctx context.Context, event *domain.WishlistClearedEvent) error {
	if event == nil {
		return nil
	}
	if s.readRepo != nil {
		_ = s.readRepo.Clear(ctx, event.UserID)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.DeleteByUser(ctx, event.UserID)
	}
	return nil
}

func documentID(userID, skuID uint64) string {
	return fmt.Sprintf("%d:%d", userID, skuID)
}
