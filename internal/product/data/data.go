package data

import (
	"context"
	"encoding/json"
	"strings"
	"ecommerce/internal/product/biz"
	"ecommerce/internal/product/data/model"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Data 结构体持有所有数据源的连接，如此处的数据库连接
type Data struct {
	db *gorm.DB
	log *zap.SugaredLogger // 添加日志器
}

// NewData 是 Data 结构体的构造函数
// 它接收数据库配置，初始化数据库连接，并返回一个包含连接的 Data 实例
func NewData(dsn string, logger *zap.SugaredLogger) (*Data, func(), error) {
	// 使用 GORM 连接到 MySQL 数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Errorf("failed to connect to database: %v", err) // 使用注入的 logger
		return nil, nil, err
	}

	// 定义一个清理函数，用于在服务关闭时关闭数据库连接
	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Errorf("failed to get database instance for cleanup: %v", err) // 使用注入的 logger
			return
		}
		if sqlDB != nil {
			logger.Info("closing database connection...") // 使用注入的 logger
			if err := sqlDB.Close(); err != nil {
				logger.Errorf("failed to close database connection: %v", err) // 使用注入的 logger
			}
		}
	}

	// 自动迁移数据库表结构
	// 这会根据 model.go 中定义的结构体创建或更新对应的表
	if err := db.AutoMigrate(&model.Category{}, &model.Brand{}, &model.Spu{}, &model.Sku{}, &model.Review{}); err != nil {
		cleanup()
		return nil, nil, err
	}

	return &Data{db: db, log: logger}, cleanup, nil // 初始化 log 字段
}

// categoryRepo 实现了 biz.CategoryRepo 接口
type categoryRepo struct {
	data *Data
}

// NewCategoryRepo 创建一个新的 CategoryRepo 实例
func NewCategoryRepo(data *Data) biz.CategoryRepo {
	return &categoryRepo{data: data}
}

// CreateCategory 实现创建分类
func (r *categoryRepo) CreateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	categoryModel := &model.Category{
		ParentID:  c.ParentID,
		Name:      c.Name,
		Level:     c.Level,
		Icon:      *c.Icon,
		SortOrder: *c.SortOrder,
		IsVisible: *c.IsVisible,
	}
	if err := r.data.db.WithContext(ctx).Create(categoryModel).Error; err != nil {
		r.data.log.Errorf("failed to create category: %v", err)
		return nil, err
	}
	c.ID = categoryModel.ID
	c.CreatedAt = categoryModel.CreatedAt
	c.UpdatedAt = categoryModel.UpdatedAt
	return c, nil
}

// UpdateCategory 实现更新分类
func (r *categoryRepo) UpdateCategory(ctx context.Context, c *biz.Category) (*biz.Category, error) {
	categoryModel := &model.Category{ID: c.ID}
	if err := r.data.db.WithContext(ctx).First(categoryModel).Error; err != nil {
		r.data.log.Errorf("category not found: %v", err)
		return nil, err
	}

	categoryModel.ParentID = c.ParentID
	categoryModel.Name = c.Name
	categoryModel.Level = c.Level
	if c.Icon != nil {
		categoryModel.Icon = *c.Icon
	}
	if c.SortOrder != nil {
		categoryModel.SortOrder = *c.SortOrder
	}
	if c.IsVisible != nil {
		categoryModel.IsVisible = *c.IsVisible
	}

	if err := r.data.db.WithContext(ctx).Save(categoryModel).Error; err != nil {
		r.data.log.Errorf("failed to update category: %v", err)
		return nil, err
	}
	c.CreatedAt = categoryModel.CreatedAt
	c.UpdatedAt = categoryModel.UpdatedAt
	return c, nil
}

// DeleteCategory 实现删除分类
func (r *categoryRepo) DeleteCategory(ctx context.Context, id uint64) error {
	if err := r.data.db.WithContext(ctx).Delete(&model.Category{}, id).Error; err != nil {
		r.data.log.Errorf("failed to delete category: %v", err)
		return err
	}
	return nil
}

