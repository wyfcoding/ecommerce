// 生成摘要：实现黑名单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"
)

const (
	blacklistTypePrefix = "risk:blacklist:type:"
	blacklistIDPrefix   = "risk:blacklist:id:"
)

type blacklistReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewBlacklistReadRepository 创建黑名单读模型仓储。
func NewBlacklistReadRepository(client redis.UniversalClient, ttl time.Duration) domain.BlacklistReadRepository {
	return &blacklistReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *blacklistReadRepository) Save(ctx context.Context, entry *domain.Blacklist) error {
	if entry == nil {
		return nil
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if err := r.client.Set(ctx, r.keyByTypeValue(entry.Type, entry.Value), data, r.ttl).Err(); err != nil {
		return err
	}
	if entry.ID != 0 {
		return r.client.Set(ctx, r.keyByID(uint64(entry.ID)), data, r.ttl).Err()
	}
	return nil
}

func (r *blacklistReadRepository) GetByTypeValue(ctx context.Context, bType domain.BlacklistType, value string) (*domain.Blacklist, error) {
	data, err := r.client.Get(ctx, r.keyByTypeValue(bType, value)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entry domain.Blacklist
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *blacklistReadRepository) DeleteByTypeValue(ctx context.Context, bType domain.BlacklistType, value string) error {
	return r.client.Del(ctx, r.keyByTypeValue(bType, value)).Err()
}

func (r *blacklistReadRepository) DeleteByID(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.keyByID(id)).Err()
}

func (r *blacklistReadRepository) keyByTypeValue(bType domain.BlacklistType, value string) string {
	return fmt.Sprintf("%s%s:%s", blacklistTypePrefix, bType, value)
}

func (r *blacklistReadRepository) keyByID(id uint64) string {
	return fmt.Sprintf("%s%d", blacklistIDPrefix, id)
}
