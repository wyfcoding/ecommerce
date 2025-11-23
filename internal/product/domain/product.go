package domain

import (
	"fmt"

	"gorm.io/gorm"
)

// ProductStatus 商品状态
type ProductStatus int

const (
	ProductStatusDraft     ProductStatus = 1 // 草稿
	ProductStatusPublished ProductStatus = 2 // 已发布
	ProductStatusOffline   ProductStatus = 3 // 已下架
	ProductStatusDeleted   ProductStatus = 4 // 已删除
)

// Product 商品聚合根
// 包含商品的基本信息、状态、SKU列表等
type Product struct {
	gorm.Model
	Name        string        `gorm:"column:name;type:varchar(255);not null" json:"name"`     // 商品名称
	Description string        `gorm:"column:description;type:text" json:"description"`        // 商品描述
	CategoryID  uint          `gorm:"column:category_id;index;not null" json:"category_id"`   // 分类ID
	BrandID     uint          `gorm:"column:brand_id;index;not null" json:"brand_id"`         // 品牌ID
	Status      ProductStatus `gorm:"column:status;type:tinyint;default:1" json:"status"`     // 商品状态
	MainImage   string        `gorm:"column:main_image;type:varchar(1024)" json:"main_image"` // 主图URL
	Images      []string      `gorm:"type:json;serializer:json" json:"images"`                // 图片列表 (JSON存储)
	Price       int64         `gorm:"column:price;type:bigint;not null" json:"price"`         // 价格 (分)
	Stock       int32         `gorm:"column:stock;type:int;default:0" json:"stock"`           // 库存
	Sales       int32         `gorm:"column:sales;type:int;default:0" json:"sales"`           // 销量
	SKUs        []*SKU        `gorm:"foreignKey:ProductID" json:"skus"`                       // SKU列表
}

// SKU SKU实体
// 库存量单位，具体的销售商品
type SKU struct {
	gorm.Model
	ProductID uint              `gorm:"column:product_id;index;not null" json:"product_id"` // 所属商品ID
	Name      string            `gorm:"column:name;type:varchar(255);not null" json:"name"` // SKU名称
	Price     int64             `gorm:"column:price;type:bigint;not null" json:"price"`     // 价格 (分)
	Stock     int32             `gorm:"column:stock;type:int;default:0" json:"stock"`       // 库存
	Sales     int32             `gorm:"column:sales;type:int;default:0" json:"sales"`       // 销量
	Image     string            `gorm:"column:image;type:varchar(1024)" json:"image"`       // SKU图片
	Specs     map[string]string `gorm:"type:json;serializer:json" json:"specs"`             // 规格参数 (JSON存储)
}

// Category 分类实体
// 商品分类树形结构
type Category struct {
	gorm.Model
	Name     string `gorm:"column:name;type:varchar(255);not null" json:"name"` // 分类名称
	ParentID uint   `gorm:"column:parent_id;index;default:0" json:"parent_id"`  // 父分类ID
	Sort     int    `gorm:"column:sort;type:int;default:0" json:"sort"`         // 排序
	Status   int    `gorm:"column:status;type:tinyint;default:1" json:"status"` // 状态 1:正常 2:禁用
}

// Brand 品牌实体
// 商品品牌信息
type Brand struct {
	gorm.Model
	Name   string `gorm:"column:name;type:varchar(255);not null" json:"name"` // 品牌名称
	Logo   string `gorm:"column:logo;type:varchar(1024)" json:"logo"`         // 品牌Logo
	Status int    `gorm:"column:status;type:tinyint;default:1" json:"status"` // 状态 1:正常 2:禁用
}

// NewProduct 创建商品工厂方法
func NewProduct(name, description string, categoryID, brandID uint, price int64, stock int32) (*Product, error) {
	if name == "" {
		return nil, fmt.Errorf("商品名称不能为空")
	}
	if price <= 0 {
		return nil, fmt.Errorf("价格必须大于0")
	}
	if stock < 0 {
		return nil, fmt.Errorf("库存不能为负数")
	}

	return &Product{
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		BrandID:     brandID,
		Status:      ProductStatusDraft,
		Price:       price,
		Stock:       stock,
		Sales:       0,
		SKUs:        []*SKU{},
	}, nil
}

// Publish 发布商品
func (p *Product) Publish() error {
	if p.Status != ProductStatusDraft {
		return fmt.Errorf("只有草稿状态的商品可以发布")
	}
	p.Status = ProductStatusPublished
	return nil
}

// Offline 下架商品
func (p *Product) Offline() error {
	if p.Status != ProductStatusPublished {
		return fmt.Errorf("只有已发布的商品可以下架")
	}
	p.Status = ProductStatusOffline
	return nil
}

// Delete 删除商品
func (p *Product) Delete() error {
	p.Status = ProductStatusDeleted
	return nil
}

// UpdateStock 更新库存
func (p *Product) UpdateStock(stock int32) error {
	if stock < 0 {
		return fmt.Errorf("库存不能为负数")
	}
	p.Stock = stock
	return nil
}

// IncreaseSales 增加销量
func (p *Product) IncreaseSales(quantity int32) error {
	if quantity <= 0 {
		return fmt.Errorf("增加的销量必须大于0")
	}
	p.Sales += quantity
	return nil
}

// AddSKU 添加SKU
func (p *Product) AddSKU(sku *SKU) error {
	if p.ID != 0 {
		sku.ProductID = p.ID
	}
	p.SKUs = append(p.SKUs, sku)
	return nil
}

// RemoveSKU 移除SKU
func (p *Product) RemoveSKU(skuID uint) error {
	for i, sku := range p.SKUs {
		if sku.ID == skuID {
			p.SKUs = append(p.SKUs[:i], p.SKUs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("SKU未找到")
}

// NewSKU 创建SKU工厂方法
func NewSKU(productID uint, name string, price int64, stock int32, image string, specs map[string]string) (*SKU, error) {
	if name == "" {
		return nil, fmt.Errorf("SKU名称不能为空")
	}
	if price <= 0 {
		return nil, fmt.Errorf("价格必须大于0")
	}
	if stock < 0 {
		return nil, fmt.Errorf("库存不能为负数")
	}

	return &SKU{
		ProductID: productID,
		Name:      name,
		Price:     price,
		Stock:     stock,
		Sales:     0,
		Image:     image,
		Specs:     specs,
	}, nil
}

// NewCategory 创建分类工厂方法
func NewCategory(name string, parentID uint) (*Category, error) {
	if name == "" {
		return nil, fmt.Errorf("分类名称不能为空")
	}

	return &Category{
		Name:     name,
		ParentID: parentID,
		Sort:     0,
		Status:   1,
	}, nil
}

// NewBrand 创建品牌工厂方法
func NewBrand(name, logo string) (*Brand, error) {
	if name == "" {
		return nil, fmt.Errorf("品牌名称不能为空")
	}

	return &Brand{
		Name:   name,
		Logo:   logo,
		Status: 1,
	}, nil
}
