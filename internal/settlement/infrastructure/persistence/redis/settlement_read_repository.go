// 生成摘要：实现结算读模型 Redis 仓储，提供按结算ID/结算单号的快速读取。
// 假设：结算单号与结算ID为全局唯一，缓存过期策略由调用方注入。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"

	"github.com/redis/go-redis/v9"
)

const (
	settlementDetailPrefix = "settlement:detail:"
	settlementNoPrefix     = "settlement:no:"
)

type settlementReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSettlementReadRepository 创建结算读模型仓储。
func NewSettlementReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SettlementReadRepository {
	return &settlementReadRepository{
		client: client,
		ttl:    ttl,
	}
}

// Save 保存或更新结算读模型。
func (r *settlementReadRepository) Save(ctx context.Context, settlement *domain.Settlement) error {
	if settlement == nil {
		return nil
	}

	data, err := json.Marshal(settlement)
	if err != nil {
		return err
	}

	idKey := r.settlementIDKey(uint64(settlement.ID))
	noKey := r.settlementNoKey(settlement.SettlementNo)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, idKey, data, r.ttl)
	pipe.Set(ctx, noKey, fmt.Sprintf("%d", settlement.ID), r.ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// GetByID 根据结算ID获取读模型。
func (r *settlementReadRepository) GetByID(ctx context.Context, settlementID uint64) (*domain.Settlement, error) {
	data, err := r.client.Get(ctx, r.settlementIDKey(settlementID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var settlement domain.Settlement
	if err := json.Unmarshal(data, &settlement); err != nil {
		return nil, err
	}
	return &settlement, nil
}

// GetByNo 根据结算单号获取读模型。
func (r *settlementReadRepository) GetByNo(ctx context.Context, settlementNo string) (*domain.Settlement, error) {
	idStr, err := r.client.Get(ctx, r.settlementNoKey(settlementNo)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var settlementID uint64
	if _, err := fmt.Sscanf(idStr, "%d", &settlementID); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, settlementID)
}

// Delete 删除读模型数据。
func (r *settlementReadRepository) Delete(ctx context.Context, settlementID uint64, settlementNo string) error {
	keys := []string{r.settlementIDKey(settlementID)}
	if settlementNo != "" {
		keys = append(keys, r.settlementNoKey(settlementNo))
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *settlementReadRepository) settlementIDKey(settlementID uint64) string {
	return fmt.Sprintf("%s%d", settlementDetailPrefix, settlementID)
}

func (r *settlementReadRepository) settlementNoKey(settlementNo string) string {
	return fmt.Sprintf("%s%s", settlementNoPrefix, settlementNo)
}
