package domain

import (
	"context"
)

// EventPublisher 领域事件发布接口，用于解耦应用层与具体的可靠消息实现（如 Outbox）
type EventPublisher interface {
	// Publish 发布事件
	Publish(ctx context.Context, topic string, event any) error

	// PublishInTx 在事务中发布事件
	// tx 通常是 *gorm.DB
	PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error
}
