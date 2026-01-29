package messaging

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

// outboxPublisher 实现 domain.EventPublisher。
type outboxPublisher struct {
	manager *outbox.Manager
}

// NewOutboxPublisher 创建 Outbox 事件发布器。
func NewOutboxPublisher(manager *outbox.Manager) domain.EventPublisher {
	return &outboxPublisher{manager: manager}
}

func (p *outboxPublisher) Publish(ctx context.Context, topic string, key string, event any) error {
	return p.manager.PublishInTx(ctx, p.manager.DB(), topic, key, event)
}

func (p *outboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	// tx 必须是 *gorm.DB
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return fmt.Errorf("transaction must be *gorm.DB")
	}
	return p.manager.PublishInTx(ctx, gormTx, topic, key, event)
}
