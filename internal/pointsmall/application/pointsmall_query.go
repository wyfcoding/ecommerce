package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

// PointsQuery handles read operations for pointsmall.
type PointsQuery struct {
	repo domain.PointsRepository
}

// NewPointsQuery creates a new PointsQuery instance.
func NewPointsQuery(repo domain.PointsRepository) *PointsQuery {
	return &PointsQuery{
		repo: repo,
	}
}

// ListProducts 获取积分商品列表。
func (q *PointsQuery) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*domain.PointsProduct, int64, error) {
	offset := (page - 1) * pageSize
	var prodStatus *domain.PointsProductStatus
	if status != nil {
		s := domain.PointsProductStatus(*status)
		prodStatus = &s
	}
	return q.repo.ListProducts(ctx, prodStatus, offset, pageSize)
}

// GetProduct 获取商品详情
func (q *PointsQuery) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	return q.repo.GetProduct(ctx, id)
}

// GetAccount 获取用户积分账户信息。
func (q *PointsQuery) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	return q.repo.GetAccount(ctx, userID)
}

// ListOrders 获取积分订单列表。
func (q *PointsQuery) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.PointsOrder, int64, error) {
	offset := (page - 1) * pageSize
	var orderStatus *domain.PointsOrderStatus
	if status != nil {
		s := domain.PointsOrderStatus(*status)
		orderStatus = &s
	}
	return q.repo.ListOrders(ctx, userID, orderStatus, offset, pageSize)
}