// ListCategories 实现获取分类列表
func (r *categoryRepo) ListCategories(ctx context.Context, parentID uint64) ([]*biz.Category, error) {
	var categoryModels []*model.Category
	db := r.data.db.WithContext(ctx)

	if parentID > 0 {
		db = db.Where("parent_id = ?", parentID)
	}

	if err := db.Find(&categoryModels).Error; err != nil {
		r.data.log.Errorf("failed to list categories: %v", err)
		return nil, err
	}

	var bizCategories []*biz.Category
	for _, cm := range categoryModels {
		bizCategories = append(bizCategories, &biz.Category{
			ID:        cm.ID,
			ParentID:  cm.ParentID,
			Name:      cm.Name,
			Level:     cm.Level,
			Icon:      &cm.Icon,
			SortOrder: &cm.SortOrder,
			IsVisible: &cm.IsVisible,
			CreatedAt: cm.CreatedAt,
			UpdatedAt: cm.UpdatedAt,
		})
	}
	return bizCategories, nil
}

// productRepo 实现了 biz.ProductRepo 接口
type productRepo struct {
	data *Data
}

// NewProductRepo 创建一个新的 ProductRepo 实例
func NewProductRepo(data *Data) biz.ProductRepo {
	return &productRepo{data: data}
}

// CreateProduct 实现创建商品
func (r *productRepo) CreateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	productModel := &model.Product{
		SpuID:         spu.SpuID,
		CategoryID:    *spu.CategoryID,
		BrandID:       *spu.BrandID,
		Title:         *spu.Title,
		SubTitle:      *spu.SubTitle,
		MainImage:     *spu.MainImage,
		GalleryImages: "", // Will be handled below
		DetailHTML:    *spu.DetailHTML,
		Status:        *spu.Status,
	}
	if len(spu.GalleryImages) > 0 {
		productModel.GalleryImages = strings.Join(spu.GalleryImages, ",")
	}

	var skuModels []*model.Sku
	for _, s := range skus {
		skuModel := &model.Sku{
			SkuID:         s.SkuID,
			SpuID:         s.SpuID,
			Title:         *s.Title,
			Price:         *s.Price,
			OriginalPrice: *s.OriginalPrice,
			Stock:         *s.Stock,
			Image:         *s.Image,
			Specs:         "", // Will be handled below
			Status:        *s.Status,
		}
		if s.Specs != nil {
			specsJSON, err := json.Marshal(s.Specs)
			if err != nil {
				r.data.log.Errorf("failed to marshal SKU specs: %v", err)
				return nil, nil, err
			}
			skuModel.Specs = string(specsJSON)
		}
		skuModels = append(skuModels, skuModel)
	}

	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(productModel).Error; err != nil {
			return err
		}
		for _, sm := range skuModels {
			sm.SpuID = productModel.SpuID // Ensure SKU is linked to the created SPU
			if err := tx.Create(sm).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		r.data.log.Errorf("failed to create product and skus: %v", err)
		return nil, nil, err
	}

	spu.SpuID = productModel.SpuID
	spu.CreatedAt = productModel.CreatedAt
	spu.UpdatedAt = productModel.UpdatedAt
	for i, s := range skus {
		s.SkuID = skuModels[i].SkuID
		s.CreatedAt = skuModels[i].CreatedAt
		s.UpdatedAt = skuModels[i].UpdatedAt
	}

	return spu, skus, nil
}

// UpdateProduct 实现更新商品
func (r *productRepo) UpdateProduct(ctx context.Context, spu *biz.Spu, skus []*biz.Sku) (*biz.Spu, []*biz.Sku, error) {
	productModel := &model.Product{SpuID: spu.SpuID}
	if err := r.data.db.WithContext(ctx).First(productModel).Error; err != nil {
		r.data.log.Errorf("product not found: %v", err)
		return nil, nil, err
	}

	if spu.CategoryID != nil {
		productModel.CategoryID = *spu.CategoryID
	}
	if spu.BrandID != nil {
		productModel.BrandID = *spu.BrandID
	}
	if spu.Title != nil {
		productModel.Title = *spu.Title
	}
	if spu.SubTitle != nil {
		productModel.SubTitle = *spu.SubTitle
	}
	if spu.MainImage != nil {
		productModel.MainImage = *spu.MainImage
	}
	if len(spu.GalleryImages) > 0 {
		productModel.GalleryImages = strings.Join(spu.GalleryImages, ",")
	}
	if spu.DetailHTML != nil {
		productModel.DetailHTML = *spu.DetailHTML
	}
	if spu.Status != nil {
		productModel.Status = *spu.Status
	}

	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(productModel).Error; err != nil {
			return err
		}

		for _, s := range skus {
			skuModel := &model.Sku{SkuID: s.SkuID}
			if err := tx.First(skuModel).Error; err != nil {
				return err // SKU not found, or handle as creation if SkuID is 0
			}

			if s.Title != nil {
				skuModel.Title = *s.Title
			}
			if s.Price != nil {
				skuModel.Price = *s.Price
			}
			if s.OriginalPrice != nil {
				skuModel.OriginalPrice = *s.OriginalPrice
			}
			if s.Stock != nil {
				skuModel.Stock = *s.Stock
			}
			if s.Image != nil {
				skuModel.Image = *s.Image
			}
			if s.Specs != nil {
				specsJSON, err := json.Marshal(s.Specs)
				if err != nil {
					return err
				}
				skuModel.Specs = string(specsJSON)
			}
			if s.Status != nil {
				skuModel.Status = *s.Status
			}
			if err := tx.Save(skuModel).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		r.data.log.Errorf("failed to update product and skus: %v", err)
		return nil, nil, err
	}

	spu.CreatedAt = productModel.CreatedAt
	spu.UpdatedAt = productModel.UpdatedAt
	for i, s := range skus {
		s.CreatedAt = skus[i].CreatedAt
		s.UpdatedAt = skus[i].UpdatedAt
	}

	return spu, skus, nil
}

