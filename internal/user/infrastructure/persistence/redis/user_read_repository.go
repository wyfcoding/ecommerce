package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/user/domain"
)

type userReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewUserReadRepository 创建用户读模型仓储
func NewUserReadRepository(client redis.UniversalClient, ttl time.Duration) domain.UserReadRepository {
	return &userReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *userReadRepository) Save(ctx context.Context, user *domain.User) error {
	if user == nil {
		return nil
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.userKey(user.ID), data, r.ttl).Err()
}

func (r *userReadRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	data, err := r.client.Get(ctx, r.userKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userReadRepository) Delete(ctx context.Context, id uint) error {
	return r.client.Del(ctx, r.userKey(id)).Err()
}

func (r *userReadRepository) userKey(id uint) string {
	return fmt.Sprintf("user:profile:%d", id)
}
