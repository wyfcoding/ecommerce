package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

const pricingStrategyPrefix = "dynamicpricing:strategy:"

type pricingStrategyReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewPricingStrategyReadRepository 创建定价策略读模型仓储。
func NewPricingStrategyReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PricingStrategyReadRepository {
	return &pricingStrategyReadRepository{client: client, ttl: ttl}
}

func (r *pricingStrategyReadRepository) Save(ctx context.Context, strategy *domain.PricingStrategy) error {
	if strategy == nil {
		return nil
	}
	data, err := json.Marshal(strategy)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(strategy.SKUID), data, r.ttl).Err()
}

func (r *pricingStrategyReadRepository) GetBySKU(ctx context.Context, skuID uint64) (*domain.PricingStrategy, error) {
	data, err := r.client.Get(ctx, r.key(skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var strategy domain.PricingStrategy
	if err := json.Unmarshal(data, &strategy); err != nil {
		return nil, err
	}
	return &strategy, nil
}

func (r *pricingStrategyReadRepository) Delete(ctx context.Context, skuID uint64) error {
	return r.client.Del(ctx, r.key(skuID)).Err()
}

func (r *pricingStrategyReadRepository) key(skuID uint64) string {
	return fmt.Sprintf("%s%d", pricingStrategyPrefix, skuID)
}
