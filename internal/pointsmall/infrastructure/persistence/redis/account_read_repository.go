// 生成摘要：实现积分账户读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

const pointsAccountPrefix = "pointsmall:account:user:"

type accountReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAccountReadRepository 创建积分账户读模型仓储。
func NewAccountReadRepository(client redis.UniversalClient, ttl time.Duration) domain.PointsAccountReadRepository {
	return &accountReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *accountReadRepository) Save(ctx context.Context, account *domain.PointsAccount) error {
	if account == nil {
		return nil
	}
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(account.UserID), data, r.ttl).Err()
}

func (r *accountReadRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var account domain.PointsAccount
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountReadRepository) DeleteByUserID(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *accountReadRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", pointsAccountPrefix, userID)
}
