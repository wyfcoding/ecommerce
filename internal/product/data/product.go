package data

import (
	"context"
	"ecommerce/internal/product/biz"
	"ecommerce/internal/product/data/model"
)

type productRepo struct {
	*Data
}

// NewProductRepo 是 productRepo 的构造函数。
func NewProductRepo(data *Data) biz.ProductRepo {
	return &productRepo{Data: data}
}

// toBizSpu 将数据库模型 data.Spu 转换为业务领域模型 biz.Spu。
func (r *productRepo) toBizSpu(p *Spu) *biz.Spu {
	if p == nil {
		return nil
	}
	return &biz.Spu{
		ID:            uint64(p.ID),
		SpuID:         p.SpuID,
		CategoryID:    p.CategoryID,
		BrandID:       p.BrandID,
		Title:         p.Title,
		SubTitle:      p.SubTitle,
		MainImage:     p.MainImage,
		GalleryImages: p.GalleryImages,
		DetailHTML:    p.DetailHTML,
		Status:        p.Status,
	}
}

// toBizSku 将数据库模型 data.Sku 转换为业务领域模型 biz.Sku。
func (r *productRepo) toBizSku(s *Sku) *biz.Sku {
	if s == nil {
		return nil
	}
	return &biz.Sku{
		ID:            uint64(s.ID),
		SkuID:         s.SkuID,
		SpuID:         s.SpuID,
		Title:         s.Title,
		Price:         s.Price,
		OriginalPrice: s.OriginalPrice,
		Stock:         s.Stock,
		Image:         s.Image,
		Specs:         s.Specs,
		Status:        s.Status,
	}
}

// CreateSpu 创建一个新的 SPU 记录。
func (r *productRepo) CreateSpu(ctx context.Context, spu *biz.Spu) (*biz.Spu, error) {
	p := &Spu{
		SpuID:         spu.SpuID,
		CategoryID:    spu.CategoryID,
		BrandID:       spu.BrandID,
		Title:         spu.Title,
		SubTitle:      spu.SubTitle,
		MainImage:     spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHTML:    spu.DetailHTML,
		Status:        spu.Status,
	}
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	return r.toBizSpu(p), nil
}

// UpdateSpu 更新一个已有的 SPU 记录。
func (r *productRepo) UpdateSpu(ctx context.Context, spu *biz.Spu) (*biz.Spu, error) {
	var p Spu
	if err := r.db.WithContext(ctx).Where("spu_id = ?", spu.SpuID).First(&p).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	updates["category_id"] = spu.CategoryID
	updates["brand_id"] = spu.BrandID
	updates["title"] = spu.Title
	updates["sub_title"] = spu.SubTitle
	updates["main_image"] = spu.MainImage
	updates["gallery_images"] = GalleryImages(spu.GalleryImages)
	updates["detail_html"] = spu.DetailHTML
	updates["status"] = spu.Status

	if err := r.db.WithContext(ctx).Model(&p).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.toBizSpu(&p), nil
}

// DeleteSpu 删除一个 SPU 记录 (软删除)。
func (r *productRepo) DeleteSpu(ctx context.Context, spuID uint64) error {
	return r.db.WithContext(ctx).Where("spu_id = ?", spuID).Delete(&model.Spu{}).Error
}

// DeleteSku 删除一个 SKU 记录 (软删除)。
func (r *productRepo) DeleteSku(ctx context.Context, skuID uint64) error {
	return r.db.WithContext(ctx).Where("sku_id = ?", skuID).Delete(&model.Sku{}).Error
}

// GetSpu 获取单个 SPU 的详情。
func (r *productRepo) GetSpu(ctx context.Context, spuID uint64) (*biz.Spu, error) {
	var spu Spu
	if err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).First(&spu).Error; err != nil {
		return nil, err
	}
	return r.toBizSpu(&spu), nil
}

// ListSpu 获取 SPU 列表 (分页)。
func (r *productRepo) ListSpu(ctx context.Context, page, pageSize int) ([]*biz.Spu, int64, error) {
	var spus []*Spu
	var total int64

	db := r.db.WithContext(ctx).Model(&Spu{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Find(&spus).Error; err != nil {
		return nil, 0, err
	}

	var bizSpus []*biz.Spu
	for _, p := range spus {
		bizSpus = append(bizSpus, r.toBizSpu(p))
	}
	return bizSpus, total, nil
}

// CreateSku 创建一个新的 SKU 记录。
func (r *productRepo) CreateSku(ctx context.Context, sku *biz.Sku) (*biz.Sku, error) {
	s := &Sku{
		SkuID:         sku.SkuID,
		SpuID:         sku.SpuID,
		Title:         sku.Title,
		Price:         sku.Price,
		OriginalPrice: sku.OriginalPrice,
		Stock:         sku.Stock,
		Image:         sku.Image,
		Specs:         sku.Specs,
		Status:        sku.Status,
	}
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		return nil, err
	}
	return r.toBizSku(s), nil
}

// UpdateSku 更新一个已有的 SKU 记录。
func (r *productRepo) UpdateSku(ctx context.Context, sku *biz.Sku) (*biz.Sku, error) {
	var s Sku
	if err := r.db.WithContext(ctx).Where("sku_id = ?", sku.SkuID).First(&s).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	updates["title"] = sku.Title
	updates["price"] = sku.Price
	updates["original_price"] = sku.OriginalPrice
	updates["stock"] = sku.Stock
	updates["image"] = sku.Image
	updates["specs"] = Specs(sku.Specs)
	updates["status"] = sku.Status

	if err := r.db.WithContext(ctx).Model(&s).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.toBizSku(&s), nil
}

// DeleteSkusBySpuID 删除一个 SPU 下的所有 SKU 记录 (软删除)。
func (r *productRepo) DeleteSkusBySpuID(ctx context.Context, spuID uint64) error {
	return r.db.WithContext(ctx).Where("spu_id = ?", spuID).Delete(&Sku{}).Error
}

// GetSkusBySpuID 获取一个 SPU 下的所有 SKU。
func (r *productRepo) GetSkusBySpuID(ctx context.Context, spuID uint64) ([]*biz.Sku, error) {
	var skus []*model.Sku
	if err := r.db.WithContext(ctx).Where("spu_id = ?", spuID).Find(&skus).Error; err != nil {
		return nil, err
	}
	var bizSkus []*biz.Sku
	for _, s := range skus {
		bizSkus = append(bizSkus, r.toBizSku(s))
	}
	return bizSkus, nil
}
