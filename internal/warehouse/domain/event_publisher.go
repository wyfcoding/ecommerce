package domain

import "context"

// EventPublisher 定义了仓库领域事件发布的接口。
type EventPublisher interface {
	// Publish 发布一个普通事件。
	Publish(ctx context.Context, topic string, key string, event any) error

	// PublishInTx 在事务内发布事件 (Outbox 模式)。
	PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error
}
