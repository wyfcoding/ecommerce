package domain

import (
	"context"
)

// EventPublisher 定义了支付领域发布事件的接口。
// 它可以支持直接发布或带事务的 Outbox 发布。
type EventPublisher interface {
	// Publish 发布一个领域事件。
	Publish(ctx context.Context, topic string, key string, event any) error

	// PublishInTx 在给定的数据库事务中发布一个领域事件（通常用于 Outbox 模式）。
	// tx 类型通常为 *gorm.DB。
	PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error
}
