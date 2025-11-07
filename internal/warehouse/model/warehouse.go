package model

import "time"

// WarehouseStatus 仓库状态
type WarehouseStatus string

const (
	WarehouseStatusActive   WarehouseStatus = "ACTIVE"   // 启用
	WarehouseStatusInactive WarehouseStatus = "INACTIVE" // 禁用
	WarehouseStatusMaintenance WarehouseStatus = "MAINTENANCE" // 维护中
)

// Warehouse 仓库
type Warehouse struct {
	ID          uint64          `gorm:"primarykey" json:"id"`
	Code        string          `gorm:"type:varchar(50);uniqueIndex;not null;comment:仓库编码" json:"code"`
	Name        string          `gorm:"type:varchar(255);not null;comment:仓库名称" json:"name"`
	Type        string          `gorm:"type:varchar(20);not null;comment:仓库类型(CENTRAL,REGIONAL,LOCAL)" json:"type"`
	Province    string          `gorm:"type:varchar(50);not null;comment:省份" json:"province"`
	City        string          `gorm:"type:varchar(50);not null;comment:城市" json:"city"`
	District    string          `gorm:"type:varchar(50);comment:区县" json:"district"`
	Address     string          `gorm:"type:varchar(500);not null;comment:详细地址" json:"address"`
	Longitude   float64         `gorm:"comment:经度" json:"longitude"`
	Latitude    float64         `gorm:"comment:纬度" json:"latitude"`
	ContactName string          `gorm:"type:varchar(100);comment:联系人" json:"contactName"`
	ContactPhone string         `gorm:"type:varchar(20);comment:联系电话" json:"contactPhone"`
	Priority    int32           `gorm:"not null;default:0;comment:优先级(数字越大优先级越高)" json:"priority"`
	Status      WarehouseStatus `gorm:"type:varchar(20);not null;comment:仓库状态" json:"status"`
	Capacity    int64           `gorm:"comment:仓库容量" json:"capacity"`
	Description string          `gorm:"type:text;comment:仓库描述" json:"description"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time      `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (Warehouse) TableName() string {
	return "warehouses"
}

// WarehouseStock 仓库库存
type WarehouseStock struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	WarehouseID uint64     `gorm:"index:idx_warehouse_sku;not null;comment:仓库ID" json:"warehouseId"`
	SkuID       uint64     `gorm:"index:idx_warehouse_sku;not null;comment:SKU ID" json:"skuId"`
	Stock       int32      `gorm:"not null;default:0;comment:可用库存" json:"stock"`
	LockedStock int32      `gorm:"not null;default:0;comment:锁定库存" json:"lockedStock"`
	SafeStock   int32      `gorm:"not null;default:0;comment:安全库存" json:"safeStock"`
	MaxStock    int32      `gorm:"comment:最大库存" json:"maxStock"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (WarehouseStock) TableName() string {
	return "warehouse_stocks"
}

// GetTotalStock 获取总库存
func (ws *WarehouseStock) GetTotalStock() int32 {
	return ws.Stock + ws.LockedStock
}

// GetAvailableStock 获取可用库存
func (ws *WarehouseStock) GetAvailableStock() int32 {
	return ws.Stock
}

// IsLowStock 是否低库存
func (ws *WarehouseStock) IsLowStock() bool {
	return ws.Stock <= ws.SafeStock
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

// StockTransfer 库存调拨
type StockTransfer struct {
	ID              uint64              `gorm:"primarykey" json:"id"`
	TransferNo      string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:调拨单号" json:"transferNo"`
	FromWarehouseID uint64              `gorm:"index;not null;comment:源仓库ID" json:"fromWarehouseId"`
	ToWarehouseID   uint64              `gorm:"index;not null;comment:目标仓库ID" json:"toWarehouseId"`
	SkuID           uint64              `gorm:"index;not null;comment:SKU ID" json:"skuId"`
	Quantity        int32               `gorm:"not null;comment:调拨数量" json:"quantity"`
	Status          StockTransferStatus `gorm:"type:varchar(20);not null;comment:调拨状态" json:"status"`
	Reason          string              `gorm:"type:varchar(500);comment:调拨原因" json:"reason"`
	ApprovedBy      uint64              `gorm:"comment:审核人ID" json:"approvedBy"`
	ApprovedAt      *time.Time          `gorm:"comment:审核时间" json:"approvedAt"`
	ShippedAt       *time.Time          `gorm:"comment:发货时间" json:"shippedAt"`
	ReceivedAt      *time.Time          `gorm:"comment:收货时间" json:"receivedAt"`
	CompletedAt     *time.Time          `gorm:"comment:完成时间" json:"completedAt"`
	Remark          string              `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy       uint64              `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt       *time.Time          `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (StockTransfer) TableName() string {
	return "stock_transfers"
}

// StockAllocation 库存分配记录
type StockAllocation struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	OrderID     uint64     `gorm:"index;not null;comment:订单ID" json:"orderId"`
	OrderItemID uint64     `gorm:"index;not null;comment:订单项ID" json:"orderItemId"`
	WarehouseID uint64     `gorm:"index;not null;comment:仓库ID" json:"warehouseId"`
	SkuID       uint64     `gorm:"index;not null;comment:SKU ID" json:"skuId"`
	Quantity    int32      `gorm:"not null;comment:分配数量" json:"quantity"`
	Status      string     `gorm:"type:varchar(20);not null;comment:分配状态" json:"status"`
	AllocatedAt time.Time  `gorm:"not null;comment:分配时间" json:"allocatedAt"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (StockAllocation) TableName() string {
	return "stock_allocations"
}

// StocktakingStatus 盘点状态
type StocktakingStatus string

const (
	StocktakingStatusPending   StocktakingStatus = "PENDING"   // 待盘点
	StocktakingStatusInProgress StocktakingStatus = "IN_PROGRESS" // 盘点中
	StocktakingStatusCompleted StocktakingStatus = "COMPLETED" // 已完成
	StocktakingStatusCancelled StocktakingStatus = "CANCELLED" // 已取消
)

// Stocktaking 库存盘点
type Stocktaking struct {
	ID          uint64            `gorm:"primarykey" json:"id"`
	StockNo     string            `gorm:"type:varchar(64);uniqueIndex;not null;comment:盘点单号" json:"stockNo"`
	WarehouseID uint64            `gorm:"index;not null;comment:仓库ID" json:"warehouseId"`
	Type        string            `gorm:"type:varchar(20);not null;comment:盘点类型(FULL,PARTIAL)" json:"type"`
	Status      StocktakingStatus `gorm:"type:varchar(20);not null;comment:盘点状态" json:"status"`
	StartTime   time.Time         `gorm:"not null;comment:开始时间" json:"startTime"`
	EndTime     *time.Time        `gorm:"comment:结束时间" json:"endTime"`
	Remark      string            `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy   uint64            `gorm:"not null;comment:创建人ID" json:"createdBy"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time        `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (Stocktaking) TableName() string {
	return "stocktakings"
}

// StocktakingItem 盘点明细
type StocktakingItem struct {
	ID            uint64     `gorm:"primarykey" json:"id"`
	StocktakingID uint64     `gorm:"index;not null;comment:盘点ID" json:"stocktakingId"`
	SkuID         uint64     `gorm:"index;not null;comment:SKU ID" json:"skuId"`
	BookStock     int32      `gorm:"not null;comment:账面库存" json:"bookStock"`
	ActualStock   int32      `gorm:"comment:实际库存" json:"actualStock"`
	Difference    int32      `gorm:"comment:差异数量" json:"difference"`
	Remark        string     `gorm:"type:varchar(500);comment:备注" json:"remark"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (StocktakingItem) TableName() string {
	return "stocktaking_items"
}

// CalculateDifference 计算差异
func (si *StocktakingItem) CalculateDifference() {
	si.Difference = si.ActualStock - si.BookStock
}

// WarehouseArea 仓库区域（用于仓库内部管理）
type WarehouseArea struct {
	ID          uint64     `gorm:"primarykey" json:"id"`
	WarehouseID uint64     `gorm:"index;not null;comment:仓库ID" json:"warehouseId"`
	Code        string     `gorm:"type:varchar(50);not null;comment:区域编码" json:"code"`
	Name        string     `gorm:"type:varchar(100);not null;comment:区域名称" json:"name"`
	Type        string     `gorm:"type:varchar(20);comment:区域类型(STORAGE,PICKING,PACKING)" json:"type"`
	Capacity    int32      `gorm:"comment:容量" json:"capacity"`
	IsActive    bool       `gorm:"not null;default:true;comment:是否启用" json:"isActive"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

// TableName 指定表名
func (WarehouseArea) TableName() string {
	return "warehouse_areas"
}
