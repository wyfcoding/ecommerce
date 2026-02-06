package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

const (
	logisticsDetailPrefix   = "logistics:detail:"
	logisticsTrackingPrefix = "logistics:tracking:"
	logisticsOrderPrefix    = "logistics:order:"
)

// logisticsReadRepository 基于 Redis 的物流读模型仓储。
type logisticsReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewLogisticsReadRepository 创建物流读模型仓储。
func NewLogisticsReadRepository(client redis.UniversalClient, ttl time.Duration) domain.LogisticsReadRepository {
	return &logisticsReadRepository{client: client, ttl: ttl}
}

func (r *logisticsReadRepository) Save(ctx context.Context, logistics *domain.Logistics) error {
	if logistics == nil {
		return nil
	}
	data, err := json.Marshal(logistics)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.idKey(uint64(logistics.ID)), data, r.ttl)
	if logistics.TrackingNo != "" {
		pipe.Set(ctx, r.trackingKey(logistics.TrackingNo), data, r.ttl)
	}
	if logistics.OrderID != 0 {
		pipe.Set(ctx, r.orderKey(logistics.OrderID), data, r.ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *logisticsReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return r.get(ctx, r.idKey(id))
}

func (r *logisticsReadRepository) GetByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return r.get(ctx, r.trackingKey(trackingNo))
}

func (r *logisticsReadRepository) GetByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	return r.get(ctx, r.orderKey(orderID))
}

func (r *logisticsReadRepository) Delete(ctx context.Context, id uint64, trackingNo string, orderID uint64) error {
	keys := make([]string, 0, 3)
	if id != 0 {
		keys = append(keys, r.idKey(id))
	}
	if trackingNo != "" {
		keys = append(keys, r.trackingKey(trackingNo))
	}
	if orderID != 0 {
		keys = append(keys, r.orderKey(orderID))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *logisticsReadRepository) get(ctx context.Context, key string) (*domain.Logistics, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var logistics domain.Logistics
	if err := json.Unmarshal(data, &logistics); err != nil {
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsReadRepository) idKey(id uint64) string {
	return fmt.Sprintf("%s%d", logisticsDetailPrefix, id)
}

func (r *logisticsReadRepository) trackingKey(trackingNo string) string {
	return fmt.Sprintf("%s%s", logisticsTrackingPrefix, trackingNo)
}

func (r *logisticsReadRepository) orderKey(orderID uint64) string {
	return fmt.Sprintf("%s%d", logisticsOrderPrefix, orderID)
}
