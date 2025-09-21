package data

import (
	"context"
	"ecommerce/internal/cart/biz"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type cartRepo struct {
	*Data
}

// NewCartRepo 是 cartRepo 的构造函数。
func NewCartRepo(data *Data) biz.CartRepo {
	return &cartRepo{Data: data}
}

// cartItemsKey 生成用户购物车商品数量的 redis key。
func (r *cartRepo) cartItemsKey(userID uint64) string {
	return fmt.Sprintf("cart:%d:items", userID)
}

// cartCheckedKey 生成用户购物车商品勾选状态的 redis key。
func (r *cartRepo) cartCheckedKey(userID uint64) string {
	return fmt.Sprintf("cart:%d:checked", userID)
}

// AddItem 向购物车中添加指定数量的商品，并默认设置为勾选状态。
func (r *cartRepo) AddItem(ctx context.Context, userID, skuID uint64, quantity uint32) error {
	pipe := r.rdb.Pipeline()
	keyItems := r.cartItemsKey(userID)
	keyChecked := r.cartCheckedKey(userID)

	pipe.HIncrBy(ctx, keyItems, strconv.FormatUint(skuID, 10), int64(quantity))
	pipe.HSet(ctx, keyChecked, strconv.FormatUint(skuID, 10), "1") // 默认勾选

	_, err := pipe.Exec(ctx)
	return err
}

// GetCart 获取用户购物车中所有商品的skuID及数量和勾选状态。
func (r *cartRepo) GetCart(ctx context.Context, userID uint64) ([]*biz.CartItem, error) {
	keyItems := r.cartItemsKey(userID)
	keyChecked := r.cartCheckedKey(userID)

	pipe := r.rdb.Pipeline()
	itemsCmd := pipe.HGetAll(ctx, keyItems)
	checkedCmd := pipe.HGetAll(ctx, keyChecked)
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	itemResults, err := itemsCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	checkedResults, err := checkedCmd.Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	cartItems := make([]*biz.CartItem, 0, len(itemResults))
	for skuIDStr, quantityStr := range itemResults {
		skuID, _ := strconv.ParseUint(skuIDStr, 10, 64)
		quantity, _ := strconv.ParseUint(quantityStr, 10, 32)
		if skuID > 0 {
			_, checked := checkedResults[skuIDStr] // 检查是否存在于 checkedResults 中
			cartItems = append(cartItems, &biz.CartItem{
				SkuID:    skuID,
				Quantity: uint32(quantity),
				Checked:  checked,
			})
		}
	}
	return cartItems, nil
}

// UpdateItem 更新购物车中商品的数量。
func (r *cartRepo) UpdateItem(ctx context.Context, userID, skuID uint64, quantity uint32) error {
	if quantity == 0 {
		return r.RemoveItem(ctx, userID, skuID)
	}
	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, r.cartItemsKey(userID), strconv.FormatUint(skuID, 10), quantity)
	// 默认更新时保持勾选状态为 true
	pipe.HSet(ctx, r.cartCheckedKey(userID), strconv.FormatUint(skuID, 10), "1")
	_, err := pipe.Exec(ctx)
	return err
}

// RemoveItem 从购物车中移除一个或多个商品。
func (r *cartRepo) RemoveItem(ctx context.Context, userID uint64, skuIDs ...uint64) error {
	if len(skuIDs) == 0 {
		return nil
	}
	// 将 skuIDs 转换为字符串切片
	skuIDStrs := make([]string, len(skuIDs))
	for i, id := range skuIDs {
		skuIDStrs[i] = strconv.FormatUint(id, 10)
	}

	// 使用 pipeline 原子化地删除两个 hash 中的字段
	pipe := r.rdb.Pipeline()
	pipe.HDel(ctx, r.cartItemsKey(userID), skuIDStrs...)
	pipe.HDel(ctx, r.cartCheckedKey(userID), skuIDStrs...)
	_, err := pipe.Exec(ctx)
	return err
}

// UpdateCheckStatus 更新一个或多个商品的勾选状态。
func (r *cartRepo) UpdateCheckStatus(ctx context.Context, userID uint64, skuIDs []uint64, checked bool) error {
	key := r.cartCheckedKey(userID)
	// 如果是勾选，则批量设置；如果是取消勾选，则批量删除。
	if checked {
		fields := make(map[string]interface{}, len(skuIDs))
		for _, id := range skuIDs {
			fields[strconv.FormatUint(id, 10)] = "1"
		}
		return r.rdb.HSet(ctx, key, fields).Err()
	} else {
		skuIDStrs := make([]string, len(skuIDs))
		for i, id := range skuIDs {
			skuIDStrs[i] = strconv.FormatUint(id, 10)
		}
		return r.rdb.HDel(ctx, key, skuIDStrs...).Err()
	}
}

// GetCheckStatus 获取购物车中所有商品的勾选状态。
func (r *cartRepo) GetCheckStatus(ctx context.Context, userID uint64) (map[uint64]bool, error) {
	key := r.cartCheckedKey(userID)
	result, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	checkedStatus := make(map[uint64]bool, len(result))
	for skuIDStr := range result {
		skuID, _ := strconv.ParseUint(skuIDStr, 10, 64)
		if skuID > 0 {
			checkedStatus[skuID] = true
		}
	}
	return checkedStatus, nil
}

// GetCartItemCount 获取购物车中商品的种类数量。
func (r *cartRepo) GetCartItemCount(ctx context.Context, userID uint64) (uint32, error) {
	key := r.cartItemsKey(userID)
	count, err := r.rdb.HLen(ctx, key).Result()
	return uint32(count), err
}

// ClearCart 清空用户购物车（通常在下单后调用）。
func (r *cartRepo) ClearCart(ctx context.Context, userID uint64) error {
	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, r.cartItemsKey(userID))
	pipe.Del(ctx, r.cartCheckedKey(userID))
	_, err := pipe.Exec(ctx)
	return err
}
