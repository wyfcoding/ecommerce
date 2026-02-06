// 生成摘要：实现收藏夹读模型 Redis 仓储，提供按用户的快速读取。
// 假设：收藏夹以 user_id 为维度维护索引，条目以 sku_id 为键存储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

const (
	wishlistItemPrefix = "wishlist:item:"
	wishlistZSetPrefix = "wishlist:zset:"
)

type wishlistReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewWishlistReadRepository 创建收藏夹读模型仓储。
func NewWishlistReadRepository(client redis.UniversalClient, ttl time.Duration) domain.WishlistReadRepository {
	return &wishlistReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *wishlistReadRepository) Save(ctx context.Context, item *domain.Wishlist) error {
	if item == nil {
		return nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	score := float64(time.Now().Unix())
	if !item.CreatedAt.IsZero() {
		score = float64(item.CreatedAt.Unix())
	}

	itemKey := r.itemKey(item.UserID, item.SkuID)
	zsetKey := r.zsetKey(item.UserID)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, itemKey, data, r.ttl)
	pipe.ZAdd(ctx, zsetKey, redis.Z{Score: score, Member: fmt.Sprintf("%d", item.SkuID)})
	pipe.Expire(ctx, zsetKey, r.ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *wishlistReadRepository) Delete(ctx context.Context, userID, skuID uint64) error {
	itemKey := r.itemKey(userID, skuID)
	zsetKey := r.zsetKey(userID)
	pipe := r.client.Pipeline()
	pipe.Del(ctx, itemKey)
	pipe.ZRem(ctx, zsetKey, fmt.Sprintf("%d", skuID))
	_, err := pipe.Exec(ctx)
	return err
}

func (r *wishlistReadRepository) Get(ctx context.Context, userID, skuID uint64) (*domain.Wishlist, error) {
	data, err := r.client.Get(ctx, r.itemKey(userID, skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var item domain.Wishlist
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *wishlistReadRepository) List(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Wishlist, int64, error) {
	zsetKey := r.zsetKey(userID)
	total, err := r.client.ZCard(ctx, zsetKey).Result()
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 10
	}

	skus, err := r.client.ZRevRange(ctx, zsetKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, err
	}
	if len(skus) == 0 {
		return nil, total, nil
	}

	keys := make([]string, 0, len(skus))
	for _, sku := range skus {
		keys = append(keys, r.itemKey(userID, toUint64(sku)))
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Wishlist, 0, len(values))
	for _, v := range values {
		if v == nil {
			continue
		}
		bytes, ok := v.(string)
		if ok {
			var item domain.Wishlist
			if err := json.Unmarshal([]byte(bytes), &item); err == nil {
				items = append(items, &item)
			}
			continue
		}
		if data, ok := v.([]byte); ok {
			var item domain.Wishlist
			if err := json.Unmarshal(data, &item); err == nil {
				items = append(items, &item)
			}
		}
	}

	return items, total, nil
}

func (r *wishlistReadRepository) Clear(ctx context.Context, userID uint64) error {
	zsetKey := r.zsetKey(userID)
	skus, err := r.client.ZRange(ctx, zsetKey, 0, -1).Result()
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	for _, sku := range skus {
		pipe.Del(ctx, r.itemKey(userID, toUint64(sku)))
	}
	pipe.Del(ctx, zsetKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *wishlistReadRepository) itemKey(userID, skuID uint64) string {
	return fmt.Sprintf("%s%d:%d", wishlistItemPrefix, userID, skuID)
}

func (r *wishlistReadRepository) zsetKey(userID uint64) string {
	return fmt.Sprintf("%s%d", wishlistZSetPrefix, userID)
}

func toUint64(val string) uint64 {
	var id uint64
	_, _ = fmt.Sscanf(val, "%d", &id)
	return id
}
