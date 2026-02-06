// 生成摘要：实现审计日志读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

const auditLogDetailPrefix = "audit:log:detail:"

type auditLogReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAuditLogReadRepository 创建审计日志读模型仓储。
func NewAuditLogReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AuditLogReadRepository {
	return &auditLogReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *auditLogReadRepository) Save(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return nil
	}
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(log.ID)), data, r.ttl).Err()
}

func (r *auditLogReadRepository) GetByID(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var log domain.AuditLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *auditLogReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *auditLogReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", auditLogDetailPrefix, id)
}
