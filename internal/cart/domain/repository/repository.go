package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/cart/domain/entity"
)

// CartRepository 购物车仓储接口
type CartRepository interface {
	Save(ctx context.Context, cart *entity.Cart) error
	GetByUserID(ctx context.Context, userID uint64) (*entity.Cart, error)
	Delete(ctx context.Context, id uint64) error
	Clear(ctx context.Context, cartID uint64) error
}
