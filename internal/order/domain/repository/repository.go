package repository

import (
	"context"
	"ecommerce/internal/order/domain/entity"
)

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uint64) (*entity.Order, error)
	GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error)
	List(ctx context.Context, userID uint64, status *entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error)
}
