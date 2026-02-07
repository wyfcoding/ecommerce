// 生成摘要：实现拼团活动读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

const groupbuyDetailPrefix = "groupbuy:detail:"

type groupbuyReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewGroupbuyReadRepository 创建拼团活动读模型仓储。
func NewGroupbuyReadRepository(client redis.UniversalClient, ttl time.Duration) domain.GroupbuyReadRepository {
	return &groupbuyReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *groupbuyReadRepository) Save(ctx context.Context, groupbuy *domain.Groupbuy) error {
	if groupbuy == nil {
		return nil
	}
	data, err := json.Marshal(groupbuy)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(groupbuy.ID)), data, r.ttl).Err()
}

func (r *groupbuyReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Groupbuy, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var groupbuy domain.Groupbuy
	if err := json.Unmarshal(data, &groupbuy); err != nil {
		return nil, err
	}
	return &groupbuy, nil
}

func (r *groupbuyReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *groupbuyReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", groupbuyDetailPrefix, id)
}
