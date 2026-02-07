// 生成摘要：实现拆分订单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

const splitOrderListPrefix = "orderoptimization:split:order:"

type splitOrderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSplitOrderReadRepository 创建拆分订单读模型仓储。
func NewSplitOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SplitOrderReadRepository {
	return &splitOrderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *splitOrderReadRepository) Save(ctx context.Context, originalOrderID uint64, orders []*domain.SplitOrder) error {
	if orders == nil {
		return nil
	}
	data, err := json.Marshal(orders)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(originalOrderID), data, r.ttl).Err()
}

func (r *splitOrderReadRepository) GetByOriginalOrderID(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	data, err := r.client.Get(ctx, r.key(originalOrderID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var orders []*domain.SplitOrder
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *splitOrderReadRepository) DeleteByOriginalOrderID(ctx context.Context, originalOrderID uint64) error {
	return r.client.Del(ctx, r.key(originalOrderID)).Err()
}

func (r *splitOrderReadRepository) key(originalOrderID uint64) string {
	return fmt.Sprintf("%s%d", splitOrderListPrefix, originalOrderID)
}
