// 生成摘要：实现会员权益读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

const benefitPrefix = "loyalty:benefit:level:"

type memberBenefitReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewMemberBenefitReadRepository 创建会员权益读模型仓储。
func NewMemberBenefitReadRepository(client redis.UniversalClient, ttl time.Duration) domain.MemberBenefitReadRepository {
	return &memberBenefitReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *memberBenefitReadRepository) Save(ctx context.Context, benefit *domain.MemberBenefit) error {
	if benefit == nil {
		return nil
	}
	data, err := json.Marshal(benefit)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(benefit.Level), data, r.ttl).Err()
}

func (r *memberBenefitReadRepository) GetByLevel(ctx context.Context, level domain.MemberLevel) (*domain.MemberBenefit, error) {
	if level == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.key(level)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var benefit domain.MemberBenefit
	if err := json.Unmarshal(data, &benefit); err != nil {
		return nil, err
	}
	return &benefit, nil
}

func (r *memberBenefitReadRepository) DeleteByLevel(ctx context.Context, level domain.MemberLevel) error {
	if level == "" {
		return nil
	}
	return r.client.Del(ctx, r.key(level)).Err()
}

func (r *memberBenefitReadRepository) key(level domain.MemberLevel) string {
	return fmt.Sprintf("%s%s", benefitPrefix, level)
}
