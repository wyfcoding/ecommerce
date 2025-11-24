package entity

import "gorm.io/gorm"

// Warehouse 仓库实体
type Warehouse struct {
	gorm.Model
	Name     string  `gorm:"type:varchar(255);not null;comment:仓库名称" json:"name"`
	Lat      float64 `gorm:"type:decimal(10,6);not null;comment:纬度" json:"lat"`
	Lon      float64 `gorm:"type:decimal(10,6);not null;comment:经度" json:"lon"`
	Priority int     `gorm:"not null;default:0;comment:优先级" json:"priority"`
	ShipCost int64   `gorm:"not null;default:0;comment:基础配送成本(分)" json:"ship_cost"`
}

// NewWarehouse 创建仓库
func NewWarehouse(name string, lat, lon float64, priority int, shipCost int64) *Warehouse {
	return &Warehouse{
		Name:     name,
		Lat:      lat,
		Lon:      lon,
		Priority: priority,
		ShipCost: shipCost,
	}
}
