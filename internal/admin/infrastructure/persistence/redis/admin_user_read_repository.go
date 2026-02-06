// 生成摘要：实现管理员读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

const adminUserPrefix = "admin:user:detail:"

type adminUserReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewAdminUserReadRepository 创建管理员读模型仓储。
func NewAdminUserReadRepository(client redis.UniversalClient, ttl time.Duration) domain.AdminUserReadRepository {
	return &adminUserReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *adminUserReadRepository) Save(ctx context.Context, user *domain.AdminUser) error {
	if user == nil {
		return nil
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(user.ID), data, r.ttl).Err()
}

func (r *adminUserReadRepository) GetByID(ctx context.Context, id uint) (*domain.AdminUser, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var user domain.AdminUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *adminUserReadRepository) Delete(ctx context.Context, id uint) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *adminUserReadRepository) key(id uint) string {
	return fmt.Sprintf("%s%d", adminUserPrefix, id)
}
