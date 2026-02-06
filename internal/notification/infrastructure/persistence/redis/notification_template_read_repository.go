// 生成摘要：实现通知模板读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

const templatePrefix = "notification:template:"

type notificationTemplateReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewNotificationTemplateReadRepository 创建通知模板读模型仓储。
func NewNotificationTemplateReadRepository(client redis.UniversalClient, ttl time.Duration) domain.NotificationTemplateReadRepository {
	return &notificationTemplateReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *notificationTemplateReadRepository) Save(ctx context.Context, template *domain.NotificationTemplate) error {
	if template == nil || template.Code == "" {
		return nil
	}
	data, err := json.Marshal(template)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(template.Code), data, r.ttl).Err()
}

func (r *notificationTemplateReadRepository) GetByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	if code == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.key(code)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var template domain.NotificationTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *notificationTemplateReadRepository) DeleteByCode(ctx context.Context, code string) error {
	if code == "" {
		return nil
	}
	return r.client.Del(ctx, r.key(code)).Err()
}

func (r *notificationTemplateReadRepository) key(code string) string {
	return fmt.Sprintf("%s%s", templatePrefix, code)
}
