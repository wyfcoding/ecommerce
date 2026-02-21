package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

type OrderCacheRepo struct {
	client *redis.Client
}

func NewOrderCacheRepo(client *redis.Client) domain.OrderReadRepository {
	return &OrderCacheRepo{client: client}
}

func (r *OrderCacheRepo) GetOrderByID(ctx context.Context, id string) (*domain.Order, error) {
	val, err := r.client.Get(ctx, "order:"+id).Result()
	if err != nil {
		return nil, err
	}
	var order domain.Order
	json.Unmarshal([]byte(val), &order)
	return &order, nil
}

func (r *OrderCacheRepo) GetByID(ctx context.Context, orderID uint64, userID uint64) (*domain.Order, error) {
	return r.GetOrderByID(ctx, fmt.Sprintf("%d", orderID))
}

func (r *OrderCacheRepo) GetByOrderNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	val, err := r.client.Get(ctx, "orderNo:"+orderNo).Result()
	if err != nil {
		return nil, err
	}
	var order domain.Order
	json.Unmarshal([]byte(val), &order)
	return &order, nil
}

func (r *OrderCacheRepo) Save(ctx context.Context, order *domain.Order) error {
	data, _ := json.Marshal(order)
	return r.client.Set(ctx, "order:"+fmt.Sprintf("%d", order.ID), data, time.Hour*24).Err()
}

func (r *OrderCacheRepo) Delete(ctx context.Context, orderID uint64, userID uint64, orderNo string) error {
	return r.client.Del(ctx, "order:"+fmt.Sprintf("%d", orderID)).Err()
}
