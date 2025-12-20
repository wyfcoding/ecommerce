package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// OrderQuery 负责处理 Order 相关的读操作和查询逻辑。
type OrderQuery struct {
	repo domain.OrderRepository
}

// NewOrderQuery 负责处理 NewOrder 相关的读操作和查询逻辑。
func NewOrderQuery(repo domain.OrderRepository) *OrderQuery {
	return &OrderQuery{
		repo: repo,
	}
}

// GetOrder 获取指定ID的订单详情。
func (s *OrderQuery) GetOrder(ctx context.Context, id uint64) (*domain.Order, error) {
	return s.repo.FindByID(ctx, uint(id))
}

// ListOrders 获取订单列表。
func (s *OrderQuery) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	offset := (page - 1) * pageSize

	if userID > 0 {
		return s.repo.ListByUserID(ctx, uint(userID), offset, pageSize)
	}
	return s.repo.List(ctx, offset, pageSize)
}
