// 生成摘要：实现积分订单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

const pointsOrderDetailPrefix = "pointsmall:order:detail:"

type orderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewOrderReadRepository 创建积分订单读模型仓储。
func NewOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PointsOrderReadRepository {
	return &orderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *orderReadRepository) Save(ctx context.Context, order *domain.PointsOrder) error {
	if order == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(order.ID)), data, r.ttl).Err()
}

func (r *orderReadRepository) GetByID(ctx context.Context, id uint64) (*domain.PointsOrder, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order domain.PointsOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *orderReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", pointsOrderDetailPrefix, id)
}
