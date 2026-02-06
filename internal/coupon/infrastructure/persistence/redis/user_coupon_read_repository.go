// 生成摘要：实现用户优惠券读模型 Redis 仓储，提供按用户优惠券ID的快速读取。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

const userCouponIDPrefix = "coupon:user:detail:id:"

type userCouponReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewUserCouponReadRepository 创建用户优惠券读模型仓储。
func NewUserCouponReadRepository(client redis.UniversalClient, ttl time.Duration) domain.UserCouponReadRepository {
	return &userCouponReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *userCouponReadRepository) SaveUserCoupon(ctx context.Context, coupon *domain.UserCoupon) error {
	if coupon == nil {
		return nil
	}
	data, err := json.Marshal(coupon)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.keyByID(coupon.ID), data, r.ttl).Err()
}

func (r *userCouponReadRepository) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	data, err := r.client.Get(ctx, r.keyByID(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var coupon domain.UserCoupon
	if err := json.Unmarshal(data, &coupon); err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *userCouponReadRepository) DeleteUserCoupon(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.keyByID(id)).Err()
}

func (r *userCouponReadRepository) keyByID(id uint64) string {
	return fmt.Sprintf("%s%d", userCouponIDPrefix, id)
}
