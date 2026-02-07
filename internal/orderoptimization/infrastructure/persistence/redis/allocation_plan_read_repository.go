// 生成摘要：实现仓库分配计划读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

const allocationPlanPrefix = "orderoptimization:allocation:order:"

type allocationPlanReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAllocationPlanReadRepository 创建仓库分配计划读模型仓储。
func NewAllocationPlanReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AllocationPlanReadRepository {
	return &allocationPlanReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *allocationPlanReadRepository) Save(ctx context.Context, plan *domain.WarehouseAllocationPlan) error {
	if plan == nil {
		return nil
	}
	data, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(plan.OrderID), data, r.ttl).Err()
}

func (r *allocationPlanReadRepository) GetByOrderID(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	data, err := r.client.Get(ctx, r.key(orderID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var plan domain.WarehouseAllocationPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *allocationPlanReadRepository) DeleteByOrderID(ctx context.Context, orderID uint64) error {
	return r.client.Del(ctx, r.key(orderID)).Err()
}

func (r *allocationPlanReadRepository) key(orderID uint64) string {
	return fmt.Sprintf("%s%d", allocationPlanPrefix, orderID)
}
