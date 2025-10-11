package biz

import (
	"context"
	v1 "ecommerce/api/product/v1"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Category 是商品分类的业务领域模型。
type Category struct {
	ID        uint64
	ParentID  uint64
	Name      string
	Level     int32
	Icon      *string // 使用指针以支持部分更新
	SortOrder *uint32
	IsVisible *bool
}

// ListCategories lists Categories.
func (uc *CategoryUsecase) ListCategories(ctx context.Context, parentID uint64) ([]*Category, error) {
	return uc.repo.ListCategories(ctx, parentID)
}

// Brand is a Brand model.
type Brand struct {
	ID          uint64
	Name        string
	Logo        *string
	Description *string
	Website     *string
	SortOrder   *uint32
	IsVisible   *bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BrandRepo is a Brand repo.
type BrandRepo interface {
	CreateBrand(ctx context.Context, brand *Brand) (*Brand, error)
	UpdateBrand(ctx context.Context, brand *Brand) (*Brand, error)
	DeleteBrand(ctx context.Context, id uint64) error
	ListBrands(ctx context.Context, pageSize, pageNum uint32, name *string, isVisible *bool) ([]*Brand, uint64, error)
}

// BrandUsecase is a Brand usecase.
type BrandUsecase struct {
	repo BrandRepo
}

// NewBrandUsecase creates a new BrandUsecase.
func NewBrandUsecase(repo BrandRepo) *BrandUsecase {
	return &BrandUsecase{repo: repo}
}

// CreateBrand creates a Brand.
func (uc *BrandUsecase) CreateBrand(ctx context.Context, req *v1.CreateBrandRequest) (*v1.BrandInfo, error) {
	brand := &Brand{
		Name: req.Name,
	}
	if req.HasLogo() {
		logo := req.GetLogo()
		brand.Logo = &logo
	}
	if req.HasDescription() {
		desc := req.GetDescription()
		brand.Description = &desc
	}
	if req.HasWebsite() {
		website := req.GetWebsite()
		brand.Website = &website
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		brand.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		brand.IsVisible = &isVisible
	}

	createdBrand, err := uc.repo.CreateBrand(ctx, brand)
	if err != nil {
		return nil, err
	}
	return bizBrandToProto(createdBrand), nil
}

// UpdateBrand updates a Brand.
func (uc *BrandUsecase) UpdateBrand(ctx context.Context, req *v1.UpdateBrandRequest) (*v1.BrandInfo, error) {
	brand := &Brand{
		ID: req.Id,
	}
	if req.HasName() {
		name := req.GetName()
		brand.Name = &name
	}
	if req.HasLogo() {
		logo := req.GetLogo()
		brand.Logo = &logo
	}
	if req.HasDescription() {
		desc := req.GetDescription()
		brand.Description = &desc
	}
	if req.HasWebsite() {
		website := req.GetWebsite()
		brand.Website = &website
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		brand.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		brand.IsVisible = &isVisible
	}

	updatedBrand, err := uc.repo.UpdateBrand(ctx, brand)
	if err != nil {
		return nil, err
	}
	return bizBrandToProto(updatedBrand), nil
}

// DeleteBrand deletes a Brand.
func (uc *BrandUsecase) DeleteBrand(ctx context.Context, id uint64) error {
	return uc.repo.DeleteBrand(ctx, id)
}

// ListBrands lists Brands.
func (uc *BrandUsecase) ListBrands(ctx context.Context, req *v1.ListBrandsRequest) ([]*v1.BrandInfo, uint64, error) {
	var name *string
	if req.HasName() {
		n := req.GetName()
		name = &n
	}
	var isVisible *bool
	if req.HasIsVisible() {
		v := req.GetIsVisible()
		isVisible = &v
	}

	brands, total, err := uc.repo.ListBrands(ctx, req.PageSize, req.PageNum, name, isVisible)
	if err != nil {
		return nil, 0, err
	}

	var brandInfos []*v1.BrandInfo
	for _, b := range brands {
		brandInfos = append(brandInfos, bizBrandToProto(b))
	}
	return brandInfos, total, nil
}

// bizBrandToProto converts biz.Brand to v1.BrandInfo
func bizBrandToProto(b *Brand) *v1.BrandInfo {
	if b == nil {
		return nil
	}
	res := &v1.BrandInfo{
		Id:   b.ID,
		Name: b.Name,
	}
	if b.Logo != nil {
		res.Logo = *b.Logo
	}
	if b.Description != nil {
		res.Description = *b.Description
	}
	if b.Website != nil {
		res.Website = *b.Website
	}
	if b.SortOrder != nil {
		res.SortOrder = *b.SortOrder
	}
	if b.IsVisible != nil {
		res.IsVisible = *b.IsVisible
	}
	return res
}

// Review is a Review model.
type Review struct {
	ID        uint64
	SpuID     uint64
	UserID    uint64
	Rating    uint32 // 评分 (1-5星)
	Comment   string
	Images    []string // 评论图片URL
	CreatedAt time.Time
}

// ReviewRepo is a Review repo.
type ReviewRepo interface {
	CreateReview(ctx context.Context, review *Review) (*Review, error)
	ListReviews(ctx context.Context, spuID uint64, pageSize, pageNum uint32, minRating *uint32) ([]*Review, uint64, error)
	DeleteReview(ctx context.Context, id uint64, userID uint64) error
}

// ReviewUsecase is a Review usecase.
type ReviewUsecase struct {
	repo ReviewRepo
}

// NewReviewUsecase creates a new ReviewUsecase.
func NewReviewUsecase(repo ReviewRepo) *ReviewUsecase {
	return &ReviewUsecase{repo: repo}
}

// CreateReview creates a Review.
func (uc *ReviewUsecase) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.ReviewInfo, error) {
	review := &Review{
		SpuID:   req.SpuId,
		UserID:  req.UserId,
		Rating:  req.Rating,
		Comment: req.Comment,
		Images:  req.Images,
	}
	createdReview, err := uc.repo.CreateReview(ctx, review)
	if err != nil {
		return nil, err
	}
	return bizReviewToProto(createdReview), nil
}

// ListReviews lists Reviews.
func (uc *ReviewUsecase) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) ([]*v1.ReviewInfo, uint64, error) {
	var minRating *uint32
	if req.HasMinRating() {
		mr := req.GetMinRating()
		minRating = &mr
	}
	reviews, total, err := uc.repo.ListReviews(ctx, req.SpuId, req.PageSize, req.PageNum, minRating)
	if err != nil {
		return nil, 0, err
	}
	var reviewInfos []*v1.ReviewInfo
	for _, r := range reviews {
		reviewInfos = append(reviewInfos, bizReviewToProto(r))
	}
	return reviewInfos, total, nil
}

