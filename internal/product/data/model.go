package data

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// GalleryImages 是一个自定义类型，用于将字符串数组 json 化存入数据库
type GalleryImages []string

func (g GalleryImages) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GalleryImages) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), g)
}

// Specs 是一个自定义类型，用于将 map[string]string json 化存入数据库
type Specs map[string]string

func (s Specs) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *Specs) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), s)
}

// Category is a GORM model for categories.
type Category struct {
	ID        uint64 `gorm:"primarykey"`
	ParentID  uint64 `gorm:"index;default:0;comment:父分类ID"`
	Name      string `gorm:"type:varchar(64);not null;uniqueIndex;comment:分类名称"`
	Level     uint32 `gorm:"type:tinyint;not null;default:1;comment:分类级别"`
	Icon      string `gorm:"type:varchar(255);comment:分类图标"`
	SortOrder uint32 `gorm:"type:int;default:0;comment:排序"`
	IsVisible bool   `gorm:"type:boolean;default:true;comment:是否可见"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Products  []Product `gorm:"foreignKey:CategoryID"` // 关联商品
}

// Brand is a GORM model for brands.
type Brand struct {
	ID          uint64 `gorm:"primarykey"`
	Name        string `gorm:"type:varchar(64);not null;uniqueIndex;comment:品牌名称"`
	Logo        string `gorm:"type:varchar(255);comment:品牌Logo"`
	Description string `gorm:"type:varchar(512);comment:品牌描述"`
	Website     string `gorm:"type:varchar(255);comment:品牌官网"`
	SortOrder   uint32 `gorm:"type:int;default:0;comment:排序"`
	IsVisible   bool   `gorm:"type:boolean;default:true;comment:是否可见"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Product is a GORM model for products (SPU).
type Product struct {
	SpuID         uint64 `gorm:"primarykey;comment:SPU ID"`
	CategoryID    uint64 `gorm:"index;not null;comment:分类ID"`
	BrandID       uint64 `gorm:"index;comment:品牌ID"`
	Title         string `gorm:"type:varchar(255);not null;comment:商品标题"`
	SubTitle      string `gorm:"type:varchar(255);comment:商品副标题"`
	MainImage     string `gorm:"type:varchar(255);comment:商品主图"`
	GalleryImages string `gorm:"type:text;comment:商品画廊图片，逗号分隔"` // 存储为逗号分隔的字符串
	DetailHTML    string `gorm:"type:longtext;comment:商品详情HTML"`
	Status        int32  `gorm:"type:tinyint;not null;default:1;comment:商品状态：1-上架，0-下架"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Skus          []Sku `gorm:"foreignKey:SpuID"` // 关联SKU
}

// Sku 库存量单位 (Stock Keeping Unit)
type Sku struct {
	gorm.Model
	SkuID         uint64 `gorm:"uniqueIndex;not null;comment:SKU ID"`
	SpuID         uint64 `gorm:"index;not null;comment:所属SPU ID"`
	Title         string `gorm:"type:varchar(255);not null;comment:SKU标题"`
	Price         uint64 `gorm:"not null;comment:销售价格, 单位分"`
	OriginalPrice uint64 `gorm:"comment:原价, 单位分"`
	Stock         uint32 `gorm:"not null;default:0;comment:库存"`
	Image         string `gorm:"type:varchar(255);comment:SKU图片URL"`
	Specs         Specs  `gorm:"type:text;comment:规格属性"`
	Status        int32  `gorm:"not null;default:1;comment:状态 1-在售 2-下架 3-删除"`
}

// Review is a GORM model for product reviews.
type Review struct {
	ID        uint64 `gorm:"primarykey"`
	SpuID     uint64 `gorm:"index;not null;comment:商品SPU ID"`
	UserID    uint64 `gorm:"index;not null;comment:用户ID"`
	Rating    uint32 `gorm:"type:tinyint;not null;comment:评分 (1-5星)"`
	Comment   string `gorm:"type:text;comment:评论内容"`
	Images    string `gorm:"type:text;comment:评论图片URL，逗号分隔"` // Stored as comma-separated string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName 指定 GORM 使用的表名
func (Category) TableName() string { return "categories" }
func (Spu) TableName() string      { return "spus" }
func (Sku) TableName() string      { return "skus" }
