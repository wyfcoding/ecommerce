package messaging

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
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
	// 非事务发布，注意这里的逻辑应根据 SkuID 获取对应的分片 DB，但 Outbox Manager 初始化时通常只绑定一个主库。
	// 在分片场景下，我们通常希望事件落入当前分片 DB 的 Outbox 表中。
	// 这里假设 p.mgr.DB() 在 Publish 时会自动被正确的上下文或分片逻辑注入。
	// 但当前的 pkg/messagequeue/outbox 通常绑定一个固定 DB。
	// 为了兼容分片，我们在 PublishInTx 中显式传递事务 DB。
	return p.mgr.PublishInTx(ctx, p.mgr.DB(), topic, "", event)
}

func (p *outboxPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return p.Publish(ctx, topic, event)
	}
	return p.mgr.PublishInTx(ctx, gormTx, topic, key, event)
}
