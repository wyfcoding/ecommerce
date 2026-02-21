package domain

import (
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

type LegacyInventory struct {
	gorm.Model
	database.BaseEntity
	SKU      string `gorm:"column:sku;uniqueIndex;not null"`
	Stock    int32  `gorm:"column:stock;not null"`
	Reserved int32  `gorm:"column:reserved;default:0"`
}
