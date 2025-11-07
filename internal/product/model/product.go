package model

import (
	"time"
)

// ProductStatus 定义商品状态的枚举。
type ProductStatus int32

const (
	ProductStatusUnspecified ProductStatus = 0 // 未指定
	ProductStatusDraft       ProductStatus = 1 // 草稿
	ProductStatusActive      ProductStatus = 2 // 上架/销售中
	ProductStatusInactive    ProductStatus = 3 // 下架
	ProductStatusDiscontinued ProductStatus = 4 // 停产
	ProductStatusArchived    ProductStatus = 5 // 归档
)

// Product 代表一个标准商品单元 (SPU - Standard Product Unit)。
// 它是商品的抽象概念，包含商品的通用信息，不涉及具体库存和价格。
type Product struct {
	ID               uint64        `gorm:"primarykey" json:"id"`                               // 商品SPU ID
	Name             string        `gorm:"type:varchar(255);not null" json:"name"`             // 商品名称
	SpuNo            string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"spu_no"` // SPU编码, 例如 "SPU123456"
	Description      string        `gorm:"type:text" json:"description"`                       // 商品详细描述 (支持HTML)
	CategoryID       uint64        `gorm:"index;not null" json:"category_id"`                  // 所属分类ID
	Category         Category      `gorm:"foreignKey:CategoryID" json:"category"`              // 所属分类
	BrandID          uint64        `gorm:"index;not null" json:"brand_id"`                     // 所属品牌ID
	Brand            Brand         `gorm:"foreignKey:BrandID" json:"brand"`                    // 所属品牌
	Status           ProductStatus `gorm:"type:tinyint;not null" json:"status"`                // 商品状态
	MainImageURL     string        `gorm:"type:varchar(255)" json:"main_image_url"`            // 商品主图URL
	GalleryImageURLs string        `gorm:"type:text" json:"gallery_image_urls"`                // 商品画廊图片URL列表 (JSON字符串存储)
	MetaTitle        string        `gorm:"type:varchar(255)" json:"meta_title"`                // SEO标题
	MetaDescription  string        `gorm:"type:varchar(500)" json:"meta_description"`          // SEO描述
	MetaKeywords     string        `gorm:"type:varchar(255)" json:"meta_keywords"`             // SEO关键词 (逗号分隔)
	URLSlug          string        `gorm:"type:varchar(255);uniqueIndex" json:"url_slug"`      // URL Slug (例如: "iphone-15-pro-max")
	Weight           float64       `gorm:"type:decimal(10,2)" json:"weight"`                   // 商品重量 (kg)
	CreatedAt        time.Time     `gorm:"autoCreateTime" json:"created_at"`                   // 创建时间
	UpdatedAt        time.Time     `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt        *time.Time    `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间

	SKUs       []SKU              `gorm:"foreignKey:ProductID" json:"skus"` // 关联的SKU列表
	Attributes []ProductAttribute `gorm:"foreignKey:ProductID" json:"attributes"` // 商品属性
}

// SKU 代表一个库存量单位 (SKU - Stock Keeping Unit)。
// SKU是实际销售的单位，例如 "iPhone 15 Pro Max 256GB 蓝色"。
type SKU struct {
	ID            uint64     `gorm:"primarykey" json:"id"`                               // SKU ID
	ProductID     uint64     `gorm:"index;not null" json:"product_id"`                   // 所属SPU ID
	SkuNo         string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"sku_no"` // SKU编码, 例如 "SKU123456-BLUE-256G"
	Name          string     `gorm:"type:varchar(255);not null" json:"name"`             // SKU名称 (通常由SPU名称 + 规格值构成)
	Price         int64      `gorm:"type:bigint;not null" json:"price"`                  // 价格 (单位: 分，避免浮点数精度问题)
	StockQuantity int32      `gorm:"type:int;not null" json:"stock_quantity"`            // 库存数量
	ImageURL      string     `gorm:"type:varchar(255)" json:"image_url"`                 // SKU特定图片URL (例如不同颜色的商品图)
	SpecValues    string     `gorm:"type:text" json:"spec_values"`                       // 规格值列表 (JSON字符串存储)
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 创建时间
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}

// Category 商品分类信息。
type Category struct {
	ID        uint64     `gorm:"primarykey" json:"id"`                               // 分类ID
	Name      string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"` // 分类名称
	ParentID  uint64     `gorm:"index;not null;default:0" json:"parent_id"`          // 父分类ID, 0表示顶级分类
	Level     int32      `gorm:"type:tinyint;not null;default:1" json:"level"`       // 分类层级
	IconURL   string     `gorm:"type:varchar(255)" json:"icon_url"`                  // 分类图标URL
	SortOrder int32      `gorm:"type:int;not null;default:0" json:"sort_order"`      // 排序值
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 创建时间
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间

	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"` // 子分类
}

// Brand 商品品牌信息。

// ProductAttribute 商品的通用属性。
// 例如: {"材质": "纯棉"}, {"产地": "中国"}
type ProductAttribute struct {
	ID        uint64     `gorm:"primarykey" json:"id"`
	ProductID uint64     `gorm:"index;not null" json:"product_id"` // 所属SPU ID
	Key       string     `gorm:"type:varchar(100);not null" json:"key"`
	Value     string     `gorm:"type:varchar(255);not null" json:"value"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// SpecValue SKU的规格值。
// 例如: {"颜色": "深空灰"}, {"存储": "256GB"}
type SpecValue struct {
	Key   string `json:"key"`   // 规格名, 例如 "颜色"
	Value string `json:"value"` // 规格值, 例如 "深空灰"
}
