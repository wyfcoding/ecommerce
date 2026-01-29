package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

// OutboxPublisher 基于 Outbox 模式的事件发布者实现。
// 确保事件在同一事务中被持久化，保证最终一致性。
type OutboxPublisher struct {
	manager *outbox.Manager
}

// NewOutboxPublisher 创建一个新的 OutboxPublisher 实例。
func NewOutboxPublisher(manager *outbox.Manager) domain.EventPublisher {
	return &OutboxPublisher{manager: manager}
}

// Publish 发布一个普通事件（非事务内）。
func (p *OutboxPublisher) Publish(ctx context.Context, topic string, key string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.manager.Save(ctx, nil, topic, key, payload)
}

// PublishInTx 在事务中发布事件，核心用于 Outbox 模式。
func (p *OutboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return fmt.Errorf("tx must be *gorm.DB, got %T", tx)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return p.manager.Save(ctx, gormTx, topic, key, payload)
}
