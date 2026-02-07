// 生成摘要：实现拼团订单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

const groupbuyOrderDetailPrefix = "groupbuy:order:detail:"

type groupbuyOrderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewGroupbuyOrderReadRepository 创建拼团订单读模型仓储。
func NewGroupbuyOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.GroupbuyOrderReadRepository {
	return &groupbuyOrderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *groupbuyOrderReadRepository) Save(ctx context.Context, order *domain.GroupbuyOrder) error {
	if order == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(order.ID)), data, r.ttl).Err()
}

func (r *groupbuyOrderReadRepository) GetByID(ctx context.Context, id uint64) (*domain.GroupbuyOrder, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order domain.GroupbuyOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *groupbuyOrderReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *groupbuyOrderReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", groupbuyOrderDetailPrefix, id)
}
