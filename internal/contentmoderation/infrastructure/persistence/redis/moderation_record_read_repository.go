// 生成摘要：实现内容审核记录读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

const moderationRecordDetailPrefix = "contentmoderation:record:detail:"

type moderationRecordReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewModerationRecordReadRepository 创建审核记录读模型仓储。
func NewModerationRecordReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ModerationRecordReadRepository {
	return &moderationRecordReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *moderationRecordReadRepository) Save(ctx context.Context, record *domain.ModerationRecord) error {
	if record == nil {
		return nil
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(record.ID)), data, r.ttl).Err()
}

func (r *moderationRecordReadRepository) GetByID(ctx context.Context, id uint64) (*domain.ModerationRecord, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record domain.ModerationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *moderationRecordReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *moderationRecordReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", moderationRecordDetailPrefix, id)
}
