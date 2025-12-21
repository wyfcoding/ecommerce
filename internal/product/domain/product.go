package domain

import (
	"fmt" // 导入格式化库，用于错误信息。

	"gorm.io/gorm" // 导入GORM库。
)

// ProductStatus 定义了商品的生命周期状态。
type ProductStatus int

const (
	ProductStatusDraft     ProductStatus = 1 // 草稿：商品信息已录入，但尚未发布。
	ProductStatusPublished ProductStatus = 2 // 已发布：商品已上架，可供用户购买。
	ProductStatusOffline   ProductStatus = 3 // 已下架：商品已从销售渠道移除。
	ProductStatusDeleted   ProductStatus = 4 // 已删除：商品已被逻辑删除。
)

// Product 实体是商品模块的聚合根。
// 它包含了商品的基本信息、分类、品牌、价格、库存、销量和关联的SKU列表。
type Product struct {
	gorm.Model                // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name        string        `gorm:"column:name;type:varchar(255);not null" json:"name"`     // 商品名称，不允许为空。
	Description string        `gorm:"column:description;type:text" json:"description"`        // 商品描述。
	CategoryID  uint          `gorm:"column:category_id;index;not null" json:"category_id"`   // 所属分类ID，索引字段，不允许为空。
	BrandID     uint          `gorm:"column:brand_id;index;not null" json:"brand_id"`         // 所属品牌ID，索引字段，不允许为空。
	Status      ProductStatus `gorm:"column:status;type:tinyint;default:1" json:"status"`     // 商品状态，默认为草稿。
	MainImage   string        `gorm:"column:main_image;type:varchar(1024)" json:"main_image"` // 商品主图URL。
	Images      []string      `gorm:"type:json;serializer:json" json:"images"`                // 商品图片列表（存储为JSON字符串）。
	Price       int64         `gorm:"column:price;type:bigint;not null" json:"price"`         // 商品默认价格（单位：分），不允许为空。
	Stock       int32         `gorm:"column:stock;type:int;default:0" json:"stock"`           // 商品总库存。
	Sales       int32         `gorm:"column:sales;type:int;default:0" json:"sales"`           // 商品总销量。
	SKUs        []*SKU        `gorm:"foreignKey:ProductID" json:"skus"`                       // 关联的SKU列表，一对多关系。
}

// SKU 实体代表商品的库存量单位（Stock Keeping Unit）。
// 它是商品的具体变体，例如不同颜色、尺寸等，每个SKU有独立的库存和价格。
type SKU struct {
	gorm.Model                   // 嵌入gorm.Model。
	ProductID  uint              `gorm:"column:product_id;index;not null" json:"product_id"` // 所属商品ID，索引字段，不允许为空。
	Name       string            `gorm:"column:name;type:varchar(255);not null" json:"name"` // SKU名称（例如，“红色，L码”）。
	Price      int64             `gorm:"column:price;type:bigint;not null" json:"price"`     // SKU价格（单位：分）。
	Stock      int32             `gorm:"column:stock;type:int;default:0" json:"stock"`       // SKU库存。
	Sales      int32             `gorm:"column:sales;type:int;default:0" json:"sales"`       // SKU销量。
	Image      string            `gorm:"column:image;type:varchar(1024)" json:"image"`       // SKU图片URL。
	Specs      map[string]string `gorm:"type:json;serializer:json" json:"specs"`             // SKU规格参数（例如，{"color": "red", "size": "L"}，存储为JSON字符串）。
}

// Category 实体代表商品分类。
// 用于组织和管理商品，支持树形结构。
type Category struct {
	gorm.Model        // 嵌入gorm.Model。
	Name       string `gorm:"column:name;type:varchar(255);not null" json:"name"` // 分类名称，不允许为空。
	ParentID   uint   `gorm:"column:parent_id;index;default:0" json:"parent_id"`  // 父分类ID，0表示顶级分类，索引字段。
	Sort       int    `gorm:"column:sort;type:int;default:0" json:"sort"`         // 排序值。
	Status     int    `gorm:"column:status;type:tinyint;default:1" json:"status"` // 状态，1:正常, 2:禁用。
}

