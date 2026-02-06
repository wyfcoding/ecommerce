package domain

import (
	"fmt"  // 导入格式化库，用于错误信息。
	"time" // 导入时间库。
)

// ProductStatus 定义了商品的生命周期状态。
type ProductStatus int

const (
	ProductStatusDraft     ProductStatus = 1 // 草稿：商品信息已录入，但尚未发布。
	ProductStatusPublished ProductStatus = 2 // 已发布：商品已上架，可供用户购买。
	ProductStatusOffline   ProductStatus = 3 // 已下架：商品已从销售渠道移除。
	ProductStatusDeleted   ProductStatus = 4 // 已删除：商品已被逻辑删除。
)

// ProductType 定义了商品类型。
type ProductType int

const (
	ProductTypePhysical ProductType = 1 // 实物商品 (需要物流)
	ProductTypeDigital  ProductType = 2 // 虚拟商品 (充值卡、会员)
	ProductTypeAsset    ProductType = 3 // 金融资产 (股票、债券)
)

// Product 实体是商品模块的聚合根。
type Product struct {
	ID          uint          `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CategoryID  uint          `json:"category_id"`
	BrandID     uint          `json:"brand_id"`
	Status      ProductStatus `json:"status"`
	Type        ProductType   `json:"type"`
	MainImage   string        `json:"main_image"`
	Images      []string      `json:"images"`
	Price       int64         `json:"price"`
	Stock       int32         `json:"stock"`
	Sales       int32         `json:"sales"`
	SKUs        []*SKU        `json:"skus"`
}

// SKU 实体代表商品的最小库存单元。
type SKU struct {
	ID        uint              `json:"id"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	ProductID uint              `json:"product_id"`
	Name      string            `json:"name"`
	Price     int64             `json:"price"`
	Stock     int32             `json:"stock"`
	Sales     int32             `json:"sales"`
	Image     string            `json:"image"`
	Specs     map[string]string `json:"specs"`
}

// Category 实体代表商品分类。
type Category struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ParentID  uint      `json:"parent_id"`
	Sort      int       `json:"sort"`
	Status    int       `json:"status"`
}

// Brand 实体代表商品品牌。
type Brand struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Logo      string    `json:"logo"`
	Status    int       `json:"status"`
}

// NewProduct 是一个工厂方法，用于创建并返回一个新的 Product 实体实例。
func NewProduct(name, description string, categoryID, brandID uint, pType ProductType, price int64, stock int32) (*Product, error) {
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
		Type:        pType,
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
