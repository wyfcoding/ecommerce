package domain

import (
	"time" // 导入时间库。
)

// WarehouseStatus 定义了仓库的运营状态。
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "ACTIVE"      // 启用：仓库正常运营，可进行入库、出库等操作。
	WarehouseStatusInactive    WarehouseStatus = "INACTIVE"    // 禁用：仓库暂停运营，通常不能进行任何库存操作。
	WarehouseStatusMaintenance WarehouseStatus = "MAINTENANCE" // 维护中：仓库正在维护，部分功能可能受限。
)

// Warehouse 实体代表一个物理或逻辑仓库。
// 它包含了仓库的基本信息、地址、联系方式、状态和容量等。
type Warehouse struct {
	ID            uint            `json:"id"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Code          string          `json:"code"`           // 仓库唯一代码，唯一索引，不允许为空。
	Name          string          `json:"name"`           // 仓库名称。
	WarehouseType string          `json:"warehouse_type"` // 仓库类型（例如，主仓，中转仓，前置仓）。
	Province      string          `json:"province"`       // 省份。
	City          string          `json:"city"`           // 城市。
	District      string          `json:"district"`       // 区/县。
	Address       string          `json:"address"`        // 详细地址。
	Longitude     float64         `json:"longitude"`      // 地理经度。
	Latitude      float64         `json:"latitude"`       // 地理纬度。
	ContactName   string          `json:"contact_name"`   // 仓库联系人姓名。
	ContactPhone  string          `json:"contact_phone"`  // 仓库联系人电话。
	Priority      int32           `json:"priority"`       // 仓库优先级，用于调度决策。
	Status        WarehouseStatus `json:"status"`         // 仓库状态，默认为非活跃。
	Capacity      int64           `json:"capacity"`       // 仓库容量。
	Description   string          `json:"description"`    // 仓库描述。
}

// WarehouseStock 实体代表仓库中某个SKU的库存信息。
// 它是仓库聚合根的一部分。
type WarehouseStock struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	WarehouseID uint64    `json:"warehouse_id"` // 关联的仓库ID，与SkuID共同构成唯一索引。
	SkuID       uint64    `json:"sku_id"`       // 关联的SKU ID，与WarehouseID共同构成唯一索引。
	Stock       int32     `json:"stock"`        // 当前库存数量。
	LockedStock int32     `json:"locked_stock"` // 已被锁定（例如，被订单预留）的库存数量。
	SafeStock   int32     `json:"safe_stock"`   // 安全库存数量，低于此值应触发补货。
	MaxStock    int32     `json:"max_stock"`    // 最大库存数量。
}

// AvailableStock 计算SKU的可用库存数量。
// 可用库存 = 总库存 - 锁定库存。
func (s *WarehouseStock) AvailableStock() int32 {
	return s.Stock - s.LockedStock
}

// StockTransferStatus 定义了库存调拨单的生命周期状态。
type StockTransferStatus string

const (
	StockTransferStatusPending   StockTransferStatus = "PENDING"   // 待处理：调拨单已创建，等待审批或执行。
	StockTransferStatusApproved  StockTransferStatus = "APPROVED"  // 已审核：调拨单已通过审批。
	StockTransferStatusShipped   StockTransferStatus = "SHIPPED"   // 已发货：商品已从调出仓库发出。
	StockTransferStatusReceived  StockTransferStatus = "RECEIVED"  // 已收货：商品已到达调入仓库。
	StockTransferStatusCompleted StockTransferStatus = "COMPLETED" // 已完成：调拨流程全部完成。
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED" // 已取消：调拨单被取消。
)

// StockTransfer 实体代表一次库存调拨。
// 它记录了商品从一个仓库调拨到另一个仓库的详细信息和状态。
type StockTransfer struct {
	ID              uint                `json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	TransferNo      string              `json:"transfer_no"`       // 调拨单号，唯一索引，不允许为空。
	FromWarehouseID uint64              `json:"from_warehouse_id"` // 调出仓库的ID，索引字段。
	ToWarehouseID   uint64              `json:"to_warehouse_id"`   // 调入仓库的ID，索引字段。
	SkuID           uint64              `json:"sku_id"`            // 调拨的SKU ID，索引字段。
	Quantity        int32               `json:"quantity"`          // 调拨数量。
	Status          StockTransferStatus `json:"status"`            // 调拨单状态，默认为待处理。
	Reason          string              `json:"reason"`            // 调拨原因。
	ApprovedBy      uint64              `json:"approved_by"`       // 审批调拨单的人员ID。
	ApprovedAt      *time.Time          `json:"approved_at"`       // 审批时间。
	ShippedAt       *time.Time          `json:"shipped_at"`        // 发货时间。
	ReceivedAt      *time.Time          `json:"received_at"`       // 收货时间。
	CompletedAt     *time.Time          `json:"completed_at"`      // 完成时间。
	Remark          string              `json:"remark"`            // 备注信息。
	CreatedBy       uint64              `json:"created_by"`        // 调拨单创建人ID。
}
