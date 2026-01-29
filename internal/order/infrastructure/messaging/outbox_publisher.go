package messaging

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type outboxPublisher struct {
	manager *outbox.Manager
}

// NewOutboxPublisher 创建基于 Outbox 模式的事件发布者。
func NewOutboxPublisher(manager *outbox.Manager) domain.EventPublisher {
	return &outboxPublisher{manager: manager}
}

func (p *outboxPublisher) Publish(ctx context.Context, topic string, key string, event any) error {
	// 普通发布逻辑，如果没有显式事务，则使用底层管理器默认配置（通常这较少在订单核心流程中使用）
	// 因为 Outbox 核心价值在于事务一致性。
	// 这里假设不提供事务时，无法通过 Outbox 保证强一致。
	return fmt.Errorf("Publish without transaction is not supported in Outbox mode for order system, use PublishInTx")
}

func (p *outboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return fmt.Errorf("invalid transaction type: expected *gorm.DB")
	}
	return p.manager.PublishInTx(ctx, gormTx, topic, key, event)
}
