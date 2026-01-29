package messaging

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type outboxPublisher struct {
	mgr *outbox.Manager
}

// NewOutboxPublisher 创建一个基于 Outbox 的事件发布者。
func NewOutboxPublisher(mgr *outbox.Manager) domain.EventPublisher {
	return &outboxPublisher{mgr: mgr}
}

func (p *outboxPublisher) Publish(ctx context.Context, topic string, key string, event any) error {
	// 注意：订单系统通常必须在事务中发布以保证一致性
	// 此处 Publish 作为兜底使用默认连接（非事务环境，风险较高，建议使用 PublishInTx）
	return p.mgr.PublishInTx(ctx, p.mgr.DB(), topic, key, event)
}

func (p *outboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return p.Publish(ctx, topic, key, event)
	}
	return p.mgr.PublishInTx(ctx, gormTx, topic, key, event)
}
