package mysql

import (
	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"gorm.io/gorm"
)

// ProductModel 商品写模型（持久化专用）。
type ProductModel struct {
	gorm.Model
	Name        string               `gorm:"column:name;type:varchar(255);not null;comment:商品名称"`
	Description string               `gorm:"column:description;type:text;comment:商品详细描述"`
	CategoryID  uint                 `gorm:"column:category_id;index;not null;comment:所属分类ID"`
	BrandID     uint                 `gorm:"column:brand_id;index;not null;comment:所属品牌ID"`
	Status      domain.ProductStatus `gorm:"column:status;type:tinyint;default:1;comment:商品状态(1:草稿, 2:已发布, 3:已下架, 4:已删除)"`
	Type        domain.ProductType   `gorm:"column:type;type:tinyint;default:1;comment:商品类型(1:实物, 2:虚拟, 3:资产)"`
	MainImage   string               `gorm:"column:main_image;type:varchar(1024);comment:商品主图URL"`
	Images      []string             `gorm:"type:json;serializer:json;comment:商品图集JSON"`
	Price       int64                `gorm:"column:price;type:bigint;not null;comment:默认售价(单位:分)"`
	Stock       int32                `gorm:"column:stock;type:int;default:0;comment:总库存量"`
	Sales       int32                `gorm:"column:sales;type:int;default:0;comment:历史总销量"`
	SKUs        []*SKUModel          `gorm:"foreignKey:ProductID;comment:关联SKU集合"`
}

func (ProductModel) TableName() string {
	return "products"
}

// SKUModel 商品 SKU 写模型。
type SKUModel struct {
	gorm.Model
	ProductID uint              `gorm:"column:product_id;index;not null;comment:关联商品ID"`
	Name      string            `gorm:"column:name;type:varchar(255);not null;comment:规格名称"`
	Price     int64             `gorm:"column:price;type:bigint;not null;comment:规格售价(单位:分)"`
	Stock     int32             `gorm:"column:stock;type:int;default:0;comment:规格库存"`
	Sales     int32             `gorm:"column:sales;type:int;default:0;comment:规格销量"`
	Image     string            `gorm:"column:image;type:varchar(1024);comment:规格图片URL"`
	Specs     map[string]string `gorm:"type:json;serializer:json;comment:规格参数JSON"`
}

func (SKUModel) TableName() string {
	return "product_skus"
}

// CategoryModel 商品分类写模型。
type CategoryModel struct {
	gorm.Model
	Name     string `gorm:"column:name;type:varchar(255);not null;comment:分类名称"`
	ParentID uint   `gorm:"column:parent_id;index;default:0;comment:父分类ID(0为根节点)"`
	Sort     int    `gorm:"column:sort;type:int;default:0;comment:排序权重"`
	Status   int    `gorm:"column:status;type:tinyint;default:1;comment:分类状态(1:启用, 2:禁用)"`
}

func (CategoryModel) TableName() string {
	return "product_categories"
}

// BrandModel 商品品牌写模型。
type BrandModel struct {
	gorm.Model
	Name   string `gorm:"column:name;type:varchar(255);not null;comment:品牌名称"`
	Logo   string `gorm:"column:logo;type:varchar(1024);comment:品牌Logo地址"`
	Status int    `gorm:"column:status;type:tinyint;default:1;comment:品牌状态(1:启用, 2:禁用)"`
}

func (BrandModel) TableName() string {
	return "product_brands"
}

func toProductModel(product *domain.Product) *ProductModel {
	if product == nil {
		return nil
	}
	model := &ProductModel{
		Model: gorm.Model{
			ID:        product.ID,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
		},
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
		BrandID:     product.BrandID,
		Status:      product.Status,
		Type:        product.Type,
		MainImage:   product.MainImage,
		Images:      product.Images,
		Price:       product.Price,
		Stock:       product.Stock,
		Sales:       product.Sales,
	}
	if len(product.SKUs) > 0 {
		model.SKUs = make([]*SKUModel, 0, len(product.SKUs))
		for _, sku := range product.SKUs {
			model.SKUs = append(model.SKUs, toSKUModel(sku))
		}
	}
	return model
}

func toDomainProduct(model *ProductModel) *domain.Product {
	if model == nil {
		return nil
	}
	product := &domain.Product{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Description: model.Description,
		CategoryID:  model.CategoryID,
		BrandID:     model.BrandID,
		Status:      model.Status,
		Type:        model.Type,
		MainImage:   model.MainImage,
		Images:      model.Images,
		Price:       model.Price,
		Stock:       model.Stock,
		Sales:       model.Sales,
	}
	if len(model.SKUs) > 0 {
		skus := make([]*domain.SKU, 0, len(model.SKUs))
		for _, sku := range model.SKUs {
			skus = append(skus, toDomainSKU(sku))
		}
		product.SKUs = skus
	}
	return product
}

func toSKUModel(sku *domain.SKU) *SKUModel {
	if sku == nil {
		return nil
	}
	return &SKUModel{
		Model: gorm.Model{
			ID:        sku.ID,
			CreatedAt: sku.CreatedAt,
			UpdatedAt: sku.UpdatedAt,
		},
		ProductID: sku.ProductID,
		Name:      sku.Name,
		Price:     sku.Price,
		Stock:     sku.Stock,
		Sales:     sku.Sales,
		Image:     sku.Image,
		Specs:     sku.Specs,
	}
}

func toDomainSKU(model *SKUModel) *domain.SKU {
	if model == nil {
		return nil
	}
	return &domain.SKU{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		ProductID: model.ProductID,
		Name:      model.Name,
		Price:     model.Price,
		Stock:     model.Stock,
		Sales:     model.Sales,
		Image:     model.Image,
		Specs:     model.Specs,
	}
}

func toCategoryModel(category *domain.Category) *CategoryModel {
	if category == nil {
		return nil
	}
	return &CategoryModel{
		Model: gorm.Model{
			ID:        category.ID,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		},
		Name:     category.Name,
		ParentID: category.ParentID,
		Sort:     category.Sort,
		Status:   category.Status,
	}
}

func toDomainCategory(model *CategoryModel) *domain.Category {
	if model == nil {
		return nil
	}
	return &domain.Category{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		ParentID:  model.ParentID,
		Sort:      model.Sort,
		Status:    model.Status,
	}
}

func toBrandModel(brand *domain.Brand) *BrandModel {
	if brand == nil {
		return nil
	}
	return &BrandModel{
		Model: gorm.Model{
			ID:        brand.ID,
			CreatedAt: brand.CreatedAt,
			UpdatedAt: brand.UpdatedAt,
		},
		Name:   brand.Name,
		Logo:   brand.Logo,
		Status: brand.Status,
	}
}

func toDomainBrand(model *BrandModel) *domain.Brand {
	if model == nil {
		return nil
	}
	return &domain.Brand{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		Logo:      model.Logo,
		Status:    model.Status,
	}
}
