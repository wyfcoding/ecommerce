// 生成摘要：实现仪表板读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

const dashboardDetailPrefix = "analytics:dashboard:detail:"

type dashboardReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewDashboardReadRepository 创建仪表板读模型仓储。
func NewDashboardReadRepository(client redis.UniversalClient, ttl time.Duration) domain.DashboardReadRepository {
	return &dashboardReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *dashboardReadRepository) Save(ctx context.Context, dashboard *domain.Dashboard) error {
	if dashboard == nil {
		return nil
	}
	data, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(dashboard.ID), data, r.ttl).Err()
}

func (r *dashboardReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	data, err := r.client.Get(ctx, r.key(uint(id))).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var dashboard domain.Dashboard
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return nil, err
	}
	return &dashboard, nil
}

func (r *dashboardReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(uint(id))).Err()
}

func (r *dashboardReadRepository) key(id uint) string {
	return fmt.Sprintf("%s%d", dashboardDetailPrefix, id)
}
