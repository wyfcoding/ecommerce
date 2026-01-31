package domain

import "context"

// EventPublisher 定义事件发布接口
type EventPublisher interface {
	// Publish 发布领域事件
	Publish(ctx context.Context, topic string, event any) error
}
