// 生成摘要：实现指标读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

const metricDetailPrefix = "analytics:metric:detail:"

type metricReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewMetricReadRepository 创建指标读模型仓储。
func NewMetricReadRepository(client redis.UniversalClient, ttl time.Duration) domain.MetricReadRepository {
	return &metricReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *metricReadRepository) Save(ctx context.Context, metric *domain.Metric) error {
	if metric == nil {
		return nil
	}
	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(metric.ID), data, r.ttl).Err()
}

func (r *metricReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Metric, error) {
	data, err := r.client.Get(ctx, r.key(uint(id))).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var metric domain.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, err
	}
	return &metric, nil
}

func (r *metricReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(uint(id))).Err()
}

func (r *metricReadRepository) key(id uint) string {
	return fmt.Sprintf("%s%d", metricDetailPrefix, id)
}
