package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity"
)

// WishlistRepository 收藏仓储接口
type WishlistRepository interface {
	Save(ctx context.Context, wishlist *entity.Wishlist) error
	Delete(ctx context.Context, userID, id uint64) error
	Get(ctx context.Context, userID, skuID uint64) (*entity.Wishlist, error)
	List(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Wishlist, int64, error)
	Count(ctx context.Context, userID uint64) (int64, error)
}
