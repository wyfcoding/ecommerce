// 生成摘要：定义订单事件存储接口，用于事件溯源与回放。
// 假设：事件存储按 user_id 分片路由，事务由调用方传入。
package domain

import (
	"context"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// OrderEventStore 定义订单聚合的事件存储接口。
// 该接口支持事务内写入，确保与订单写模型一致性。
type OrderEventStore interface {
	// SaveInTx 在事务中保存事件流。
	SaveInTx(ctx context.Context, tx any, userID uint64, aggregateID string, events []eventsourcing.DomainEvent, expectedVersion int64) error
	// LoadFromVersion 从指定版本加载事件。
	LoadFromVersion(ctx context.Context, userID uint64, aggregateID string, fromVersion int64) ([]eventsourcing.DomainEvent, error)
	// GetSnapshot 获取聚合快照。
	GetSnapshot(ctx context.Context, userID uint64, aggregateID string) (any, int64, error)
	// SaveSnapshotInTx 在事务中保存聚合快照。
	SaveSnapshotInTx(ctx context.Context, tx any, userID uint64, aggregateID string, state any, version int64) error
}
