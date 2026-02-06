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
	brandDetailPrefix = "product:brand:"
)

// brandReadRepository 基于 Redis 的品牌读模型仓储。
type brandReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewBrandReadRepository 创建品牌读模型仓储。
func NewBrandReadRepository(client redis.UniversalClient, ttl time.Duration) domain.BrandReadRepository {
	return &brandReadRepository{client: client, ttl: ttl}
}

func (r *brandReadRepository) Save(ctx context.Context, brand *domain.Brand) error {
	if brand == nil {
		return nil
	}
	data, err := json.Marshal(brand)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.brandKey(uint64(brand.ID)), data, r.ttl).Err()
}

func (r *brandReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	data, err := r.client.Get(ctx, r.brandKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var brand domain.Brand
	if err := json.Unmarshal(data, &brand); err != nil {
		return nil, err
	}
	return &brand, nil
}

func (r *brandReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.brandKey(id)).Err()
}

func (r *brandReadRepository) brandKey(id uint64) string {
	return fmt.Sprintf("%s%d", brandDetailPrefix, id)
}
