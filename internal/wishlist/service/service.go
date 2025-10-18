package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"ecommerce/internal/wishlist/model"
	"ecommerce/internal/wishlist/repository"
	// 伪代码: 模拟 product 服务的 gRPC 客户端
	// productpb "ecommerce/gen/product/v1"
)

// WishlistService 定义了心愿单服务的业务逻辑接口
type WishlistService interface {
	AddItem(ctx context.Context, userID, productID uint) (*model.WishlistItem, error)
	RemoveItem(ctx context.Context, userID, productID uint) error
	ListItems(ctx context.Context, userID uint) ([]model.WishlistItem, error)
}

// wishlistService 是接口的具体实现
type wishlistService struct {
	repo   repository.WishlistRepository
	logger *zap.Logger
	// productClient productpb.ProductServiceClient
}

// NewWishlistService 创建一个新的 wishlistService 实例
func NewWishlistService(repo repository.WishlistRepository, logger *zap.Logger) WishlistService {
	return &wishlistService{repo: repo, logger: logger}
}

// AddItem 向心愿单添加一个商品
func (s *wishlistService) AddItem(ctx context.Context, userID, productID uint) (*model.WishlistItem, error) {
	s.logger.Info("Adding item to wishlist", zap.Uint("userID", userID), zap.Uint("productID", productID))

	// 1. 检查商品是否存在 (通过调用商品服务)
	// _, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: productID})
	// if err != nil {
	// 	 return nil, fmt.Errorf("商品不存在或商品服务不可用")
	// }

	// 2. 检查是否已在心愿单中
	exists, err := s.repo.ItemExists(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("商品已在您的心愿单中")
	}

	// 3. 添加到数据库
	return s.repo.AddItem(ctx, userID, productID)
}

// RemoveItem 从心愿单移除一个商品
func (s *wishlistService) RemoveItem(ctx context.Context, userID, productID uint) error {
	s.logger.Info("Removing item from wishlist", zap.Uint("userID", userID), zap.Uint("productID", productID))
	return s.repo.RemoveItem(ctx, userID, productID)
}

// ListItems 获取用户的心愿单列表
func (s *wishlistService) ListItems(ctx context.Context, userID uint) ([]model.WishlistItem, error) {
	s.logger.Info("Listing wishlist items", zap.Uint("userID", userID))
	items, err := s.repo.ListItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// (可选) 丰富商品信息
	// for i, item := range items {
	// 	 product, err := s.productClient.GetProduct(ctx, &productpb.GetProductRequest{Id: item.ProductID})
	// 	 if err == nil && product != nil {
	// 		 items[i].Product = product // 假设模型中有这个字段
	// 	 }
	// }

	return items, nil
}