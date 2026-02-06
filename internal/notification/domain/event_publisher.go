package domain

import "context"

// EventPublisher 领域事件发布接口，用于解耦应用层与可靠消息实现（如 Outbox）。
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, event any) error
	PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error
}
