package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity"     // 导入收藏夹领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/repository" // 导入收藏夹领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// WishlistService 结构体定义了收藏夹管理相关的应用服务。
// 它协调领域层和基础设施层，处理收藏夹中商品的添加、移除、列表查询和状态检查等业务逻辑。
type WishlistService struct {
	repo   repository.WishlistRepository // 依赖WishlistRepository接口，用于数据持久化操作。
	logger *slog.Logger                  // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewWishlistService 创建并返回一个新的 WishlistService 实例。
func NewWishlistService(repo repository.WishlistRepository, logger *slog.Logger) *WishlistService {
	return &WishlistService{
		repo:   repo,
		logger: logger,
	}
}

// Add 将商品添加到用户的收藏夹。
// 如果商品（SKU）已存在于收藏夹中，则直接返回现有记录，不创建重复项。
// ctx: 上下文。
// userID: 用户ID。
// productID: 商品ID。
// skuID: SKU ID。
// productName: 商品名称。
// skuName: SKU名称。
// price: 价格。
// imageURL: 图片URL。
// 返回收藏夹实体和可能发生的错误。
func (s *WishlistService) Add(ctx context.Context, userID, productID, skuID uint64, productName, skuName string, price uint64, imageURL string) (*entity.Wishlist, error) {
	// 1. 检查商品（SKU）是否已存在于用户的收藏夹中。
	existing, err := s.repo.Get(ctx, userID, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing wishlist item", "user_id", userID, "sku_id", skuID, "error", err)
		return nil, err
	}
	if existing != nil {
		s.logger.InfoContext(ctx, "item already in wishlist", "user_id", userID, "sku_id", skuID)
		return existing, nil // 如果已存在，直接返回现有记录。
	}

	// 2. 创建收藏夹实体。
	wishlist := &entity.Wishlist{
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		ProductName: productName,
		SkuName:     skuName,
		Price:       price,
		ImageURL:    imageURL,
	}

	// 3. 通过仓储接口保存收藏夹实体。
	if err := s.repo.Save(ctx, wishlist); err != nil {
		s.logger.ErrorContext(ctx, "failed to add to wishlist", "user_id", userID, "sku_id", skuID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "added to wishlist successfully", "user_id", userID, "sku_id", skuID)
	return wishlist, nil
}

// Remove 将商品从用户的收藏夹中移除。
// ctx: 上下文。
// userID: 用户ID。
// id: 收藏夹条目ID。
// 返回可能发生的错误。
func (s *WishlistService) Remove(ctx context.Context, userID, id uint64) error {
	if err := s.repo.Delete(ctx, userID, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove from wishlist", "user_id", userID, "wishlist_item_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "removed from wishlist successfully", "user_id", userID, "wishlist_item_id", id)
	return nil
}

// RemoveByProduct 将指定商品从用户的收藏夹中移除。
// ctx: 上下文。
// userID: 用户ID。
// skuID: SKU ID。
// 返回可能发生的错误。
func (s *WishlistService) RemoveByProduct(ctx context.Context, userID, skuID uint64) error {
	if err := s.repo.DeleteByProduct(ctx, userID, skuID); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove from wishlist by product", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "removed from wishlist by product successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// List 获取指定用户的收藏夹列表。
// ctx: 上下文。
// userID: 用户ID。
// page, pageSize: 分页参数。
// 返回收藏夹实体列表、总数和可能发生的错误。
func (s *WishlistService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Wishlist, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, userID, offset, pageSize)
}

// CheckStatus 检查指定商品（SKU）是否已在用户的收藏夹中。
// ctx: 上下文。
// userID: 用户ID。
// skuID: SKU ID。
// 返回布尔值（是否在收藏夹中）和可能发生的错误。
func (s *WishlistService) CheckStatus(ctx context.Context, userID, skuID uint64) (bool, error) {
	item, err := s.repo.Get(ctx, userID, skuID)
	if err != nil {
		return false, err
	}
	return item != nil, nil // 如果 item 不为nil，则表示商品在收藏夹中。
}

// Clear 清空指定用户的收藏夹。
// ctx: 上下文。
// userID: 用户ID。
// 返回可能发生的错误。
func (s *WishlistService) Clear(ctx context.Context, userID uint64) error {
	if err := s.repo.Clear(ctx, userID); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear wishlist", "user_id", userID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "wishlist cleared successfully", "user_id", userID)
	return nil
}
