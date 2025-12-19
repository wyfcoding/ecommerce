package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

type OrderQuery struct {
	repo domain.OrderRepository
}

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
