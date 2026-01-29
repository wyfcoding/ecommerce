package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
)

// OrderQueryService 处理所有订单相关的查询操作 (Queries)。
type OrderQueryService struct {
	repo   domain.OrderRepository
	cache  cache.Cache
	logger *slog.Logger
}

// NewOrderQueryService 构造函数。
func NewOrderQueryService(repo domain.OrderRepository, cache cache.Cache, logger *slog.Logger, _ *metrics.Metrics) *OrderQueryService {
	return &OrderQueryService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// GetOrder 获取单个订单详情 (带简单缓存策略)。
func (s *OrderQueryService) GetOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Order, error) {
	cacheKey := fmt.Sprintf("order:detail:%d", orderID)
	var order domain.Order
	if err := s.cache.Get(ctx, cacheKey, &order); err == nil {
		return &order, nil
	}

	o, err := s.repo.FindByID(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if o != nil {
		_ = s.cache.Set(ctx, cacheKey, o, 0)
	}
	return o, nil
}

// GetOrderByNo 通过编号获取订单。
func (s *OrderQueryService) GetOrderByNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	return s.repo.FindByOrderNo(ctx, userID, orderNo)
}

// ListOrders 分页查询。
func (s *OrderQueryService) ListOrders(ctx context.Context, offset, limit int) ([]*domain.Order, int64, error) {
	return s.repo.List(ctx, offset, limit)
}

// ListUserOrders 查询用户订单。
func (s *OrderQueryService) ListUserOrders(ctx context.Context, userID uint64, status *int, offset, limit int) ([]*domain.Order, int64, error) {
	return s.repo.ListByUserID(ctx, userID, status, offset, limit)
}
