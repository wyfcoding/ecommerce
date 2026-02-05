// 生成摘要：实现订单读模型 Redis 仓储，提供按订单ID/订单号的快速读取。
// 假设：订单号与订单ID为全局唯一，缓存过期策略由调用方注入。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"github.com/redis/go-redis/v9"
)

const (
	orderDetailPrefix = "order:detail:"
	orderNoPrefix     = "order:no:"
)

// orderReadRepository 基于 Redis 的订单读模型仓储。
type orderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewOrderReadRepository 创建订单读模型仓储。
func NewOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.OrderReadRepository {
	return &orderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

// Save 保存或更新订单读模型。
func (r *orderReadRepository) Save(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return nil
	}

	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	orderIDKey := r.orderIDKey(uint64(order.ID))
	orderNoKey := r.orderNoKey(order.OrderNo)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, orderIDKey, data, r.ttl)
	pipe.Set(ctx, orderNoKey, fmt.Sprintf("%d", order.ID), r.ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// GetByID 根据订单ID获取读模型。
func (r *orderReadRepository) GetByID(ctx context.Context, _ uint64, orderID uint64) (*domain.Order, error) {
	data, err := r.client.Get(ctx, r.orderIDKey(orderID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOrderNo 根据订单号获取读模型。
func (r *orderReadRepository) GetByOrderNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	idStr, err := r.client.Get(ctx, r.orderNoKey(orderNo)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var orderID uint64
	if _, err := fmt.Sscanf(idStr, "%d", &orderID); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, userID, orderID)
}

// Delete 删除读模型数据。
func (r *orderReadRepository) Delete(ctx context.Context, _ uint64, orderID uint64, orderNo string) error {
	keys := []string{r.orderIDKey(orderID)}
	if orderNo != "" {
		keys = append(keys, r.orderNoKey(orderNo))
	}
	return r.client.Del(ctx, keys...).Err()
}

// orderIDKey 生成订单详情缓存键。
func (r *orderReadRepository) orderIDKey(orderID uint64) string {
	return fmt.Sprintf("%s%d", orderDetailPrefix, orderID)
}

// orderNoKey 生成订单号映射缓存键。
func (r *orderReadRepository) orderNoKey(orderNo string) string {
	return fmt.Sprintf("%s%s", orderNoPrefix, orderNo)
}
