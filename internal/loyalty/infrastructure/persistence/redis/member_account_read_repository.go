// 生成摘要：实现会员账户读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

const accountPrefix = "loyalty:account:"

type memberAccountReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewMemberAccountReadRepository 创建会员账户读模型仓储。
func NewMemberAccountReadRepository(client redis.UniversalClient, ttl time.Duration) domain.MemberAccountReadRepository {
	return &memberAccountReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *memberAccountReadRepository) Save(ctx context.Context, account *domain.MemberAccount) error {
	if account == nil {
		return nil
	}
	data, err := json.Marshal(account)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(account.UserID), data, r.ttl).Err()
}

func (r *memberAccountReadRepository) GetByUserID(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	data, err := r.client.Get(ctx, r.key(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var account domain.MemberAccount
	if err := json.Unmarshal(data, &account); err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *memberAccountReadRepository) Delete(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.key(userID)).Err()
}

func (r *memberAccountReadRepository) key(userID uint64) string {
	return fmt.Sprintf("%s%d", accountPrefix, userID)
}
