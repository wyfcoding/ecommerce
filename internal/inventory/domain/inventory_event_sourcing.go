package domain

import (
	"fmt"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// RebuildInventoryFromEvents 从事件历史重建库存聚合状态。
func RebuildInventoryFromEvents(events []eventsourcing.DomainEvent) (*Inventory, error) {
	if len(events) == 0 {
		return nil, nil
	}

	inventory := &Inventory{}
	eventsourcing.LoadFromHistory(inventory, events)
	if inventory.SkuID != 0 {
		inventory.SetID(fmt.Sprintf("%d", inventory.SkuID))
	}
	inventory.PersistenceVer = inventory.Version()
	return inventory, nil
}
