// 生成摘要：实现渠道订单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

const localOrderDetailPrefix = "multichannel:order:detail:"

type localOrderReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewLocalOrderReadRepository 创建渠道订单读模型仓储。
func NewLocalOrderReadRepository(client redis.UniversalClient, ttl time.Duration) domain.LocalOrderReadRepository {
	return &localOrderReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *localOrderReadRepository) Save(ctx context.Context, order *domain.LocalOrder) error {
	if order == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(order.ChannelID, order.ChannelOrderID), data, r.ttl).Err()
}

func (r *localOrderReadRepository) GetByChannelOrderID(ctx context.Context, channelID uint64, channelOrderID string) (*domain.LocalOrder, error) {
	data, err := r.client.Get(ctx, r.key(channelID, channelOrderID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var order domain.LocalOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *localOrderReadRepository) Delete(ctx context.Context, channelID uint64, channelOrderID string) error {
	return r.client.Del(ctx, r.key(channelID, channelOrderID)).Err()
}

func (r *localOrderReadRepository) key(channelID uint64, channelOrderID string) string {
	return fmt.Sprintf("%s%d:%s", localOrderDetailPrefix, channelID, channelOrderID)
}
