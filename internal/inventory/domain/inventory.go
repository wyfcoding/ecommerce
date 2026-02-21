package domain

import (
	"fmt"
	"time"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// StockType 库存类型
type StockType int

const (
	Sellable  StockType = 1 // 可售库存
	Reserved  StockType = 2 // 预占库存 (下单未支付)
	Frozen    StockType = 3 // 冻结库存 (售后/风控)
	Defective StockType = 4 // 残次品库存
)

// InventoryItem 核心库存实体
type InventoryItem struct {
	SKUID       string    `json:"sku_id"`
	ProductID   string    `json:"product_id"`
	Category    string    `json:"category"`
	WarehouseID string    `json:"warehouse_id"`
	Quantity    int64     `json:"quantity"`  // 当前数量
	UnitCost    float64   `json:"unit_cost"` // 单位成本
	Total       int64     `json:"total"`     // 物理总库存
	Available   int64     `json:"available"` // 可售 = Total - Reserved - Frozen
	Reserved    int64     `json:"reserved"`
	Frozen      int64     `json:"frozen"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DeductRequest 扣减请求
type DeductRequest struct {
	TxID        string // 幂等键
	SKUID       string
	WarehouseID string
	Quantity    int64
}

// InventoryRepository 接口定义

// InventoryStatus 库存状态
type InventoryStatus int

const (
	InventoryStatusActive InventoryStatus = 1
	InventoryStatusFrozen InventoryStatus = 2
)

// HedgingConfig 对冲策略配置
type HedgingConfig struct {
	Enabled          bool    `json:"enabled"`
	HedgeRatio       float64 `json:"hedge_ratio"`       // 对冲比例 (0.0 - 1.0)
	InstrumentSymbol string  `json:"instrument_symbol"` // 对应金融产品的代码 (如 BTC-USDT-SWAP)
	MinHedgeQty      int32   `json:"min_hedge_qty"`     // 最小对冲触发数量
	AutoHedge        bool    `json:"auto_hedge"`        // 是否自动执行对冲
}

// Inventory 库存聚合根 (Write Model)
// 注意：InventoryItem 是 Read Model / DTO，Inventory 是 Write Model Aggregate.
// 之前版本中可能丢失了此定义，现进行恢复。
type Inventory struct {
	eventsourcing.AggregateRoot
	DbID             uint64          `json:"id"`
	SkuID            uint64          `json:"sku_id"`
	ProductID        uint64          `json:"product_id"`
	WarehouseID      uint64          `json:"warehouse_id"`
	AvailableStock   int32           `json:"available_stock"`
	LockedStock      int32           `json:"locked_stock"`
	TotalStock       int32           `json:"total_stock"`
	Status           InventoryStatus `json:"status"`
	WarningThreshold int32           `json:"warning_threshold"`
	PersistenceVer   int64           `json:"persistence_ver"`
	HedgingConfig    *HedgingConfig  `json:"hedging_config"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (i *Inventory) AggregateID() string {
	return fmt.Sprintf("%d", i.SkuID)
}

func (i *Inventory) GetID() uint64 {
	return i.DbID
}

func NewInventory(skuID, productID, warehouseID uint64, initialStock int32) *Inventory {
	return &Inventory{
		SkuID:          skuID,
		ProductID:      productID,
		WarehouseID:    warehouseID,
		TotalStock:     initialStock,
		AvailableStock: initialStock,
		Status:         InventoryStatusActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Apply 应用事件更新状态
func (i *Inventory) Apply(event eventsourcing.DomainEvent) {
	switch e := event.(type) {
	case *StockLockedEvent:
		i.LockedStock += e.Quantity
		i.AvailableStock -= e.Quantity
	case *StockUnlockedEvent:
		i.LockedStock -= e.Quantity
		i.AvailableStock += e.Quantity
	case *StockDeductedEvent:
		// 假设扣减通常发生在预占之后（即Confirm），所以减少锁定和总库存
		// 如果是直接扣减（非预占），则应减少可用和总库存
		// 这里暂定为：扣减总库存。至于扣锁定还是可用，取决于业务逻辑调用 Deduct 时的上下文
		// 简单起见，若有锁定则优先扣锁定
		i.TotalStock -= e.Quantity
		if i.LockedStock >= e.Quantity {
			i.LockedStock -= e.Quantity
		} else {
			i.AvailableStock -= e.Quantity
		}
	case *StockAddedEvent:
		i.TotalStock += e.Quantity
		i.AvailableStock += e.Quantity
	case *HedgeConfigUpdatedEvent:
		i.HedgingConfig = e.Config
	}
}

// InventoryLog 库存变更日志
type InventoryLog struct {
	ID             uint64    `json:"id"`
	InventoryID    uint64    `json:"inventory_id"`
	SkuID          uint64    `json:"sku_id"`
	Action         string    `json:"action"`
	ChangeQuantity int32     `json:"change_quantity"`
	OldAvailable   int32     `json:"old_available"`
	NewAvailable   int32     `json:"new_available"`
	OldLocked      int32     `json:"old_locked"`
	NewLocked      int32     `json:"new_locked"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// helper to apply and record
func (i *Inventory) applyAndRecord(event eventsourcing.DomainEvent) {
	i.Apply(event)
	i.ApplyChange(event)
}

func (i *Inventory) Add(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	event := &StockAddedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(StockAddedEventType, i.ID(), i.Version()+1),
		SkuID:     i.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	i.applyAndRecord(event)

	return &InventoryLog{
		InventoryID:    i.DbID,
		SkuID:          i.SkuID,
		ChangeQuantity: quantity,
		NewAvailable:   i.AvailableStock,
		Action:         "ADD",
		Reason:         reason,
		CreatedAt:      time.Now(),
	}, nil
}

func (i *Inventory) Deduct(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	// Strict check? Assuming calling service checked available/locked logic.
	// But aggregate should invariants.
	if i.TotalStock < quantity {
		return nil, fmt.Errorf("insufficient total stock")
	}

	event := &StockDeductedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(StockDeductedEventType, i.ID(), i.Version()+1),
		SkuID:     i.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	i.applyAndRecord(event)

	// Hedging Logic
	if i.HedgingConfig != nil && i.HedgingConfig.Enabled {
		hedgeEvent := &HedgeNeededEvent{
			BaseEvent:    eventsourcing.NewBaseEvent(HedgeNeededEventType, i.ID(), i.Version()+1),
			SkuID:        i.SkuID,
			ChangeQty:    -quantity,
			CurrentStock: i.TotalStock,
			HedgeConfig:  i.HedgingConfig,
		}
		// Hedge event doesn't change state, but we record it
		i.applyAndRecord(hedgeEvent)
	}

	return &InventoryLog{
		InventoryID:    i.DbID,
		SkuID:          i.SkuID,
		ChangeQuantity: -quantity,
		NewAvailable:   i.AvailableStock,
		Action:         "DEDUCT",
		Reason:         reason,
		CreatedAt:      time.Now(),
	}, nil
}

func (i *Inventory) Lock(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if i.AvailableStock < quantity {
		return nil, fmt.Errorf("insufficient available stock")
	}
	event := &StockLockedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(StockLockedEventType, i.ID(), i.Version()+1),
		SkuID:     i.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	i.applyAndRecord(event)

	return &InventoryLog{
		InventoryID:    i.DbID,
		SkuID:          i.SkuID,
		ChangeQuantity: quantity,
		NewLocked:      i.LockedStock,
		Action:         "LOCK",
		Reason:         reason,
		CreatedAt:      time.Now(),
	}, nil
}

func (i *Inventory) Unlock(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if i.LockedStock < quantity {
		return nil, fmt.Errorf("insufficient locked stock to unlock")
	}
	event := &StockUnlockedEvent{
		BaseEvent: eventsourcing.NewBaseEvent(StockUnlockedEventType, i.ID(), i.Version()+1),
		SkuID:     i.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: time.Now(),
	}
	i.applyAndRecord(event)

	return &InventoryLog{
		InventoryID:    i.DbID,
		SkuID:          i.SkuID,
		ChangeQuantity: -quantity,
		Action:         "UNLOCK",
		Reason:         reason,
		CreatedAt:      time.Now(),
	}, nil
}

func (i *Inventory) ConfirmDeduction(quantity int32, reason string) (*InventoryLog, error) {
	return i.Deduct(quantity, reason)
}
