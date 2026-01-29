package messaging

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type outboxPublisher struct {
	mgr *outbox.Manager
}

// NewOutboxPublisher 创建基于 Outbox 的事件发布者
func NewOutboxPublisher(mgr *outbox.Manager) domain.EventPublisher {
	return &outboxPublisher{mgr: mgr}
}

func (p *outboxPublisher) Publish(ctx context.Context, topic string, event any) error {
	return p.mgr.PublishInTx(ctx, p.mgr.DB(), topic, "", event)
}

func (p *outboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return p.Publish(ctx, topic, event)
	}
	return p.mgr.PublishInTx(ctx, gormTx, topic, key, event)
}
