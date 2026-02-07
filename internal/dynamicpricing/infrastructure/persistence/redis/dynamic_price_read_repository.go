package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

const dynamicPriceLatestPrefix = "dynamicpricing:price:latest:"

type dynamicPriceReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewDynamicPriceReadRepository 创建动态价格读模型仓储。
func NewDynamicPriceReadRepository(client redis.UniversalClient, ttl time.Duration) domain.DynamicPriceReadRepository {
	return &dynamicPriceReadRepository{client: client, ttl: ttl}
}

func (r *dynamicPriceReadRepository) SaveLatest(ctx context.Context, price *domain.DynamicPrice) error {
	if price == nil {
		return nil
	}
	data, err := json.Marshal(price)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(price.SKUID), data, r.ttl).Err()
}

func (r *dynamicPriceReadRepository) GetLatest(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	data, err := r.client.Get(ctx, r.key(skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var price domain.DynamicPrice
	if err := json.Unmarshal(data, &price); err != nil {
		return nil, err
	}
	return &price, nil
}

func (r *dynamicPriceReadRepository) DeleteLatest(ctx context.Context, skuID uint64) error {
	return r.client.Del(ctx, r.key(skuID)).Err()
}

func (r *dynamicPriceReadRepository) key(skuID uint64) string {
	return fmt.Sprintf("%s%d", dynamicPriceLatestPrefix, skuID)
}
