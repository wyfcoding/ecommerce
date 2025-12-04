package entity

import (
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
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
	gorm.Model                    // 嵌入gorm.Model。
	Code          string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:仓库编码" json:"code"`        // 仓库唯一代码，唯一索引，不允许为空。
	Name          string          `gorm:"type:varchar(128);not null;comment:仓库名称" json:"name"`                   // 仓库名称。
	WarehouseType string          `gorm:"type:varchar(64);comment:仓库类型" json:"warehouse_type"`                   // 仓库类型（例如，主仓，中转仓，前置仓）。
	Province      string          `gorm:"type:varchar(64);comment:省" json:"province"`                            // 省份。
	City          string          `gorm:"type:varchar(64);comment:市" json:"city"`                                // 城市。
	District      string          `gorm:"type:varchar(64);comment:区" json:"district"`                            // 区/县。
	Address       string          `gorm:"type:varchar(255);comment:详细地址" json:"address"`                         // 详细地址。
	Longitude     float64         `gorm:"type:decimal(10,6);comment:经度" json:"longitude"`                        // 地理经度。
	Latitude      float64         `gorm:"type:decimal(10,6);comment:纬度" json:"latitude"`                         // 地理纬度。
	ContactName   string          `gorm:"type:varchar(64);comment:联系人" json:"contact_name"`                      // 仓库联系人姓名。
	ContactPhone  string          `gorm:"type:varchar(32);comment:联系电话" json:"contact_phone"`                    // 仓库联系人电话。
	Priority      int32           `gorm:"default:0;comment:优先级" json:"priority"`                                 // 仓库优先级，用于调度决策。
	Status        WarehouseStatus `gorm:"type:varchar(32);not null;default:'INACTIVE';comment:状态" json:"status"` // 仓库状态，默认为非活跃。
	Capacity      int64           `gorm:"default:0;comment:容量" json:"capacity"`                                  // 仓库容量。
	Description   string          `gorm:"type:text;comment:描述" json:"description"`                               // 仓库描述。
}

// WarehouseStock 实体代表仓库中某个SKU的库存信息。
// 它是仓库聚合根的一部分。
type WarehouseStock struct {
	gorm.Model         // 嵌入gorm.Model。
	WarehouseID uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:仓库ID" json:"warehouse_id"` // 关联的仓库ID，与SkuID共同构成唯一索引。
	SkuID       uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:SKU ID" json:"sku_id"`     // 关联的SKU ID，与WarehouseID共同构成唯一索引。
	Stock       int32  `gorm:"not null;default:0;comment:库存数量" json:"stock"`                     // 当前库存数量。
	LockedStock int32  `gorm:"not null;default:0;comment:锁定库存" json:"locked_stock"`              // 已被锁定（例如，被订单预留）的库存数量。
	SafeStock   int32  `gorm:"not null;default:0;comment:安全库存" json:"safe_stock"`                // 安全库存数量，低于此值应触发补货。
	MaxStock    int32  `gorm:"not null;default:0;comment:最大库存" json:"max_stock"`                 // 最大库存数量。
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
	gorm.Model                          // 嵌入gorm.Model。
	TransferNo      string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:调拨单号" json:"transfer_no"` // 调拨单号，唯一索引，不允许为空。
	FromWarehouseID uint64              `gorm:"index;not null;comment:调出仓库ID" json:"from_warehouse_id"`                // 调出仓库的ID，索引字段。
	ToWarehouseID   uint64              `gorm:"index;not null;comment:调入仓库ID" json:"to_warehouse_id"`                  // 调入仓库的ID，索引字段。
	SkuID           uint64              `gorm:"index;not null;comment:SKU ID" json:"sku_id"`                           // 调拨的SKU ID，索引字段。
	Quantity        int32               `gorm:"not null;comment:调拨数量" json:"quantity"`                                 // 调拨数量。
	Status          StockTransferStatus `gorm:"type:varchar(32);not null;default:'PENDING';comment:状态" json:"status"`  // 调拨单状态，默认为待处理。
	Reason          string              `gorm:"type:varchar(255);comment:调拨原因" json:"reason"`                          // 调拨原因。
	ApprovedBy      uint64              `gorm:"comment:审核人ID" json:"approved_by"`                                      // 审批调拨单的人员ID。
	ApprovedAt      *time.Time          `gorm:"comment:审核时间" json:"approved_at"`                                       // 审批时间。
	ShippedAt       *time.Time          `gorm:"comment:发货时间" json:"shipped_at"`                                        // 发货时间。
	ReceivedAt      *time.Time          `gorm:"comment:收货时间" json:"received_at"`                                       // 收货时间。
	CompletedAt     *time.Time          `gorm:"comment:完成时间" json:"completed_at"`                                      // 完成时间。
	Remark          string              `gorm:"type:text;comment:备注" json:"remark"`                                    // 备注信息。
	CreatedBy       uint64              `gorm:"not null;comment:创建人ID" json:"created_by"`                              // 调拨单创建人ID。
}
