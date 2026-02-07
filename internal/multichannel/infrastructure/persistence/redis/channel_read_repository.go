// 生成摘要：实现渠道读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

const channelListKey = "multichannel:channels:list"

type channelReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewChannelReadRepository 创建渠道读模型仓储。
func NewChannelReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ChannelReadRepository {
	return &channelReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *channelReadRepository) SaveAll(ctx context.Context, channels []*domain.Channel) error {
	if channels == nil {
		return nil
	}
	data, err := json.Marshal(channels)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, channelListKey, data, r.ttl).Err()
}

func (r *channelReadRepository) GetAll(ctx context.Context) ([]*domain.Channel, error) {
	data, err := r.client.Get(ctx, channelListKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []*domain.Channel
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *channelReadRepository) DeleteAll(ctx context.Context) error {
	return r.client.Del(ctx, channelListKey).Err()
}
