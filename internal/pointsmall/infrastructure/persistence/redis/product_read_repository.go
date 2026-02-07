// 生成摘要：实现积分商品读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

const pointsProductDetailPrefix = "pointsmall:product:detail:"

type productReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewProductReadRepository 创建积分商品读模型仓储。
func NewProductReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PointsProductReadRepository {
	return &productReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *productReadRepository) Save(ctx context.Context, product *domain.PointsProduct) error {
	if product == nil {
		return nil
	}
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(product.ID)), data, r.ttl).Err()
}

func (r *productReadRepository) GetByID(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var product domain.PointsProduct
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *productReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", pointsProductDetailPrefix, id)
}
