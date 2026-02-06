// 生成摘要：实现定价规则读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

const (
	pricingRuleDetailPrefix = "pricing:rule:detail:"
	pricingRuleActivePrefix = "pricing:rule:active:"
)

type pricingRuleReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewPricingRuleReadRepository 创建定价规则读模型仓储。
func NewPricingRuleReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PricingRuleReadRepository {
	return &pricingRuleReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *pricingRuleReadRepository) Save(ctx context.Context, rule *domain.PricingRule) error {
	if rule == nil {
		return nil
	}
	data, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	if err := r.client.Set(ctx, r.detailKey(rule.ID), data, r.ttl).Err(); err != nil {
		return err
	}
	activeKey := r.activeKey(rule.ProductID, rule.SkuID)
	if rule.IsActive() {
		return r.client.Set(ctx, activeKey, data, r.ttl).Err()
	}
	_ = r.client.Del(ctx, activeKey).Err()
	return nil
}

func (r *pricingRuleReadRepository) GetByID(ctx context.Context, id uint64) (*domain.PricingRule, error) {
	data, err := r.client.Get(ctx, r.detailKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rule domain.PricingRule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *pricingRuleReadRepository) GetActive(ctx context.Context, productID, skuID uint64) (*domain.PricingRule, error) {
	data, err := r.client.Get(ctx, r.activeKey(productID, skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rule domain.PricingRule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *pricingRuleReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.detailKey(id)).Err()
}

func (r *pricingRuleReadRepository) detailKey(id uint64) string {
	return fmt.Sprintf("%s%d", pricingRuleDetailPrefix, id)
}

func (r *pricingRuleReadRepository) activeKey(productID, skuID uint64) string {
	return fmt.Sprintf("%s%d:%d", pricingRuleActivePrefix, productID, skuID)
}
