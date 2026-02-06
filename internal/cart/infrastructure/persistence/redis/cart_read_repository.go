// 生成摘要：实现购物车读模型 Redis 仓储，提供按用户ID的快速读取。
// 假设：购物车以 user_id 为主键缓存。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

type cartReadRepository struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

// NewCartReadRepository 创建基于 Redis 的购物车读模型仓储。
func NewCartReadRepository(client redis.UniversalClient, ttl time.Duration) domain.CartReadRepository {
	return &cartReadRepository{
		client: client,
		prefix: "cart:detail:",
		ttl:    ttl,
	}
}

func (r *cartReadRepository) Save(ctx context.Context, cart *domain.Cart) error {
	if cart == nil {
		return nil
	}
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(cart.UserID), data, r.ttl).Err()
}

func (r *cartReadRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.Cart, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cart domain.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *cartReadRepository) Delete(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *cartReadRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", r.prefix, userID)
}
