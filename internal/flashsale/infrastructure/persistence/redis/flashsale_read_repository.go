package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

const flashsaleDetailPrefix = "flashsale:detail:"

type flashsaleReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewFlashsaleReadRepository 创建秒杀活动读模型仓储。
func NewFlashsaleReadRepository(client redis.UniversalClient, ttl time.Duration) domain.FlashsaleReadRepository {
	return &flashsaleReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *flashsaleReadRepository) Save(ctx context.Context, flashsale *domain.Flashsale) error {
	if flashsale == nil {
		return nil
	}
	data, err := json.Marshal(flashsale)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(uint64(flashsale.ID)), data, r.ttl).Err()
}

func (r *flashsaleReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var flashsale domain.Flashsale
	if err := json.Unmarshal(data, &flashsale); err != nil {
		return nil, err
	}
	return &flashsale, nil
}

func (r *flashsaleReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *flashsaleReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", flashsaleDetailPrefix, id)
}
