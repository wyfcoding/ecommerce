package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
	"golang.org/x/sync/singleflight"
)

// OrderQuery 负责处理 Order 相关的读操作和查询逻辑。
type OrderQuery struct {
	repo        domain.OrderRepository
	cache       cache.Cache
	logger      *slog.Logger
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
	sf          singleflight.Group
}

// NewOrderQuery 创建 Order 查询服务实例。
func NewOrderQuery(
	repo domain.OrderRepository,
	cache cache.Cache,
	logger *slog.Logger,
	m *metrics.Metrics,
) *OrderQuery {
	cacheHits := m.NewCounterVec(&prometheus.CounterOpts{
		Name: "order_cache_hits_total",
		Help: "订单缓存命中总数",
	}, []string{"layer"})

	cacheMisses := m.NewCounterVec(&prometheus.CounterOpts{
		Name: "order_cache_misses_total",
		Help: "订单缓存未命中总数",
	}, []string{})

	return &OrderQuery{
		repo:        repo,
		cache:       cache,
		logger:      logger,
		cacheHits:   cacheHits,
		cacheMisses: cacheMisses,
	}
}

// GetOrder 获取指定ID的订单详情。
func (q *OrderQuery) GetOrder(ctx context.Context, userID, id uint64) (*domain.Order, error) {
	cacheKey := fmt.Sprintf("order:%d:%d", userID, id)

	var order domain.Order
	if err := q.cache.Get(ctx, cacheKey, &order); err == nil {
		q.cacheHits.WithLabelValues("L1").Inc()
		return &order, nil
	}
	q.cacheMisses.WithLabelValues().Inc()

	val, err, _ := q.sf.Do(cacheKey, func() (any, error) {
		o, err := q.repo.FindByID(ctx, userID, uint(id))
		if err != nil {
			return nil, err
		}
		if o != nil {
			_ = q.cache.Set(context.Background(), cacheKey, o, 10*time.Minute)
		}
		return o, nil
	})

	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil // Not found
	}

	return val.(*domain.Order), nil
}

// ListOrders 获取订单列表。
func (q *OrderQuery) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Order, int64, error) {
	offset := (page - 1) * pageSize

	// List queries usually not cached directly unless highly frequent and static.
	// For orders, we can query DB directly or use list cache with invalidation.
	// Staying simple: DB query.

	if userID > 0 {
		return q.repo.ListByUserID(ctx, uint(userID), offset, pageSize)
	}
	return q.repo.List(ctx, offset, pageSize)
}

// Additional helpers if needed
func (q *OrderQuery) GetOrderByNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	return q.repo.FindByOrderNo(ctx, userID, orderNo)
}
