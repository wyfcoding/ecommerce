package domain

import "gorm.io/gorm" // 导入GORM库。

// Warehouse 实体代表一个仓库。
// 它包含了仓库的名称、地理位置、优先级和基础配送成本等信息。
type Warehouse struct {
	gorm.Model         // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name       string  `gorm:"type:varchar(255);not null;comment:仓库名称" json:"name"`   // 仓库名称，不允许为空。
	Lat        float64 `gorm:"type:decimal(10,6);not null;comment:纬度" json:"lat"`     // 仓库的地理纬度。
	Lon        float64 `gorm:"type:decimal(10,6);not null;comment:经度" json:"lon"`     // 仓库的地理经度。
	Priority   int     `gorm:"not null;default:0;comment:优先级" json:"priority"`        // 仓库的优先级，数字越大优先级越高。
	ShipCost   int64   `gorm:"not null;default:0;comment:基础配送成本(分)" json:"ship_cost"` // 从该仓库发货的基础配送成本（单位：分）。
}

// NewWarehouse 创建并返回一个新的 Warehouse 实体实例。
// name: 仓库名称。
// lat, lon: 地理坐标（纬度、经度）。
// priority: 优先级。
// shipCost: 基础配送成本。
func NewWarehouse(name string, lat, lon float64, priority int, shipCost int64) *Warehouse {
	return &Warehouse{
		Name:     name,
		Lat:      lat,
		Lon:      lon,
		Priority: priority,
		ShipCost: shipCost,
	}
}