// DeleteReview deletes a Review.
func (uc *ReviewUsecase) DeleteReview(ctx context.Context, id uint64, userID uint64) error {
	return uc.repo.DeleteReview(ctx, id, userID)
}

// bizReviewToProto converts biz.Review to v1.ReviewInfo
func bizReviewToProto(r *Review) *v1.ReviewInfo {
	if r == nil {
		return nil
	}
	return &v1.ReviewInfo{
		Id:        r.ID,
		SpuId:     r.SpuID,
		UserId:    r.UserID,
		Rating:    r.Rating,
		Comment:   r.Comment,
		Images:    r.Images,
		CreatedAt: r.CreatedAt.Format(time.RFC3339), // Format time to string
	}
}

// Spu is a Spu model.
type Spu struct {
	ID            uint64 // 数据库自增ID
	CategoryID    uint64
	BrandID       uint64
	Title         string
	SubTitle      string
	MainImage     string
	GalleryImages []string
	DetailHTML    string
	Status        int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Sku 是商品SKU的业务领域模型。
type Sku struct {
	ID            uint64 // 数据库自增ID
	SkuID         uint64 // 业务ID
	SpuID         uint64
	Title         string
	Price         uint64
	OriginalPrice uint64
	Stock         uint32
	Image         string
	Specs         map[string]string
	Status        int32
}

// MongoProduct represents a product document for MongoDB indexing in biz layer.
type MongoProduct struct {
	ID         string
	SpuID      uint64
	CategoryID uint64
	BrandID    uint64
	Title      string
	SubTitle   string
	MainImage  string
	DetailHTML string
	Status     int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CategoryRepo 定义了分类数据仓库的接口。
type CategoryRepo interface {
	CreateCategory(ctx context.Context, c *Category) (*Category, error)
	UpdateCategory(ctx context.Context, c *Category) (*Category, error)
	DeleteCategory(ctx context.Context, id uint64) error
	GetCategory(ctx context.Context, id uint64) (*Category, error)
	ListCategories(ctx context.Context, parentID uint64) ([]*Category, error)
}

// ProductRepo 定义了商品数据仓库的接口。
type ProductRepo interface {
	CreateSpu(ctx context.Context, spu *Spu) (*Spu, error)
	UpdateSpu(ctx context.Context, spu *Spu) (*Spu, error)
	DeleteSpu(ctx context.Context, spuID uint64) error
	GetSpu(ctx context.Context, spuID uint64) (*Spu, error)
	ListProducts(ctx context.Context, pageSize, pageNum uint32, categoryID *uint64, status *int32, brandID *uint64, minPrice *uint64, maxPrice *uint64, query *string, sortBy *string) ([]*Spu, uint64, error)

	CreateSku(ctx context.Context, sku *Sku) (*Sku, error)
	UpdateSku(ctx context.Context, sku *Sku) (*Sku, error)
	DeleteSkusBySpuID(ctx context.Context, spuID uint64) error
	DeleteSku(ctx context.Context, skuID uint64) error
	GetSkusBySpuID(ctx context.Context, spuID uint64) ([]*Sku, error)

	SaveProductToMongo(ctx context.Context, product *MongoProduct) error // Added for MongoDB indexing
}

// ProductUsecase is a Product usecase.
type ProductUsecase struct {
	repo   ProductRepo
	tracer trace.Tracer // Added OpenTelemetry tracer
}

// NewProductUsecase creates a new ProductUsecase.
func NewProductUsecase(repo ProductRepo, tracer trace.Tracer) *ProductUsecase {
	return &ProductUsecase{repo: repo, tracer: tracer}
}

// CreateProduct creates a Product.
func (uc *ProductUsecase) CreateProduct(ctx context.Context, spu *Spu, skus []*Sku) (createdSpu *Spu, createdSkus []*Sku, err error) {
	ctx, span := uc.tracer.Start(ctx, "ProductUsecase.CreateProduct")
	defer span.End()

	// Create SPU
	createdSpu, err = uc.repo.CreateSpu(ctx, spu)
	if err != nil {
		span.RecordError(err)
		return nil, nil, err
	}

	// Create SKUs
	var createdSkus []*Sku
	for _, sku := range skus {
		sku.SpuID = createdSpu.ID // Link SKU to the newly created SPU
		createdSku, err := uc.repo.CreateSku(ctx, sku)
		if err != nil {
			// TODO: Handle rollback for created SPU and previous SKUs if any
			return nil, nil, err
		}
		createdSkus = append(createdSkus, createdSku)
	}

	return createdSpu, createdSkus, nil
}

// UpdateProduct updates a Product.
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error) {
	// Update SPU
	updatedSpu, err := uc.repo.UpdateSpu(ctx, spu)
	if err != nil {
		return nil, nil, err
	}

	// Delete existing SKUs for this SPU
	err = uc.repo.DeleteSkusBySpuID(ctx, updatedSpu.ID)
	if err != nil {
		// TODO: Handle rollback for updated SPU
		return nil, nil, err
	}

	// Create new SKUs
	var createdSkus []*Sku
	for _, sku := range skus {
		sku.SpuID = updatedSpu.ID // Link SKU to the updated SPU
		createdSku, err := uc.repo.CreateSku(ctx, sku)
		if err != nil {
			// TODO: Handle rollback for updated SPU and previously created SKUs
			return nil, nil, err
		}
		createdSkus = append(createdSkus, createdSku)
	}

	return updatedSpu, createdSkus, nil
}

// DeleteProduct deletes a Product.
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, spuID uint64) error {
	// Delete SPU
	err := uc.repo.DeleteSpu(ctx, spuID)
	if err != nil {
		return err
	}

	// Delete all associated SKUs
	err = uc.repo.DeleteSkusBySpuID(ctx, spuID)
	if err != nil {
		// TODO: Handle rollback for deleted SPU if SKU deletion fails
		return err
	}

	return nil
}

