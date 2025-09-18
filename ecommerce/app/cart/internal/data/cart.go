package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ecommerce/ecommerce/app/cart/internal/biz"

	"github.com/go-redis/redis/v8"
)

type cartRepo struct {
	rdb *redis.Client
}

// NewCartRepo .
func NewCartRepo(rdb *redis.Client) biz.CartRepo {
	return &cartRepo{rdb: rdb}
}

func (r *cartRepo) getCartKey(userID uint64) string {
	return fmt.Sprintf("cart:%d", userID)
}

func (r *cartRepo) GetItem(ctx context.Context, userID, skuID uint64) (*biz.CartItem, error) {
	key := r.getCartKey(userID)
	skuKey := fmt.Sprintf("%d", skuID)

	val, err := r.rdb.HGet(ctx, key, skuKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("item not found")
		}
		return nil, err
	}

	var item biz.CartItem
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *cartRepo) SaveItem(ctx context.Context, userID uint64, item *biz.CartItem) error {
	key := r.getCartKey(userID)
	skuKey := fmt.Sprintf("%d", item.SkuID)

	val, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if err := r.rdb.HSet(ctx, key, skuKey, val).Err(); err != nil {
		return err
	}
	// 可以为购物车设置一个过期时间，比如 30 天
	r.rdb.Expire(ctx, key, time.Hour*24*30)
	return nil
}

func (r *cartRepo) GetAllItems(ctx context.Context, userID uint64) ([]*biz.CartItem, error) {
	key := r.getCartKey(userID)
	// HGetAll 获取哈希中所有字段和值
	result, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	items := make([]*biz.CartItem, 0, len(result))
	for _, val := range result {
		var item biz.CartItem
		if err := json.Unmarshal([]byte(val), &item); err != nil {
			// 单个 item 解析失败，可以记录日志但继续处理其他 item
			continue
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *cartRepo) UpdateItem(ctx context.Context, userID uint64, item *biz.CartItem) error {
	// SaveItem 逻辑上可以复用
	return r.SaveItem(ctx, userID, item)
}

func (r *cartRepo) RemoveItems(ctx context.Context, userID uint64, skuIDs []uint64) error {
	key := r.getCartKey(userID)
	skuKeys := make([]string, len(skuIDs))
	for i, id := range skuIDs {
		skuKeys[i] = fmt.Sprintf("%d", id)
	}
	// HDel 可以一次删除多个字段
	return r.rdb.HDel(ctx, key, skuKeys...).Err()
}
