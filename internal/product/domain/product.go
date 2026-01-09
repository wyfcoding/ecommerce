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
type Product struct {
	gorm.Model                // 嵌入 gorm.Model，包含 ID, CreatedAt, UpdatedAt, DeletedAt 等通用字段
	Name        string        `gorm:"column:name;type:varchar(255);not null;comment:商品名称" json:"name"`
	Description string        `gorm:"column:description;type:text;comment:商品详细描述" json:"description"`
	CategoryID  uint          `gorm:"column:category_id;index;not null;comment:所属分类ID" json:"category_id"`
	BrandID     uint          `gorm:"column:brand_id;index;not null;comment:所属品牌ID" json:"brand_id"`
	Status      ProductStatus `gorm:"column:status;type:tinyint;default:1;comment:商品状态(1:草稿, 2:已发布, 3:已下架, 4:已删除)" json:"status"`
	MainImage   string        `gorm:"column:main_image;type:varchar(1024);comment:商品主图URL" json:"main_image"`
	Images      []string      `gorm:"type:json;serializer:json;comment:商品图集JSON" json:"images"`
	Price       int64         `gorm:"column:price;type:bigint;not null;comment:默认售价(单位:分)" json:"price"`
	Stock       int32         `gorm:"column:stock;type:int;default:0;comment:总库存量" json:"stock"`
	Sales       int32         `gorm:"column:sales;type:int;default:0;comment:历史总销量" json:"sales"`
	SKUs        []*SKU        `gorm:"foreignKey:ProductID;comment:关联SKU集合" json:"skus"`
}

// SKU 实体代表商品的最小库存单元。
type SKU struct {
	gorm.Model
	ProductID uint              `gorm:"column:product_id;index;not null;comment:关联商品ID" json:"product_id"`
	Name      string            `gorm:"column:name;type:varchar(255);not null;comment:规格名称" json:"name"`
	Price     int64             `gorm:"column:price;type:bigint;not null;comment:规格售价(单位:分)" json:"price"`
	Stock     int32             `gorm:"column:stock;type:int;default:0;comment:规格库存" json:"stock"`
	Sales     int32             `gorm:"column:sales;type:int;default:0;comment:规格销量" json:"sales"`
	Image     string            `gorm:"column:image;type:varchar(1024);comment:规格图片URL" json:"image"`
	Specs     map[string]string `gorm:"type:json;serializer:json;comment:规格参数JSON" json:"specs"`
}

// Category 实体代表商品分类。
type Category struct {
	gorm.Model
	Name     string `gorm:"column:name;type:varchar(255);not null;comment:分类名称" json:"name"`
	ParentID uint   `gorm:"column:parent_id;index;default:0;comment:父分类ID(0为根节点)" json:"parent_id"`
	Sort     int    `gorm:"column:sort;type:int;default:0;comment:排序权重" json:"sort"`
	Status   int    `gorm:"column:status;type:tinyint;default:1;comment:分类状态(1:启用, 2:禁用)" json:"status"`
}

// Brand 实体代表商品品牌。
type Brand struct {
	gorm.Model
	Name   string `gorm:"column:name;type:varchar(255);not null;comment:品牌名称" json:"name"`
	Logo   string `gorm:"column:logo;type:varchar(1024);comment:品牌Logo地址" json:"logo"`
	Status int    `gorm:"column:status;type:tinyint;default:1;comment:品牌状态(1:启用, 2:禁用)" json:"status"`
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
