package repository

import (
	"gorm.io/gorm"
)

// ShippingRule represents a shipping cost calculation rule.
type ShippingRule struct {
	gorm.Model
	Name        string  `gorm:"size:255;not null;comment:规则名称" json:"name"`
	Origin      string  `gorm:"size:255;comment:发货地 (e.g., province, city)" json:"origin"`
	Destination string  `gorm:"size:255;comment:目的地 (e.g., province, city)" json:"destination"`
	MinWeight   float64 `gorm:"comment:最小重量 (kg)" json:"minWeight"`
	MaxWeight   float64 `gorm:"comment:最大重量 (kg)" json:"maxWeight"`
	BaseCost    uint64  `gorm:"not null;comment:基础运费 (分)" json:"baseCost"`
	PerKgCost   uint64  `gorm:"comment:每公斤额外费用 (分)" json:"perKgCost"`
	// Add other rule parameters like volume, item type, etc.
}

// Address represents a simplified address for shipping rules.
type Address struct {
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
}

// TableName specifies the table name for ShippingRule.
func (ShippingRule) TableName() string {
	return "shipping_rules"
}
