package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

const (
	skuDetailPrefix = "product:sku:"
)

// skuReadRepository 基于 Redis 的 SKU 读模型仓储。
type skuReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSKUReadRepository 创建 SKU 读模型仓储。
func NewSKUReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SKUReadRepository {
	return &skuReadRepository{client: client, ttl: ttl}
}

func (r *skuReadRepository) Save(ctx context.Context, sku *domain.SKU) error {
	if sku == nil {
		return nil
	}
	data, err := json.Marshal(sku)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.skuKey(uint64(sku.ID)), data, r.ttl).Err()
}

func (r *skuReadRepository) GetByID(ctx context.Context, id uint64) (*domain.SKU, error) {
	data, err := r.client.Get(ctx, r.skuKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var sku domain.SKU
	if err := json.Unmarshal(data, &sku); err != nil {
		return nil, err
	}
	return &sku, nil
}

func (r *skuReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.skuKey(id)).Err()
}

func (r *skuReadRepository) skuKey(id uint64) string {
	return fmt.Sprintf("%s%d", skuDetailPrefix, id)
}
