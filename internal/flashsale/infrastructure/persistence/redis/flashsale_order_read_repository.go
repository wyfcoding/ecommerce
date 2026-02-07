package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

const flashsaleOrderDetailPrefix = "flashsale:order:detail:"

type flashsaleOrderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewFlashsaleOrderReadRepository 创建秒杀订单读模型仓储。
func NewFlashsaleOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.FlashsaleOrderReadRepository {
	return &flashsaleOrderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *flashsaleOrderReadRepository) Save(ctx context.Context, order *domain.FlashsaleOrder) error {
	if order == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(order.ID)), data, r.ttl).Err()
}

func (r *flashsaleOrderReadRepository) GetByID(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order domain.FlashsaleOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *flashsaleOrderReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *flashsaleOrderReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", flashsaleOrderDetailPrefix, id)
}
