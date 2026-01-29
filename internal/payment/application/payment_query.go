package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/cache"
	"github.com/wyfcoding/pkg/metrics"
	"golang.org/x/sync/singleflight"
)

// PaymentQuery 负责处理 Payment 相关的读操作和查询逻辑。
type PaymentQuery struct {
	repo        domain.PaymentRepository
	cache       cache.Cache
	sf          *singleflight.Group
	logger      *slog.Logger
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
}

// NewPaymentQuery 构造函数。
func NewPaymentQuery(repo domain.PaymentRepository, cache cache.Cache, logger *slog.Logger, m *metrics.Metrics) *PaymentQuery {
	return &PaymentQuery{
		repo:   repo,
		cache:  cache,
		sf:     &singleflight.Group{},
		logger: logger,
		cacheHits: m.NewCounterVec(&prometheus.CounterOpts{
			Name: "payment_query_cache_hits_total",
			Help: "Total number of payment cache hits",
		}, []string{"layer"}),
		cacheMisses: m.NewCounterVec(&prometheus.CounterOpts{
			Name: "payment_query_cache_misses_total",
			Help: "Total number of payment cache misses",
		}, []string{}),
	}
}

// GetPaymentStatus 获取支付状态 (支持缓存)
func (q *PaymentQuery) GetPaymentStatus(ctx context.Context, userID, id uint64) (*domain.Payment, error) {
	cacheKey := fmt.Sprintf("payment:%d:%d", userID, id)

	var payment domain.Payment
	if err := q.cache.Get(ctx, cacheKey, &payment); err == nil {
		q.cacheHits.WithLabelValues("L1").Inc()
		return &payment, nil
	}
	q.cacheMisses.WithLabelValues().Inc()

	val, err, _ := q.sf.Do(cacheKey, func() (any, error) {
		p, err := q.repo.FindByID(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		if p != nil {
			_ = q.cache.Set(context.Background(), cacheKey, p, 10*time.Minute)
		}
		return p, nil
	})

	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil // Not found
	}

	return val.(*domain.Payment), nil
}

// FindSuccessPaymentsByDate 查找指定日期的成功支付（对账用，不缓存或短缓存）
func (q *PaymentQuery) FindSuccessPaymentsByDate(ctx context.Context, date time.Time) ([]*domain.Payment, error) {
	// 对账查询通常走从库或专门的分析库，这里直接查 Repo
	return q.repo.FindSuccessPaymentsByDate(ctx, date)
}

// GetUserIDByPaymentNo 根据支付号查询用户 ID (用于分片路由)
func (q *PaymentQuery) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	return q.repo.GetUserIDByPaymentNo(ctx, paymentNo)
}
