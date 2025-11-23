package entity

import (
	"time"

	"gorm.io/gorm"
)

// WarehouseStatus 仓库状态
type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "ACTIVE"      // 启用
	WarehouseStatusInactive    WarehouseStatus = "INACTIVE"    // 禁用
	WarehouseStatusMaintenance WarehouseStatus = "MAINTENANCE" // 维护中
)

// Warehouse 仓库实体
type Warehouse struct {
	gorm.Model
	Code          string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:仓库编码" json:"code"`
	Name          string          `gorm:"type:varchar(128);not null;comment:仓库名称" json:"name"`
	WarehouseType string          `gorm:"type:varchar(64);comment:仓库类型" json:"warehouse_type"`
	Province      string          `gorm:"type:varchar(64);comment:省" json:"province"`
	City          string          `gorm:"type:varchar(64);comment:市" json:"city"`
	District      string          `gorm:"type:varchar(64);comment:区" json:"district"`
	Address       string          `gorm:"type:varchar(255);comment:详细地址" json:"address"`
	Longitude     float64         `gorm:"type:decimal(10,6);comment:经度" json:"longitude"`
	Latitude      float64         `gorm:"type:decimal(10,6);comment:纬度" json:"latitude"`
	ContactName   string          `gorm:"type:varchar(64);comment:联系人" json:"contact_name"`
	ContactPhone  string          `gorm:"type:varchar(32);comment:联系电话" json:"contact_phone"`
	Priority      int32           `gorm:"default:0;comment:优先级" json:"priority"`
	Status        WarehouseStatus `gorm:"type:varchar(32);not null;default:'INACTIVE';comment:状态" json:"status"`
	Capacity      int64           `gorm:"default:0;comment:容量" json:"capacity"`
	Description   string          `gorm:"type:text;comment:描述" json:"description"`
}

// WarehouseStock 仓库库存实体
type WarehouseStock struct {
	gorm.Model
	WarehouseID uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:仓库ID" json:"warehouse_id"`
	SkuID       uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:SKU ID" json:"sku_id"`
	Stock       int32  `gorm:"not null;default:0;comment:库存数量" json:"stock"`
	LockedStock int32  `gorm:"not null;default:0;comment:锁定库存" json:"locked_stock"`
	SafeStock   int32  `gorm:"not null;default:0;comment:安全库存" json:"safe_stock"`
	MaxStock    int32  `gorm:"not null;default:0;comment:最大库存" json:"max_stock"`
}

// AvailableStock 可用库存
func (s *WarehouseStock) AvailableStock() int32 {
	return s.Stock - s.LockedStock
}

// StockTransferStatus 库存调拨状态
type StockTransferStatus string

const (
	StockTransferStatusPending   StockTransferStatus = "PENDING"   // 待处理
	StockTransferStatusApproved  StockTransferStatus = "APPROVED"  // 已审核
	StockTransferStatusShipped   StockTransferStatus = "SHIPPED"   // 已发货
	StockTransferStatusReceived  StockTransferStatus = "RECEIVED"  // 已收货
	StockTransferStatusCompleted StockTransferStatus = "COMPLETED" // 已完成
	StockTransferStatusCancelled StockTransferStatus = "CANCELLED" // 已取消
)

// StockTransfer 库存调拨实体
type StockTransfer struct {
	gorm.Model
	TransferNo      string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:调拨单号" json:"transfer_no"`
	FromWarehouseID uint64              `gorm:"index;not null;comment:调出仓库ID" json:"from_warehouse_id"`
	ToWarehouseID   uint64              `gorm:"index;not null;comment:调入仓库ID" json:"to_warehouse_id"`
	SkuID           uint64              `gorm:"index;not null;comment:SKU ID" json:"sku_id"`
	Quantity        int32               `gorm:"not null;comment:调拨数量" json:"quantity"`
	Status          StockTransferStatus `gorm:"type:varchar(32);not null;default:'PENDING';comment:状态" json:"status"`
	Reason          string              `gorm:"type:varchar(255);comment:调拨原因" json:"reason"`
	ApprovedBy      uint64              `gorm:"comment:审核人ID" json:"approved_by"`
	ApprovedAt      *time.Time          `gorm:"comment:审核时间" json:"approved_at"`
	ShippedAt       *time.Time          `gorm:"comment:发货时间" json:"shipped_at"`
	ReceivedAt      *time.Time          `gorm:"comment:收货时间" json:"received_at"`
	CompletedAt     *time.Time          `gorm:"comment:完成时间" json:"completed_at"`
	Remark          string              `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy       uint64              `gorm:"not null;comment:创建人ID" json:"created_by"`
}
