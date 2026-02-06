// 变更说明：查询侧改为优先读取 Redis/Elasticsearch，并保留 MySQL 回退与事件回放能力。
// 假设：读模型由事件驱动同步，ES 可按 CreatedAt 排序。
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/metrics"
)

// OrderQueryService 处理所有订单相关的查询操作 (Queries)。
type OrderQueryService struct {
	repo       domain.OrderRepository
	readRepo   domain.OrderReadRepository
	searchRepo domain.OrderSearchRepository
	eventStore domain.OrderEventStore
	logger     *slog.Logger
}

// NewOrderQueryService 构造函数。
func NewOrderQueryService(
	repo domain.OrderRepository,
	readRepo domain.OrderReadRepository,
	searchRepo domain.OrderSearchRepository,
	eventStore domain.OrderEventStore,
	logger *slog.Logger,
	_ *metrics.Metrics,
) *OrderQueryService {
	return &OrderQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		eventStore: eventStore,
		logger:     logger,
	}
}

// GetOrder 获取单个订单详情 (带简单缓存策略)。
func (s *OrderQueryService) GetOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Order, error) {
	if s.readRepo != nil {
		if order, err := s.readRepo.GetByID(ctx, userID, orderID); err == nil && order != nil {
			return order, nil
		}
	}

	o, err := s.repo.FindByID(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if o != nil {
		return o, nil
	}

	if s.eventStore != nil {
		aggregateID := fmt.Sprintf("%d", orderID)
		events, err := s.eventStore.LoadFromVersion(ctx, userID, aggregateID, 0)
		if err != nil {
			s.logger.WarnContext(ctx, "event store load failed", "order_id", orderID, "error", err)
			return nil, nil
		}
		if len(events) > 0 {
			order, rebuildErr := domain.RebuildOrderFromEvents(events)
			if rebuildErr != nil {
				s.logger.WarnContext(ctx, "order rebuild failed", "order_id", orderID, "error", rebuildErr)
				return nil, nil
			}
			return order, nil
		}
	}

	return nil, nil
}

// GetOrderByNo 通过编号获取订单。
func (s *OrderQueryService) GetOrderByNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	if s.readRepo != nil {
		if order, err := s.readRepo.GetByOrderNo(ctx, userID, orderNo); err == nil && order != nil {
			return order, nil
		}
	}

	if s.searchRepo != nil {
		if order, err := s.searchRepo.FindByOrderNo(ctx, orderNo); err == nil && order != nil {
			return order, nil
		}
	}

	order, err := s.repo.FindByOrderNo(ctx, userID, orderNo)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// ListOrders 分页查询。
func (s *OrderQueryService) ListOrders(ctx context.Context, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Order, int64, error) {
	if s.searchRepo != nil {
		list, total, err := s.searchRepo.Search(ctx, nil, status, offset, limit, startTime, endTime, sortBy)
		if err == nil {
			return list, total, nil
		}
		s.logger.WarnContext(ctx, "order search fallback to mysql", "error", err)
	}

	list, total, err := s.repo.List(ctx, status, offset, limit, startTime, endTime, sortBy)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListUserOrders 查询用户订单。
func (s *OrderQueryService) ListUserOrders(ctx context.Context, userID uint64, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Order, int64, error) {
	if s.searchRepo != nil {
		list, total, err := s.searchRepo.Search(ctx, &userID, status, offset, limit, startTime, endTime, sortBy)
		if err == nil {
			return list, total, nil
		}
		s.logger.WarnContext(ctx, "order search fallback to mysql", "user_id", userID, "error", err)
	}

	list, total, err := s.repo.ListByUserID(ctx, userID, status, offset, limit, startTime, endTime, sortBy)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
