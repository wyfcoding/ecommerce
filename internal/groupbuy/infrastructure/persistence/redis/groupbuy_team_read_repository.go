// 生成摘要：实现拼团团队读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

const groupbuyTeamDetailPrefix = "groupbuy:team:detail:"

type groupbuyTeamReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewGroupbuyTeamReadRepository 创建拼团团队读模型仓储。
func NewGroupbuyTeamReadRepository(client redis.UniversalClient, ttl time.Duration) domain.GroupbuyTeamReadRepository {
	return &groupbuyTeamReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *groupbuyTeamReadRepository) Save(ctx context.Context, team *domain.GroupbuyTeam) error {
	if team == nil {
		return nil
	}
	data, err := json.Marshal(team)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(team.ID)), data, r.ttl).Err()
}

func (r *groupbuyTeamReadRepository) GetByID(ctx context.Context, id uint64) (*domain.GroupbuyTeam, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var team domain.GroupbuyTeam
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *groupbuyTeamReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *groupbuyTeamReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", groupbuyTeamDetailPrefix, id)
}
