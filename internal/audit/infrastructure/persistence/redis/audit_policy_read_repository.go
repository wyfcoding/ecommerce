// 生成摘要：实现审计策略读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

const auditPolicyDetailPrefix = "audit:policy:detail:"

type auditPolicyReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAuditPolicyReadRepository 创建审计策略读模型仓储。
func NewAuditPolicyReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AuditPolicyReadRepository {
	return &auditPolicyReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *auditPolicyReadRepository) Save(ctx context.Context, policy *domain.AuditPolicy) error {
	if policy == nil {
		return nil
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(policy.ID)), data, r.ttl).Err()
}

func (r *auditPolicyReadRepository) GetByID(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var policy domain.AuditPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *auditPolicyReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *auditPolicyReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", auditPolicyDetailPrefix, id)
}
