// 生成摘要：实现广告位读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

const bannerPrefix = "marketing:banner:detail:"

type bannerReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewBannerReadRepository 创建广告位读模型仓储。
func NewBannerReadRepository(client redis.UniversalClient, ttl time.Duration) domain.BannerReadRepository {
	return &bannerReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *bannerReadRepository) Save(ctx context.Context, banner *domain.Banner) error {
	if banner == nil {
		return nil
	}
	data, err := json.Marshal(banner)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(banner.ID), data, r.ttl).Err()
}

func (r *bannerReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Banner, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var banner domain.Banner
	if err := json.Unmarshal(data, &banner); err != nil {
		return nil, err
	}
	return &banner, nil
}

func (r *bannerReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *bannerReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", bannerPrefix, id)
}
