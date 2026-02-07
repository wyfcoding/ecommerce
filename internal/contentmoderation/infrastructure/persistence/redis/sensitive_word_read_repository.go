// 生成摘要：实现敏感词读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

const sensitiveWordDetailPrefix = "contentmoderation:word:detail:"

type sensitiveWordReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSensitiveWordReadRepository 创建敏感词读模型仓储。
func NewSensitiveWordReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SensitiveWordReadRepository {
	return &sensitiveWordReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *sensitiveWordReadRepository) Save(ctx context.Context, word *domain.SensitiveWord) error {
	if word == nil {
		return nil
	}
	data, err := json.Marshal(word)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(word.ID)), data, r.ttl).Err()
}

func (r *sensitiveWordReadRepository) GetByID(ctx context.Context, id uint64) (*domain.SensitiveWord, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var word domain.SensitiveWord
	if err := json.Unmarshal(data, &word); err != nil {
		return nil, err
	}
	return &word, nil
}

func (r *sensitiveWordReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *sensitiveWordReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", sensitiveWordDetailPrefix, id)
}
