// 生成摘要：实现优惠券读模型 Redis 仓储，提供按优惠券ID/编号的快速读取。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

const (
	couponIDPrefix = "coupon:detail:id:"
	couponNoPrefix = "coupon:detail:no:"
)

type couponReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewCouponReadRepository 创建优惠券读模型仓储。
func NewCouponReadRepository(client redis.UniversalClient, ttl time.Duration) domain.CouponReadRepository {
	return &couponReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *couponReadRepository) SaveCoupon(ctx context.Context, coupon *domain.Coupon) error {
	if coupon == nil {
		return nil
	}
	data, err := json.Marshal(coupon)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.keyByID(coupon.ID), data, r.ttl)
	if coupon.CouponNo != "" {
		pipe.Set(ctx, r.keyByNo(coupon.CouponNo), data, r.ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *couponReadRepository) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	data, err := r.client.Get(ctx, r.keyByID(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var coupon domain.Coupon
	if err := json.Unmarshal(data, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponReadRepository) GetCouponByNo(ctx context.Context, couponNo string) (*domain.Coupon, error) {
	if couponNo == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.keyByNo(couponNo)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var coupon domain.Coupon
	if err := json.Unmarshal(data, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponReadRepository) DeleteCoupon(ctx context.Context, id uint64, couponNo string) error {
	keys := make([]string, 0, 2)
	if id != 0 {
		keys = append(keys, r.keyByID(id))
	}
	if couponNo != "" {
		keys = append(keys, r.keyByNo(couponNo))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *couponReadRepository) keyByID(id uint64) string {
	return fmt.Sprintf("%s%d", couponIDPrefix, id)
}

func (r *couponReadRepository) keyByNo(couponNo string) string {
	return couponNoPrefix + couponNo
}
