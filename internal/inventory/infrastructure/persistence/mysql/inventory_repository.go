package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/database/sharding"
	"github.com/wyfcoding/pkg/eventsourcing"

	"gorm.io/gorm"
)

type inventoryRepository struct {
	sharding *sharding.Manager
}

// NewInventoryRepository 创建分片库存仓储。
func NewInventoryRepository(sharding *sharding.Manager) domain.InventoryRepository {
	return &inventoryRepository{sharding: sharding}
}

// Save 将库存实体保存到对应分片。
func (r *inventoryRepository) Save(ctx context.Context, inventory *domain.Inventory) error {
	if inventory == nil {
		return nil
	}
	db := r.sharding.GetDB(inventory.SkuID).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return r.saveInTx(ctx, tx, inventory, false)
	})
}

// SaveWithOptimisticLock 使用乐观锁保存。
func (r *inventoryRepository) SaveWithOptimisticLock(ctx context.Context, inventory *domain.Inventory) error {
	if inventory == nil {
		return nil
	}
	db := r.sharding.GetDB(inventory.SkuID).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return r.saveInTx(ctx, tx, inventory, true)
	})
}

func (r *inventoryRepository) saveInTx(ctx context.Context, tx *gorm.DB, inventory *domain.Inventory, optimistic bool) error {
	model := toInventoryModel(inventory)
	if model == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)

	if optimistic {
		currentVersion := inventory.PersistenceVer
		inventory.PersistenceVer++
		model.Version = inventory.PersistenceVer

		res := gormTx.Model(&InventoryModel{}).
			Where("id = ? AND version = ?", model.ID, currentVersion).
			Select("*").
			Updates(model)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("optimistic lock failed")
		}
	} else {
		if model.ID == 0 {
			if err := gormTx.Create(model).Error; err != nil {
				return err
			}
		} else {
			if err := gormTx.Save(model).Error; err != nil {
				return err
			}
		}
	}

	if inventory != nil {
		if synced := toDomainInventory(model); synced != nil {
			*inventory = *synced
		}
	}
	return nil
}

// SaveLog 保存库存日志到对应分片。
func (r *inventoryRepository) SaveLog(ctx context.Context, log *domain.InventoryLog) error {
	if log == nil {
		return nil
	}
	db := r.sharding.GetDB(log.SkuID)
	model := toInventoryLogModel(log)
	if model == nil {
		return nil
	}
	if err := db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	if synced := toDomainInventoryLog(model); synced != nil {
		*log = *synced
	}
	return nil
}

// GetBySkuID 定向查询分片。
func (r *inventoryRepository) GetBySkuID(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	db := r.sharding.GetDB(skuID)
	var inventory InventoryModel
	if err := db.WithContext(ctx).Where("sku_id = ?", skuID).First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainInventory(&inventory), nil
}

// GetBySkuIDs 执行跨分片的批量查询。
func (r *inventoryRepository) GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*domain.Inventory, error) {
	var allList []*domain.Inventory
	for _, id := range skuIDs {
		inv, err := r.GetBySkuID(ctx, id)
		if err != nil {
			return nil, err
		}
		if inv != nil {
			allList = append(allList, inv)
		}
	}
	return allList, nil
}

// List 扫描集群中所有的分片数据库。
func (r *inventoryRepository) List(ctx context.Context, offset, limit int) ([]*domain.Inventory, int64, error) {
	dbs := r.sharding.GetAllDBs()
	var allList []*domain.Inventory
	var totalCount int64

	for _, db := range dbs {
		var list []InventoryModel
		var count int64
		query := db.WithContext(ctx).Model(&InventoryModel{})
		if err := query.Count(&count).Error; err != nil {
			return nil, 0, err
		}
		totalCount += count

		if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
			return nil, 0, err
		}
		for i := range list {
			inv := toDomainInventory(&list[i])
			if inv != nil {
				allList = append(allList, inv)
			}
		}
	}

	if len(allList) > limit {
		allList = allList[:limit]
	}

	return allList, totalCount, nil
}

// Delete 从对应分片删除。
func (r *inventoryRepository) Delete(ctx context.Context, skuID uint64) error {
	db := r.sharding.GetDB(skuID)
	return db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&InventoryModel{}).Error
}

// GetLogs 获取指定分片下的日志。
func (r *inventoryRepository) GetLogs(ctx context.Context, skuID uint64, inventoryID uint64, offset, limit int) ([]*domain.InventoryLog, int64, error) {
	db := r.sharding.GetDB(skuID)
	var list []InventoryLogModel
	var total int64

	query := db.WithContext(ctx).Model(&InventoryLogModel{}).Where("inventory_id = ?", inventoryID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	logs := make([]*domain.InventoryLog, 0, len(list))
	for i := range list {
		log := toDomainInventoryLog(&list[i])
		if log != nil {
			logs = append(logs, log)
		}
	}

	return logs, total, nil
}

// ExecWithBarrier 实现 ExecWithBarrier 接口。
func (r *inventoryRepository) ExecWithBarrier(ctx context.Context, barrier any, fn func(ctx context.Context) error) error {
	// 暂时简单实现，如果是使用 DTM 等分布式事务管理器，这里应该是调用其 Barrier。
	return fn(ctx)
}

// eventStore 实现 domain.EventStore 接口。
type eventStore struct {
	sharding *sharding.Manager
}

// NewEventStore 创建 EventStore。
func NewEventStore(sharding *sharding.Manager) domain.EventStore {
	return &eventStore{sharding: sharding}
}

func (s *eventStore) Save(ctx context.Context, events []eventsourcing.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}
	// 使用聚合根 ID (SKU ID) 来确定分片
	idStr := events[0].AggregateID()
	var skuID uint64
	fmt.Sscanf(idStr, "%d", &skuID)

	db := s.sharding.GetDB(skuID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range events {
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}
			record := struct {
				gorm.Model
				AggregateID string `gorm:"column:aggregate_id;not null;index"`
				EventType   string `gorm:"column:event_type;not null"`
				Version     int64  `gorm:"column:version;not null"`
				Data        []byte `gorm:"column:data;type:json"`
			}{
				AggregateID: e.AggregateID(),
				EventType:   e.EventType(),
				Version:     e.Version(),
				Data:        data,
			}
			if err := tx.Table("inventory_events").Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *eventStore) GetHistory(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error) {
	var skuID uint64
	fmt.Sscanf(aggregateID, "%d", &skuID)
	db := s.sharding.GetDB(skuID)

	var records []struct {
		EventType string
		Version   int64
		Data      []byte
	}
	if err := db.WithContext(ctx).Table("inventory_events").Where("aggregate_id = ?", aggregateID).Order("version asc").Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]eventsourcing.DomainEvent, len(records))
	for i, r := range records {
		var event eventsourcing.DomainEvent
		switch r.EventType {
		case domain.StockLockedEventType:
			event = &domain.StockLockedEvent{}
		case domain.StockUnlockedEventType:
			event = &domain.StockUnlockedEvent{}
		case domain.StockDeductedEventType:
			event = &domain.StockDeductedEvent{}
		case domain.StockAddedEventType:
			event = &domain.StockAddedEvent{}
		default:
			return nil, fmt.Errorf("unknown event type: %s", r.EventType)
		}
		if err := json.Unmarshal(r.Data, event); err != nil {
			return nil, err
		}
		events[i] = event
	}
	return events, nil
}
