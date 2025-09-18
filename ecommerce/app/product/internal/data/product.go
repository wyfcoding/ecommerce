package data

import (
	"context"
	"errors"
	"fmt"

	"ecommerce/ecommerce/app/order/internal/biz"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// productRepo 是 ProductRepo 的实现
type productRepo struct {
	db *gorm.DB
}

// NewProductRepo 创建一个新的 productRepo
func NewProductRepo(db *gorm.DB) biz.ProductRepo {
	return &productRepo{db: db}
}

func (r *productRepo) GetSkuInfosByIDs(ctx context.Context, skuIDs []uint64) ([]*biz.Sku, error) {
	var skus []*ProductSku
	if err := r.db.WithContext(ctx).Where("sku_id IN ?", skuIDs).Find(&skus).Error; err != nil {
		return nil, err
	}
	// ... 模型转换逻辑 ...
}

func (r *productRepo) LockStock(ctx context.Context, items []*biz.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var sku ProductSku

			// 1. 使用 FOR UPDATE 悲观锁，锁定要操作的行
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sku_id = ?", item.SkuID).First(&sku).Error; err != nil {
				return errors.New("failed to lock sku row")
			}

			// 2. 检查库存
			if sku.Stock < uint(item.Quantity) {
				return fmt.Errorf("sku %d insufficient stock, wants %d, has %d", item.SkuID, item.Quantity, sku.Stock)
			}

			// 3. 扣减库存
			newStock := sku.Stock - uint(item.Quantity)
			if err := tx.Model(&ProductSku{}).Where("sku_id = ?", item.SkuID).Update("stock", newStock).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListCategories 查询商品分类列表
func (r *productRepo) ListCategories(ctx context.Context, parentID uint64) ([]*biz.Category, error) {
	var categories []*ProductCategory
	if err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Order("sort_order asc").Find(&categories).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Category, 0, len(categories))
	for _, c := range categories {
		res = append(res, &biz.Category{
			ID:        c.ID,
			ParentID:  c.ParentID,
			Name:      c.Name,
			Level:     c.Level,
			Icon:      c.Icon,
			SortOrder: c.SortOrder,
			IsVisible: c.IsVisible,
		})
	}
	return res, nil
}

// GetSpu 查询 SPU 信息
func (r *productRepo) GetSpu(ctx context.Context, spuID uint64) (*biz.Spu, error) {
	var spu ProductSpu
	if err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).First(&spu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("spu not found") // 返回一个更业务化的错误
		}
		return nil, err
	}

	return &biz.Spu{
		SpuID:         spu.SpuID,
		CategoryID:    spu.CategoryID,
		BrandID:       spu.BrandID,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHTML:    spu.DetailHTML,
		Status:        spu.Status,
	}, nil
}

// ListSkusBySpuID 查询指定 SPU 下的所有 SKU
func (r *productRepo) ListSkusBySpuID(ctx context.Context, spuID uint64) ([]*biz.Sku, error) {
	var skus []*ProductSku
	if err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).Find(&skus).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Sku, 0, len(skus))
	for _, s := range skus {
		res = append(res, &biz.Sku{
			SkuID:         s.SkuID,
			SpuID:         s.SpuID,
			Title:         s.Title,
			Price:         s.Price,
			OriginalPrice: s.OriginalPrice,
			Stock:         s.Stock,
			Image:         s.Image,
			Specs:         s.Specs,
			Status:        s.Status,
		})
	}
	return res, nil
}
