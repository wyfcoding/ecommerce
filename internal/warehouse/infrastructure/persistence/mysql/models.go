package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"gorm.io/gorm"
)

// WarehouseModel 仓库写模型（持久化专用）。
type WarehouseModel struct {
	gorm.Model
	Code          string                 `gorm:"type:varchar(64);uniqueIndex;not null;comment:仓库编码"`
	Name          string                 `gorm:"type:varchar(128);not null;comment:仓库名称"`
	WarehouseType string                 `gorm:"type:varchar(64);comment:仓库类型"`
	Province      string                 `gorm:"type:varchar(64);comment:省"`
	City          string                 `gorm:"type:varchar(64);comment:市"`
	District      string                 `gorm:"type:varchar(64);comment:区"`
	Address       string                 `gorm:"type:varchar(255);comment:详细地址"`
	Longitude     float64                `gorm:"type:decimal(10,6);comment:经度"`
	Latitude      float64                `gorm:"type:decimal(10,6);comment:纬度"`
	ContactName   string                 `gorm:"type:varchar(64);comment:联系人"`
	ContactPhone  string                 `gorm:"type:varchar(32);comment:联系电话"`
	Priority      int32                  `gorm:"default:0;comment:优先级"`
	Status        domain.WarehouseStatus `gorm:"type:varchar(32);not null;default:'INACTIVE';comment:状态"`
	Capacity      int64                  `gorm:"default:0;comment:容量"`
	Description   string                 `gorm:"type:text;comment:描述"`
}

func (WarehouseModel) TableName() string {
	return "warehouses"
}

// WarehouseStockModel 库存写模型（持久化专用）。
type WarehouseStockModel struct {
	gorm.Model
	WarehouseID uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:仓库ID"`
	SkuID       uint64 `gorm:"uniqueIndex:idx_wh_sku;not null;comment:SKU ID"`
	Stock       int32  `gorm:"not null;default:0;comment:库存数量"`
	LockedStock int32  `gorm:"not null;default:0;comment:锁定库存"`
	SafeStock   int32  `gorm:"not null;default:0;comment:安全库存"`
	MaxStock    int32  `gorm:"not null;default:0;comment:最大库存"`
}

func (WarehouseStockModel) TableName() string {
	return "warehouse_stocks"
}

// StockTransferModel 调拨单写模型（持久化专用）。
type StockTransferModel struct {
	gorm.Model
	TransferNo      string                     `gorm:"type:varchar(64);uniqueIndex;not null;comment:调拨单号"`
	FromWarehouseID uint64                     `gorm:"index;not null;comment:调出仓库ID"`
	ToWarehouseID   uint64                     `gorm:"index;not null;comment:调入仓库ID"`
	SkuID           uint64                     `gorm:"index;not null;comment:SKU ID"`
	Quantity        int32                      `gorm:"not null;comment:调拨数量"`
	Status          domain.StockTransferStatus `gorm:"type:varchar(32);not null;default:'PENDING';comment:状态"`
	Reason          string                     `gorm:"type:varchar(255);comment:调拨原因"`
	ApprovedBy      uint64                     `gorm:"comment:审核人ID"`
	ApprovedAt      *time.Time                 `gorm:"comment:审核时间"`
	ShippedAt       *time.Time                 `gorm:"comment:发货时间"`
	ReceivedAt      *time.Time                 `gorm:"comment:收货时间"`
	CompletedAt     *time.Time                 `gorm:"comment:完成时间"`
	Remark          string                     `gorm:"type:text;comment:备注"`
	CreatedBy       uint64                     `gorm:"not null;comment:创建人ID"`
}

func (StockTransferModel) TableName() string {
	return "stock_transfers"
}