// DeleteProduct 实现删除商品
func (r *productRepo) DeleteProduct(ctx context.Context, spuID uint64) error {
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("spu_id = ?", spuID).Delete(&model.Sku{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Product{}, spuID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.data.log.Errorf("failed to delete product and skus: %v", err)
		return err
	}
	return nil
}

// GetProductDetails 实现获取商品详情
func (r *productRepo) GetProductDetails(ctx context.Context, spuID uint64) (*biz.Spu, []*biz.Sku, error) {
	productModel := &model.Product{}
	if err := r.data.db.WithContext(ctx).Preload("Skus").First(productModel, spuID).Error; err != nil {
		r.data.log.Errorf("product not found: %v", err)
		return nil, nil, err
	}

	spu := &biz.Spu{
		SpuID:         productModel.SpuID,
		CategoryID:    &productModel.CategoryID,
		BrandID:       &productModel.BrandID,
		Title:         &productModel.Title,
		SubTitle:      &productModel.SubTitle,
		MainImage:     &productModel.MainImage,
		GalleryImages: strings.Split(productModel.GalleryImages, ","),
		DetailHTML:    &productModel.DetailHTML,
		Status:        &productModel.Status,
		CreatedAt:     productModel.CreatedAt,
		UpdatedAt:     productModel.UpdatedAt,
	}

	var bizSkus []*biz.Sku
	for _, sm := range productModel.Skus {
		var specsMap map[string]string
		if sm.Specs != "" {
			if err := json.Unmarshal([]byte(sm.Specs), &specsMap); err != nil {
				r.data.log.Errorf("failed to unmarshal SKU specs: %v", err)
				return nil, nil, err
			}
		}
		bizSkus = append(bizSkus, &biz.Sku{
			SkuID:         sm.SkuID,
			SpuID:         sm.SpuID,
			Title:         &sm.Title,
			Price:         &sm.Price,
			OriginalPrice: &sm.OriginalPrice,
			Stock:         &sm.Stock,
			Image:         &sm.Image,
			Specs:         specsMap,
			Status:        &sm.Status,
			CreatedAt:     sm.CreatedAt,
			UpdatedAt:     sm.UpdatedAt,
		})
	}
	return spu, bizSkus, nil
}

// ListProducts 实现获取商品列表
func (r *productRepo) ListProducts(ctx context.Context, pageSize, pageNum uint32, categoryID *uint64, status *int32, brandID *uint64, minPrice *uint64, maxPrice *uint64, query *string, sortBy *string) ([]*biz.Spu, uint64, error) {
	var productModels []*model.Product
	db := r.data.db.WithContext(ctx)

	if categoryID != nil && *categoryID > 0 {
		db = db.Where("category_id = ?", *categoryID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if brandID != nil && *brandID > 0 {
		db = db.Where("brand_id = ?", *brandID)
	}
	if minPrice != nil {
		db = db.Where("price >= ?", *minPrice)
	}
	if maxPrice != nil {
		db = db.Where("price <= ?", *maxPrice)
	}
	if query != nil && *query != "" {
		db = db.Where("title LIKE ? OR sub_title LIKE ? OR detail_html LIKE ?", "%"+*query+"%", "%"+*query+"%", "%"+*query+"%")
	}

	// Sorting
	if sortBy != nil && *sortBy != "" {
		switch *sortBy {
		case "price_asc":
			db = db.Order("price ASC")
		case "price_desc":
			db = db.Order("price DESC")
		case "sales_desc": // Assuming a 'sales' field exists or can be derived
			db = db.Order("sales DESC")
		case "created_at_desc":
			db = db.Order("created_at DESC")
		default:
			// Default sort order if sortBy is invalid or not provided
			db = db.Order("created_at DESC")
		}
	} else {
		// Default sort if no sortBy is provided
		db = db.Order("created_at DESC")
	}

	var total int64
	if err := db.Model(&model.Product{}).Count(&total).Error; err != nil {
		r.data.log.Errorf("failed to count products: %v", err)
		return nil, 0, err
	}

	if pageSize > 0 && pageNum > 0 {
		offset := (pageNum - 1) * pageSize
		db = db.Limit(int(pageSize)).Offset(int(offset))
	}

	if err := db.Find(&productModels).Error; err != nil {
		r.data.log.Errorf("failed to list products: %v", err)
		return nil, 0, err
	}

	var bizSpus []*biz.Spu
	for _, pm := range productModels {
		bizSpus = append(bizSpus, &biz.Spu{
			SpuID:         pm.SpuID,
			CategoryID:    &pm.CategoryID,
			BrandID:       &pm.BrandID,
			Title:         &pm.Title,
			SubTitle:      &pm.SubTitle,
			MainImage:     &pm.MainImage,
			GalleryImages: strings.Split(pm.GalleryImages, ","),
			DetailHTML:    &pm.DetailHTML,
			Status:        &pm.Status,
			CreatedAt:     pm.CreatedAt,
			UpdatedAt:     pm.UpdatedAt,
		})
	}
	return bizSpus, uint64(total), nil
}

// brandRepo 实现了 biz.BrandRepo 接口
type brandRepo struct {
	data *Data
}

// NewBrandRepo 创建一个新的 BrandRepo 实例
func NewBrandRepo(data *Data) biz.BrandRepo {
	return &brandRepo{data: data}
}

// CreateBrand 实现创建品牌
func (r *brandRepo) CreateBrand(ctx context.Context, b *biz.Brand) (*biz.Brand, error) {
	brandModel := &model.Brand{
		Name:        b.Name,
		Logo:        *b.Logo,
		Description: *b.Description,
		Website:     *b.Website,
		SortOrder:   *b.SortOrder,
		IsVisible:   *b.IsVisible,
	}
	if err := r.data.db.WithContext(ctx).Create(brandModel).Error; err != nil {
		r.data.log.Errorf("failed to create brand: %v", err)
		return nil, err
	}
	b.ID = brandModel.ID
	b.CreatedAt = brandModel.CreatedAt
	b.UpdatedAt = brandModel.UpdatedAt
	return b, nil
}

// UpdateBrand 实现更新品牌
func (r *brandRepo) UpdateBrand(ctx context.Context, b *biz.Brand) (*biz.Brand, error) {
	brandModel := &model.Brand{ID: b.ID}
	if err := r.data.db.WithContext(ctx).First(brandModel).Error; err != nil {
		r.data.log.Errorf("brand not found: %v", err)
		return nil, err
	}

	if b.Name != "" {
		brandModel.Name = b.Name
	}
	if b.Logo != nil {
		brandModel.Logo = *b.Logo
	}
	if b.Description != nil {
		brandModel.Description = *b.Description
	}
	if b.Website != nil {
		brandModel.Website = *b.Website
	}
	if b.SortOrder != nil {
		brandModel.SortOrder = *b.SortOrder
	}
	if b.IsVisible != nil {
		brandModel.IsVisible = *b.IsVisible
	}

	if err := r.data.db.WithContext(ctx).Save(brandModel).Error; err != nil {
		r.data.log.Errorf("failed to update brand: %v", err)
		return nil, err
	}
	b.CreatedAt = brandModel.CreatedAt
	b.UpdatedAt = brandModel.UpdatedAt
	return b, nil
}

// DeleteBrand 实现删除品牌
func (r *brandRepo) DeleteBrand(ctx context.Context, id uint64) error {
	if err := r.data.db.WithContext(ctx).Delete(&model.Brand{}, id).Error; err != nil {
		r.data.log.Errorf("failed to delete brand: %v", err)
		return err
	}
	return nil
}

// ListBrands 实现获取品牌列表
func (r *brandRepo) ListBrands(ctx context.Context, pageSize, pageNum uint32, name *string, isVisible *bool) ([]*biz.Brand, uint64, error) {
	var brands []*model.Brand
	db := r.data.db.WithContext(ctx)

	if name != nil && *name != "" {
		db = db.Where("name LIKE ? ", "%"+*name+"%")
	}
	if isVisible != nil {
		db = db.Where("is_visible = ?", *isVisible)
	}

	var total int64
	if err := db.Model(&model.Brand{}).Count(&total).Error; err != nil {
		r.data.log.Errorf("failed to count brands: %v", err)
		return nil, 0, err
	}

	if pageSize > 0 && pageNum > 0 {
		offset := (pageNum - 1) * pageSize
		db = db.Limit(int(pageSize)).Offset(int(offset))
	}

	if err := db.Find(&brands).Error; err != nil {
		r.data.log.Errorf("failed to list brands: %v", err)
		return nil, 0, err
	}

	var bizBrands []*biz.Brand
	for _, b := range brands {
		bizBrands = append(bizBrands, &biz.Brand{
			ID:          b.ID,
			Name:        b.Name,
			Logo:        &b.Logo,
			Description: &b.Description,
			Website:     &b.Website,
			SortOrder:   &b.SortOrder,
			IsVisible:   &b.IsVisible,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		})
	}
	return bizBrands, uint64(total), nil
}

// reviewRepo 实现了 biz.ReviewRepo 接口
type reviewRepo struct {
	data *Data
}

// NewReviewRepo 创建一个新的 ReviewRepo 实例
func NewReviewRepo(data *Data) biz.ReviewRepo {
	return &reviewRepo{data: data}
}

// CreateReview 实现创建评论
func (r *reviewRepo) CreateReview(ctx context.Context, review *biz.Review) (*biz.Review, error) {
	reviewModel := &model.Review{
		SpuID:   review.SpuID,
		UserID:  review.UserID,
		Rating:  review.Rating,
		Comment: review.Comment,
		Images:  strings.Join(review.Images, ","),
	}
	if err := r.data.db.WithContext(ctx).Create(reviewModel).Error; err != nil {
		r.data.log.Errorf("failed to create review: %v", err)
		return nil, err
	}
	review.ID = reviewModel.ID
	review.CreatedAt = reviewModel.CreatedAt
	review.UpdatedAt = reviewModel.UpdatedAt
	return review, nil
}

// ListReviews 实现获取评论列表
func (r *reviewRepo) ListReviews(ctx context.Context, spuID uint64, pageSize, pageNum uint32, minRating *uint32) ([]*biz.Review, uint64, error) {
	var reviewModels []*model.Review
	db := r.data.db.WithContext(ctx)

	db = db.Where("spu_id = ?", spuID)
	if minRating != nil && *minRating > 0 {
		db = db.Where("rating >= ?", *minRating)
	}

	var total int64
	if err := db.Model(&model.Review{}).Where("spu_id = ?", spuID).Count(&total).Error; err != nil {
		r.data.log.Errorf("failed to count reviews: %v", err)
		return nil, 0, err
	}

	if pageSize > 0 && pageNum > 0 {
		offset := (pageNum - 1) * pageSize
		db = db.Limit(int(pageSize)).Offset(int(offset))
	}

	if err := db.Find(&reviewModels).Error; err != nil {
		r.data.log.Errorf("failed to list reviews: %v", err)
		return nil, 0, err
	}

	var bizReviews []*biz.Review
	for _, rm := range reviewModels {
		bizReviews = append(bizReviews, &biz.Review{
			ID:        rm.ID,
			SpuID:     rm.SpuID,
			UserID:    rm.UserID,
			Rating:    rm.Rating,
			Comment:   rm.Comment,
			Images:    strings.Split(rm.Images, ","),
			CreatedAt: rm.CreatedAt,
			UpdatedAt: rm.UpdatedAt,
		})
	}
	return bizReviews, uint64(total), nil
}

// DeleteReview 实现删除评论
func (r *reviewRepo) DeleteReview(ctx context.Context, id uint64, userID uint64) error {
	result := r.data.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Review{})
	if result.Error != nil {
		r.data.log.Errorf("failed to delete review: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Or a custom error indicating not found or not authorized
	}
	return nil
}

// InTx 实现 biz.Transaction 接口，用于在事务中执行操作。
func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(fn)
}