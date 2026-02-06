// 生成摘要：实现售后读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

const (
	afterSalesIDPrefix = "aftersales:detail:id:"
	afterSalesNoPrefix = "aftersales:detail:no:"
)

type afterSalesReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAfterSalesReadRepository 创建售后读模型仓储。
func NewAfterSalesReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AfterSalesReadRepository {
	return &afterSalesReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *afterSalesReadRepository) Save(ctx context.Context, afterSales *domain.AfterSales) error {
	if afterSales == nil {
		return nil
	}
	data, err := json.Marshal(afterSales)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.keyByID(afterSales.ID), data, r.ttl)
	if afterSales.AfterSalesNo != "" {
		pipe.Set(ctx, r.keyByNo(afterSales.AfterSalesNo), data, r.ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *afterSalesReadRepository) GetByID(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	data, err := r.client.Get(ctx, r.keyByID(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var item domain.AfterSales
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *afterSalesReadRepository) GetByNo(ctx context.Context, no string) (*domain.AfterSales, error) {
	if no == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.keyByNo(no)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var item domain.AfterSales
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *afterSalesReadRepository) Delete(ctx context.Context, id uint64, no string) error {
	keys := make([]string, 0, 2)
	if id != 0 {
		keys = append(keys, r.keyByID(id))
	}
	if no != "" {
		keys = append(keys, r.keyByNo(no))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *afterSalesReadRepository) keyByID(id uint64) string {
	return fmt.Sprintf("%s%d", afterSalesIDPrefix, id)
}

func (r *afterSalesReadRepository) keyByNo(no string) string {
	return afterSalesNoPrefix + no
}
