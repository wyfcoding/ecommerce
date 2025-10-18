package model

import (
	"time"

	"gorm.io/gorm"
)

// Product 商品模型
// 包含了商品的核心信息
type Product struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`         // 商品名称
	SKU             string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"` // 商品库存单位 (Stock Keeping Unit)
	Description     string         `gorm:"type:text" json:"description"`                     // 商品详细描述
	Price           float64        `gorm:"type:decimal(10,2);not null" json:"price"`        // 商品价格
	Stock           int            `gorm:"not null;default:0" json:"stock"`                 // 库存数量
	CategoryID      uint           `gorm:"not null;index" json:"category_id"`              // 分类ID
	Category        Category       `gorm:"foreignKey:CategoryID" json:"category"`           // 关联的分类
	BrandID         uint           `gorm:"not null;index" json:"brand_id"`                 // 品牌ID
	Brand           Brand          `gorm:"foreignKey:BrandID" json:"brand"`                 // 关联的品牌
	IsPublished     bool           `gorm:"not null;default:false" json:"is_published"`      // 是否上架
	MainImageURL    string         `gorm:"type:varchar(255)" json:"main_image_url"`         // 商品主图
	AdditionalImages string        `gorm:"type:text" json:"additional_images"`              // 更多图片 (JSON 字符串或逗号分隔)
	Weight          float64        `gorm:"type:decimal(10,2)" json:"weight"`                // 商品重量 (kg)
	Dimensions      string         `gorm:"type:varchar(100)" json:"dimensions"`             // 尺寸 (长x宽x高)
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Category 商品分类模型
type Category struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 分类名称
	ParentID    *uint     `gorm:"index" json:"parent_id"`                            // 父分类ID，用于支持多级分类
	Description string    `gorm:"type:varchar(255)" json:"description"`              // 分类描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Brand 商品品牌模型
type Brand struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"` // 品牌名称
	LogoURL     string    `gorm:"type:varchar(255)" json:"logo_url"`                 // 品牌Logo
	Description string    `gorm:"type:text" json:"description"`                      // 品牌故事
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 自定义各模型的数据库表名
func (Product) TableName() string {
	return "products"
}

func (Category) TableName() string {
	return "categories"
}

func (Brand) TableName() string {
	return "brands"
}
