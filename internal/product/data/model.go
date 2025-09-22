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
	gorm.Model
	ParentID  uint64 `gorm:"index;default:0;comment:父分类ID" json:"parentId"`
	Name      string `gorm:"uniqueIndex;not null;type:varchar(64);comment:分类名称" json:"name"`
	Level     uint32 `gorm:"not null;default:1;type:tinyint;comment:分类级别" json:"level"`
	Icon      string `gorm:"type:varchar(255);comment:分类图标" json:"icon"`
	SortOrder uint32 `gorm:"default:0;type:int;comment:排序" json:"sortOrder"`
	IsVisible bool   `gorm:"default:true;type:boolean;comment:是否可见" json:"isVisible"`
	Spus      []Spu     `gorm:"foreignKey:CategoryID" json:"spus"` // 关联商品
}

// Brand is a GORM model for brands.
type Brand struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;type:varchar(64);comment:品牌名称"`
	Logo        string `gorm:"type:varchar(255);comment:品牌Logo"`
	Description string `gorm:"type:varchar(512);comment:品牌描述"`
	Website     string `gorm:"type:varchar(255);comment:品牌官网"`
	SortOrder   uint32 `gorm:"default:0;type:int;comment:排序"`
	IsVisible   bool   `gorm:"default:true;type:boolean;comment:是否可见"`
}

// Spu is a GORM model for products (SPU).
type Spu struct {
	gorm.Model
	CategoryID     uint64        `gorm:"index;not null;comment:分类ID" json:"categoryId"`
	BrandID        uint64        `gorm:"index;not null;comment:品牌ID" json:"brandId"`
	Title          string        `gorm:"not null;type:varchar(255);comment:SPU标题" json:"title"`
	SubTitle       string        `gorm:"type:varchar(255);comment:SPU副标题" json:"subTitle"`
	MainImage      string        `gorm:"type:varchar(255);comment:主图URL" json:"mainImage"`
	GalleryImages  GalleryImages `gorm:"type:text;comment:画廊图片URL，JSON数组" json:"galleryImages"`
	DetailHTML     string        `gorm:"type:longtext;comment:商品详情HTML" json:"detailHtml"`
	Status         int32         `gorm:"not null;default:1;comment:状态 1-上架 2-下架 3-删除" json:"status"`
	Skus           []Sku         `gorm:"foreignKey:SpuID" json:"skus"` // 关联商品
}

// Sku 库存量单位 (Stock Keeping Unit)
type Sku struct {
	gorm.Model
	SkuID         uint64 `gorm:"uniqueIndex;not null;comment:SKU ID"`
	SpuID         uint64 `gorm:"index;not null;comment:所属SPU ID"`
	Title         string `gorm:"not null;type:varchar(255);comment:SKU标题"`
	Price         uint64 `gorm:"not null;comment:销售价格, 单位分"`
	OriginalPrice uint64 `gorm:"comment:原价, 单位分"`
	Stock         uint32 `gorm:"not null;default:0;comment:库存"`
	Image         string `gorm:"type:varchar(255);comment:SKU图片URL"`
	Specs         Specs  `gorm:"type:text;comment:规格属性"`
	Status        int32  `gorm:"not null;default:1;comment:状态 1-在售 2-下架 3-删除"`
}

// MongoProduct represents a product document for MongoDB indexing.
type MongoProduct struct {
	ID          string    `bson:"_id,omitempty"` // MongoDB primary key
	SpuID       uint64    `bson:"spu_id"`
	CategoryID  uint64    `bson:"category_id"`
	BrandID     uint64    `bson:"brand_id"`
	Title       string    `bson:"title"`
	SubTitle    string    `bson:"sub_title"`
	MainImage   string    `bson:"main_image"`
	DetailHTML  string    `bson:"detail_html"`
	Status      int32     `bson:"status"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	// Add other fields relevant for indexing/search in MongoDB
}

// Review is a GORM model for product reviews.
type Review struct {
	gorm.Model
	SpuID     uint64 `gorm:"index;not null;comment:商品SPU ID"`
	UserID    uint64 `gorm:"index;not null;comment:用户ID"`
	Rating    uint32 `gorm:"not null;type:tinyint;comment:评分 (1-5星)"`
	Comment   string `gorm:"type:text;comment:评论内容"`
	Images    string `gorm:"type:text;comment:评论图片URL，逗号分隔"` // Stored as comma-separated string
}

// TableName 指定 GORM 使用的表名
func (Category) TableName() string { return "categories" }
func (Spu) TableName() string      { return "spus" }
func (Sku) TableName() string      { return "skus" }
