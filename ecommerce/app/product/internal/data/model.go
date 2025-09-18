package data

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONStringArray a custom type for string arrays stored as JSON
type JSONStringArray []string

func (a *JSONStringArray) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), a)
}

func (a JSONStringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// JSONMap a custom type for map[string]string stored as JSON
type JSONMap map[string]string

func (m *JSONMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// ProductCategory 商品分类模型
type ProductCategory struct {
	ID        uint64 `gorm:"primarykey"`
	ParentID  uint64 `gorm:"index:idx_parent_id;not null;default:0"`
	Name      string `gorm:"size:64;not null"`
	Level     uint8  `gorm:"not null;default:1"`
	Icon      string `gorm:"size:255;default:''"`
	SortOrder uint   `gorm:"not null;default:0"`
	IsVisible bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (ProductCategory) TableName() string {
	return "product_category"
}

// ProductSpu 商品SPU模型
type ProductSpu struct {
	ID            uint64          `gorm:"primarykey"`
	SpuID         uint64          `gorm:"uniqueIndex:uk_spu_id;not null"`
	CategoryID    uint64          `gorm:"index:idx_category_id;not null"`
	BrandID       uint64          `gorm:"index:idx_brand_id;not null;default:0"`
	Title         string          `gorm:"size:255;not null"`
	SubTitle      string          `gorm:"size:255"`
	MainImage     string          `gorm:"size:255;not null"`
	GalleryImages JSONStringArray `gorm:"type:json"`
	DetailHTML    string          `gorm:"type:text"`
	Status        int8            `gorm:"not null;default:1"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (ProductSpu) TableName() string {
	return "product_spu"
}

// ProductSku 商品SKU模型
type ProductSku struct {
	ID            uint64  `gorm:"primarykey"`
	SkuID         uint64  `gorm:"uniqueIndex:uk_sku_id;not null"`
	SpuID         uint64  `gorm:"index:idx_spu_id;not null"`
	Title         string  `gorm:"size:255;not null"`
	Price         uint64  `gorm:"not null"`
	OriginalPrice uint64  `gorm:"not null"`
	Stock         uint    `gorm:"not null;default:0"`
	Image         string  `gorm:"size:255;default:''"`
	Specs         JSONMap `gorm:"type:json"`
	Status        int8    `gorm:"not null;default:1"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (ProductSku) TableName() string {
	return "product_sku"
}
