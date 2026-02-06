// 生成摘要：实现通知读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

const notificationPrefix = "notification:detail:"

type notificationReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewNotificationReadRepository 创建通知读模型仓储。
func NewNotificationReadRepository(client redis.UniversalClient, ttl time.Duration) domain.NotificationReadRepository {
	return &notificationReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *notificationReadRepository) Save(ctx context.Context, notification *domain.Notification) error {
	if notification == nil {
		return nil
	}
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(notification.ID), data, r.ttl).Err()
}

func (r *notificationReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Notification, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var notification domain.Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *notificationReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", notificationPrefix, id)
}
