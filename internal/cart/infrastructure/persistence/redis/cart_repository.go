package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

type cartRepository struct {
	client redis.UniversalClient
	prefix string
	ttl    time.Duration
}

// NewCartRepository 创建基于 Redis 的购物车仓储。
func NewCartRepository(client redis.UniversalClient) domain.CartRepository {
	return &cartRepository{
		client: client,
		prefix: "cart:",
		ttl:    7 * 24 * time.Hour, // 购物车保留 7 天
	}
}

func (r *cartRepository) Save(ctx context.Context, cart *domain.Cart) error {
	key := r.key(cart.UserID)
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.ttl).Err()
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.Cart, error) {
	key := r.key(userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return &domain.Cart{UserID: userID, Items: []*domain.CartItem{}}, nil
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

func (r *cartRepository) Delete(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *cartRepository) Clear(ctx context.Context, userID uint64) error {
	return r.Delete(ctx, userID)
}

func (r *cartRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", r.prefix, userID)
}
