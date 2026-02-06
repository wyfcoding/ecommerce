package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistCommandService 处理收藏夹模块的写操作和核心业务逻辑。
type WishlistCommandService struct {
	repo      domain.WishlistRepository
	publisher domain.EventPublisher
	logger    *slog.Logger
}

// NewWishlistCommandService 创建并返回一个新的 WishlistCommandService 实例。
func NewWishlistCommandService(repo domain.WishlistRepository, publisher domain.EventPublisher, logger *slog.Logger) *WishlistCommandService {
	return &WishlistCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// AddToWishlist 将商品添加到收藏夹。
func (m *WishlistCommandService) AddToWishlist(ctx context.Context, userID, productID, skuID uint64, productName, skuName, imageURL string, price uint64) (*domain.Wishlist, error) {
	// 检查是否已存在。
	existing, err := m.repo.Get(ctx, userID, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Return existing if already there.
	}

	// 检查收藏夹条目数量限制（假设限制为100）。
	count, err := m.repo.Count(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 100 {
		return nil, fmt.Errorf("wishlist is full (max 100 items)")
	}

	item := &domain.Wishlist{
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		ProductName: productName,
		SkuName:     skuName,
		Price:       price,
		ImageURL:    imageURL,
	}

	if err := m.repo.Save(ctx, item); err != nil {
		m.logger.Error("failed to add to wishlist", "error", err, "user_id", userID, "sku_id", skuID)
		return nil, err
	}
	m.publishItemAdded(ctx, item)

	return item, nil
}

// RemoveFromWishlist 从收藏夹中移除指定商品。
func (m *WishlistCommandService) RemoveFromWishlist(ctx context.Context, userID, skuID uint64) error {
	if err := m.repo.DeleteByProduct(ctx, userID, skuID); err != nil {
		m.logger.Error("failed to remove from wishlist", "error", err, "user_id", userID, "sku_id", skuID)
		return err
	}
	m.publishItemRemoved(ctx, userID, skuID)
	return nil
}

// RemoveFromWishlistByID 从收藏夹中移除指定条目ID。
func (m *WishlistCommandService) RemoveFromWishlistByID(ctx context.Context, userID, id uint64) error {
	item, err := m.repo.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if err := m.repo.Delete(ctx, userID, id); err != nil {
		m.logger.Error("failed to remove from wishlist by id", "error", err, "user_id", userID, "id", id)
		return err
	}
	if item != nil {
		m.publishItemRemoved(ctx, userID, item.SkuID)
	}
	return nil
}

// ClearWishlist 清空用户的收藏夹。
func (m *WishlistCommandService) ClearWishlist(ctx context.Context, userID uint64) error {
	if err := m.repo.Clear(ctx, userID); err != nil {
		m.logger.Error("failed to clear wishlist", "error", err, "user_id", userID)
		return err
	}
	m.publishCleared(ctx, userID)
	return nil
}

func (m *WishlistCommandService) publishItemAdded(ctx context.Context, item *domain.Wishlist) {
	if m.publisher == nil || item == nil {
		return
	}
	event := &domain.WishlistItemAddedEvent{
		UserID:    item.UserID,
		SkuID:     item.SkuID,
		Timestamp: time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.WishlistItemAddedEventType, fmt.Sprintf("%d", item.UserID), event)
}

func (m *WishlistCommandService) publishItemRemoved(ctx context.Context, userID, skuID uint64) {
	if m.publisher == nil {
		return
	}
	event := &domain.WishlistItemRemovedEvent{
		UserID:    userID,
		SkuID:     skuID,
		Timestamp: time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.WishlistItemRemovedEventType, fmt.Sprintf("%d", userID), event)
}

func (m *WishlistCommandService) publishCleared(ctx context.Context, userID uint64) {
	if m.publisher == nil {
		return
	}
	event := &domain.WishlistClearedEvent{
		UserID:    userID,
		Timestamp: time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.WishlistClearedEventType, fmt.Sprintf("%d", userID), event)
}