// GetProductDetails gets Product details.
func (uc *ProductUsecase) GetProductDetails(ctx context.Context, spuID uint64) (*Spu, []*Sku, error) {
	// Get SPU
	spu, err := uc.repo.GetSpu(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	// Get associated SKUs
	skuss, err := uc.repo.GetSkusBySpuID(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	return spu, skuss, nil
}

// ListProducts lists Products.
func (uc *ProductUsecase) ListProducts(ctx context.Context, pageSize, pageNum uint32, categoryID *uint64, status *int32, brandID *uint64, minPrice *uint64, maxPrice *uint64, query *string, sortBy *string) ([]*Spu, uint64, error) {
	return uc.repo.ListProducts(ctx, pageSize, pageNum, categoryID, status, brandID, minPrice, maxPrice, query, sortBy)
}

// IndexProduct indexes a product in MongoDB.
func (uc *ProductUsecase) IndexProduct(ctx context.Context, spuID uint64) error {
	// Get SPU details from MySQL
	spu, err := uc.repo.GetSpu(ctx, spuID)
	if err != nil {
		return err
	}

	// Convert biz.Spu to MongoProduct
	mongoProduct := &MongoProduct{
		SpuID:      spu.ID,
		CategoryID: spu.CategoryID,
		BrandID:    spu.BrandID,
		Title:      spu.Title,
		SubTitle:   spu.SubTitle,
		MainImage:  spu.MainImage,
		DetailHTML: spu.DetailHTML,
		Status:     spu.Status,
		CreatedAt:  spu.CreatedAt,
		UpdatedAt:  spu.UpdatedAt,
	}

	// Save to MongoDB
	return uc.repo.SaveProductToMongo(ctx, mongoProduct)
}
