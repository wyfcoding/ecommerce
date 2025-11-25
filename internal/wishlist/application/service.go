package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/repository"

	"log/slog"
)

type WishlistService struct {
	repo   repository.WishlistRepository
	logger *slog.Logger
}

func NewWishlistService(repo repository.WishlistRepository, logger *slog.Logger) *WishlistService {
	return &WishlistService{
		repo:   repo,
		logger: logger,
	}
}

// Add 添加收藏
func (s *WishlistService) Add(ctx context.Context, userID, productID, skuID uint64, productName, skuName string, price uint64, imageURL string) (*entity.Wishlist, error) {
	// Check if already exists
	existing, err := s.repo.Get(ctx, userID, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Already exists, return it
	}

	wishlist := &entity.Wishlist{
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		ProductName: productName,
		SkuName:     skuName,
		Price:       price,
		ImageURL:    imageURL,
	}

	if err := s.repo.Save(ctx, wishlist); err != nil {
		return nil, err
	}
	return wishlist, nil
}

// Remove 移除收藏
func (s *WishlistService) Remove(ctx context.Context, userID, id uint64) error {
	return s.repo.Delete(ctx, userID, id)
}

// List 获取收藏列表
func (s *WishlistService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Wishlist, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, userID, offset, pageSize)
}

// CheckStatus 检查收藏状态
func (s *WishlistService) CheckStatus(ctx context.Context, userID, skuID uint64) (bool, error) {
	item, err := s.repo.Get(ctx, userID, skuID)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}
