package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
)

type addressReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAddressReadRepository 创建地址读模型仓储
func NewAddressReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AddressReadRepository {
	return &addressReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *addressReadRepository) Save(ctx context.Context, userID uint, addresses []*domain.Address) error {
	if len(addresses) == 0 {
		return r.Delete(ctx, userID)
	}
	data, err := json.Marshal(addresses)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.addrKey(userID), data, r.ttl).Err()
}

func (r *addressReadRepository) GetByUserID(ctx context.Context, userID uint) ([]*domain.Address, error) {
	data, err := r.client.Get(ctx, r.addrKey(userID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var addrs []*domain.Address
	if err := json.Unmarshal(data, &addrs); err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *addressReadRepository) Delete(ctx context.Context, userID uint) error {
	return r.client.Del(ctx, r.addrKey(userID)).Err()
}

func (r *addressReadRepository) addrKey(userID uint) string {
	return fmt.Sprintf("user:addresses:%d", userID)
}
