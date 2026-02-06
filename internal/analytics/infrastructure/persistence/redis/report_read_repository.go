// 生成摘要：实现报告读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

const reportDetailPrefix = "analytics:report:detail:"

type reportReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewReportReadRepository 创建报告读模型仓储。
func NewReportReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ReportReadRepository {
	return &reportReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *reportReadRepository) Save(ctx context.Context, report *domain.Report) error {
	if report == nil {
		return nil
	}
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(report.ID), data, r.ttl).Err()
}

func (r *reportReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Report, error) {
	data, err := r.client.Get(ctx, r.key(uint(id))).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var report domain.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *reportReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(uint(id))).Err()
}

func (r *reportReadRepository) key(id uint) string {
	return fmt.Sprintf("%s%d", reportDetailPrefix, id)
}
