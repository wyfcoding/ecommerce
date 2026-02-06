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
	categoryDetailPrefix = "product:category:"
)

// categoryReadRepository 基于 Redis 的分类读模型仓储。
type categoryReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewCategoryReadRepository 创建分类读模型仓储。
func NewCategoryReadRepository(client redis.UniversalClient, ttl time.Duration) domain.CategoryReadRepository {
	return &categoryReadRepository{client: client, ttl: ttl}
}

func (r *categoryReadRepository) Save(ctx context.Context, category *domain.Category) error {
	if category == nil {
		return nil
	}
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.categoryKey(uint64(category.ID)), data, r.ttl).Err()
}

func (r *categoryReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Category, error) {
	data, err := r.client.Get(ctx, r.categoryKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var category domain.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.categoryKey(id)).Err()
}

func (r *categoryReadRepository) categoryKey(id uint64) string {
	return fmt.Sprintf("%s%d", categoryDetailPrefix, id)
}