func toWarehouseModel(warehouse *domain.Warehouse) *WarehouseModel {
	if warehouse == nil {
		return nil
	}
	return &WarehouseModel{
		Model: gorm.Model{
			ID:        warehouse.ID,
			CreatedAt: warehouse.CreatedAt,
			UpdatedAt: warehouse.UpdatedAt,
		},
		Code:          warehouse.Code,
		Name:          warehouse.Name,
		WarehouseType: warehouse.WarehouseType,
		Province:      warehouse.Province,
		City:          warehouse.City,
		District:      warehouse.District,
		Address:       warehouse.Address,
		Longitude:     warehouse.Longitude,
		Latitude:      warehouse.Latitude,
		ContactName:   warehouse.ContactName,
		ContactPhone:  warehouse.ContactPhone,
		Priority:      warehouse.Priority,
		Status:        warehouse.Status,
		Capacity:      warehouse.Capacity,
		Description:   warehouse.Description,
	}
}

func toDomainWarehouse(model *WarehouseModel) *domain.Warehouse {
	if model == nil {
		return nil
	}
	return &domain.Warehouse{
		ID:            model.ID,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		Code:          model.Code,
		Name:          model.Name,
		WarehouseType: model.WarehouseType,
		Province:      model.Province,
		City:          model.City,
		District:      model.District,
		Address:       model.Address,
		Longitude:     model.Longitude,
		Latitude:      model.Latitude,
		ContactName:   model.ContactName,
		ContactPhone:  model.ContactPhone,
		Priority:      model.Priority,
		Status:        model.Status,
		Capacity:      model.Capacity,
		Description:   model.Description,
	}
}

func toWarehouseStockModel(stock *domain.WarehouseStock) *WarehouseStockModel {
	if stock == nil {
		return nil
	}
	return &WarehouseStockModel{
		Model: gorm.Model{
			ID:        stock.ID,
			CreatedAt: stock.CreatedAt,
			UpdatedAt: stock.UpdatedAt,
		},
		WarehouseID: stock.WarehouseID,
		SkuID:       stock.SkuID,
		Stock:       stock.Stock,
		LockedStock: stock.LockedStock,
		SafeStock:   stock.SafeStock,
		MaxStock:    stock.MaxStock,
	}
}

func toDomainWarehouseStock(model *WarehouseStockModel) *domain.WarehouseStock {
	if model == nil {
		return nil
	}
	return &domain.WarehouseStock{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		WarehouseID: model.WarehouseID,
		SkuID:       model.SkuID,
		Stock:       model.Stock,
		LockedStock: model.LockedStock,
		SafeStock:   model.SafeStock,
		MaxStock:    model.MaxStock,
	}
}

func toStockTransferModel(transfer *domain.StockTransfer) *StockTransferModel {
	if transfer == nil {
		return nil
	}
	return &StockTransferModel{
		Model: gorm.Model{
			ID:        transfer.ID,
			CreatedAt: transfer.CreatedAt,
			UpdatedAt: transfer.UpdatedAt,
		},
		TransferNo:      transfer.TransferNo,
		FromWarehouseID: transfer.FromWarehouseID,
		ToWarehouseID:   transfer.ToWarehouseID,
		SkuID:           transfer.SkuID,
		Quantity:        transfer.Quantity,
		Status:          transfer.Status,
		Reason:          transfer.Reason,
		ApprovedBy:      transfer.ApprovedBy,
		ApprovedAt:      transfer.ApprovedAt,
		ShippedAt:       transfer.ShippedAt,
		ReceivedAt:      transfer.ReceivedAt,
		CompletedAt:     transfer.CompletedAt,
		Remark:          transfer.Remark,
		CreatedBy:       transfer.CreatedBy,
	}
}

func toDomainStockTransfer(model *StockTransferModel) *domain.StockTransfer {
	if model == nil {
		return nil
	}
	return &domain.StockTransfer{
		ID:              model.ID,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
		TransferNo:      model.TransferNo,
		FromWarehouseID: model.FromWarehouseID,
		ToWarehouseID:   model.ToWarehouseID,
		SkuID:           model.SkuID,
		Quantity:        model.Quantity,
		Status:          model.Status,
		Reason:          model.Reason,
		ApprovedBy:      model.ApprovedBy,
		ApprovedAt:      model.ApprovedAt,
		ShippedAt:       model.ShippedAt,
		ReceivedAt:      model.ReceivedAt,
		CompletedAt:     model.CompletedAt,
		Remark:          model.Remark,
		CreatedBy:       model.CreatedBy,
	}
}
