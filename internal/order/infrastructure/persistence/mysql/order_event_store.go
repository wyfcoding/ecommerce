// 生成摘要：实现订单事件存储（基于 GORM + 分片），支持事务内写入与快照。
// 假设：分片路由与订单写模型保持一致，事件表使用 order_events 与 order_event_snapshots。
// 变更说明：迁移至 mysql 子目录，明确存储技术边界。
package mysql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/database/sharding"
	"github.com/wyfcoding/pkg/eventsourcing"
	gormstore "github.com/wyfcoding/pkg/eventsourcing/persistence/gorm"

	"gorm.io/gorm"
)

const (
	orderEventTableName    = "order_events"
	orderSnapshotTableName = "order_event_snapshots"
)

// orderEventStore 基于分片的事件存储实现。
type orderEventStore struct {
	stores     []*gormstore.GormEventStore
	shardCount int
	logger     *slog.Logger
}

// NewOrderEventStore 创建订单事件存储。
func NewOrderEventStore(manager *sharding.Manager, logger *slog.Logger) (domain.OrderEventStore, error) {
	dbs := manager.GetAllDBs()
	if len(dbs) == 0 {
		return nil, errors.New("no shard database available for order event store")
	}

	stores := make([]*gormstore.GormEventStore, len(dbs))
	for i, db := range dbs {
		store, err := gormstore.NewGormEventStoreWithTables(db, orderEventTableName, orderSnapshotTableName)
		if err != nil {
			return nil, fmt.Errorf("init order event store shard %d failed: %w", i, err)
		}
		stores[i] = store
	}

	return &orderEventStore{
		stores:     stores,
		shardCount: len(dbs),
		logger:     logger,
	}, nil
}

// SaveInTx 在事务中保存事件流。
func (s *orderEventStore) SaveInTx(ctx context.Context, tx any, userID uint64, aggregateID string, events []eventsourcing.DomainEvent, expectedVersion int64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New("invalid transaction type for event store")
	}
	return s.storeByUserID(userID).SaveWithTx(ctx, gormTx, aggregateID, events, expectedVersion)
}

// LoadFromVersion 从指定版本加载事件。
func (s *orderEventStore) LoadFromVersion(ctx context.Context, userID uint64, aggregateID string, fromVersion int64) ([]eventsourcing.DomainEvent, error) {
	return s.storeByUserID(userID).LoadFromVersion(ctx, aggregateID, fromVersion)
}

// GetSnapshot 获取聚合快照。
func (s *orderEventStore) GetSnapshot(ctx context.Context, userID uint64, aggregateID string) (any, int64, error) {
	return s.storeByUserID(userID).GetSnapshot(ctx, aggregateID)
}

// SaveSnapshotInTx 在事务中保存聚合快照。
func (s *orderEventStore) SaveSnapshotInTx(ctx context.Context, tx any, userID uint64, aggregateID string, state any, version int64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return errors.New("invalid transaction type for snapshot store")
	}
	return s.storeByUserID(userID).SaveSnapshotWithTx(ctx, gormTx, aggregateID, state, version)
}

// storeByUserID 根据用户ID路由到对应分片的事件存储。
func (s *orderEventStore) storeByUserID(userID uint64) *gormstore.GormEventStore {
	if s.shardCount <= 1 {
		return s.stores[0]
	}
	index := int(userID % uint64(s.shardCount))
	return s.stores[index]
}
