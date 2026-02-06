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
	productDetailPrefix = "product:detail:"
)

// productReadRepository 基于 Redis 的商品读模型仓储。
type productReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewProductReadRepository 创建商品读模型仓储。
func NewProductReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ProductReadRepository {
	return &productReadRepository{client: client, ttl: ttl}
}

func (r *productReadRepository) Save(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return nil
	}
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.productKey(uint64(product.ID)), data, r.ttl).Err()
}

func (r *productReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Product, error) {
	data, err := r.client.Get(ctx, r.productKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var product domain.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.productKey(id)).Err()
}

func (r *productReadRepository) productKey(id uint64) string {
	return fmt.Sprintf("%s%d", productDetailPrefix, id)
}
