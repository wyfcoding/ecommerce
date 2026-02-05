package domain

import (
	"context"
)

// EventPublisher 定义事件发布接口
type EventPublisher interface {
	// Publish 发布领域事件
	Publish(ctx context.Context, topic string, key string, event any) error

	// PublishInTx 在事务中发布事件
	PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error
}