// Brand 实体代表商品品牌。
type Brand struct {
	gorm.Model        // 嵌入gorm.Model。
	Name       string `gorm:"column:name;type:varchar(255);not null" json:"name"` // 品牌名称，不允许为空。
	Logo       string `gorm:"column:logo;type:varchar(1024)" json:"logo"`         // 品牌Logo图片URL。
	Status     int    `gorm:"column:status;type:tinyint;default:1" json:"status"` // 状态，1:正常, 2:禁用。
}

// NewProduct 是一个工厂方法，用于创建并返回一个新的 Product 实体实例。
func NewProduct(name, description string, categoryID, brandID uint, price int64, stock int32) (*Product, error) {
	if name == "" {
		return nil, fmt.Errorf("product name cannot be empty")
	}
	if price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if stock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}

	return &Product{
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		BrandID:     brandID,
		Status:      ProductStatusDraft, // 新商品默认为草稿状态。
		Price:       price,
		Stock:       stock,
		Sales:       0,
		SKUs:        []*SKU{}, // 初始化SKU列表。
	}, nil
}

// Publish 将商品状态变更为“已发布”。
func (p *Product) Publish() error {
	if p.Status != ProductStatusDraft {
		return fmt.Errorf("only products in draft status can be published")
	}
	p.Status = ProductStatusPublished
	return nil
}

// Offline 将商品状态变更为“已下架”。
func (p *Product) Offline() error {
	if p.Status != ProductStatusPublished {
		return fmt.Errorf("only published products can be taken offline")
	}
	p.Status = ProductStatusOffline
	return nil
}

// Delete 将商品状态变更为“已删除”（逻辑删除）。
func (p *Product) Delete() error {
	p.Status = ProductStatusDeleted
	return nil
}

// UpdateStock 更新商品的库存数量。
func (p *Product) UpdateStock(stock int32) error {
	if stock < 0 {
		return fmt.Errorf("stock cannot be negative")
	}
	p.Stock = stock
	return nil
}

// IncreaseSales 增加商品的销量。
func (p *Product) IncreaseSales(quantity int32) error {
	if quantity <= 0 {
		return fmt.Errorf("increased sales must be greater than 0")
	}
	p.Sales += quantity
	return nil
}

// AddSKU 为商品添加一个SKU。
func (p *Product) AddSKU(sku *SKU) error {
	// 如果商品ID已存在，则将SKU的ProductID设置为商品ID。
	if p.ID != 0 {
		sku.ProductID = p.ID
	}
	p.SKUs = append(p.SKUs, sku) // 将SKU添加到商品关联的SKU列表中。
	return nil
}

// RemoveSKU 移除商品的一个SKU。
func (p *Product) RemoveSKU(skuID uint) error {
	for i, sku := range p.SKUs {
		if sku.ID == skuID {
			p.SKUs = append(p.SKUs[:i], p.SKUs[i+1:]...) // 从列表中移除SKU。
			return nil
		}
	}
	return fmt.Errorf("SKU not found")
}

// NewSKU 是一个工厂方法，用于创建并返回一个新的 SKU 实体实例。
func NewSKU(productID uint, name string, price int64, stock int32, image string, specs map[string]string) (*SKU, error) {
	if name == "" {
		return nil, fmt.Errorf("SKU name cannot be empty")
	}
	if price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if stock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}

	return &SKU{
		ProductID: productID,
		Name:      name,
		Price:     price,
		Stock:     stock,
		Sales:     0, // 初始销量为0。
		Image:     image,
		Specs:     specs,
	}, nil
}

// NewCategory 是一个工厂方法，用于创建并返回一个新的 Category 实体实例。
func NewCategory(name string, parentID uint) (*Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name cannot be empty")
	}

	return &Category{
		Name:     name,
		ParentID: parentID,
		Sort:     0, // 默认排序值为0。
		Status:   1, // 默认状态为正常。
	}, nil
}

// NewBrand 是一个工厂方法，用于创建并返回一个新的 Brand 实体实例。
func NewBrand(name, logo string) (*Brand, error) {
	if name == "" {
		return nil, fmt.Errorf("brand name cannot be empty")
	}

	return &Brand{
		Name:   name,
		Logo:   logo,
		Status: 1, // 默认状态为正常。
	}, nil
}
