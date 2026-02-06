// 生成摘要：实现售后配置读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

const afterSalesConfigPrefix = "aftersales:config:"

type configReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewConfigReadRepository 创建售后配置读模型仓储。
func NewConfigReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AfterSalesConfigReadRepository {
	return &configReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *configReadRepository) Save(ctx context.Context, cfg *domain.AfterSalesConfig) error {
	if cfg == nil || cfg.Key == "" {
		return nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(cfg.Key), data, r.ttl).Err()
}

func (r *configReadRepository) GetByKey(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	if key == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg domain.AfterSalesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *configReadRepository) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	return r.client.Del(ctx, r.key(key)).Err()
}

func (r *configReadRepository) key(key string) string {
	return fmt.Sprintf("%s%s", afterSalesConfigPrefix, key)
}
