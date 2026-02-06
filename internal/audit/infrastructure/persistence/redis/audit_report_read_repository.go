// 生成摘要：实现审计报告读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

const auditReportDetailPrefix = "audit:report:detail:"

type auditReportReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAuditReportReadRepository 创建审计报告读模型仓储。
func NewAuditReportReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AuditReportReadRepository {
	return &auditReportReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *auditReportReadRepository) Save(ctx context.Context, report *domain.AuditReport) error {
	if report == nil {
		return nil
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(report.ID)), data, r.ttl).Err()
}

func (r *auditReportReadRepository) GetByID(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var report domain.AuditReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *auditReportReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *auditReportReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", auditReportDetailPrefix, id)
}
