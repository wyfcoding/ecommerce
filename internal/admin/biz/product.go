package biz

import "context"

// Spu 是商品标准产品单元的业务领域模型。
type Spu struct {
	ID          uint64
	CategoryID  uint64
	BrandID     uint64
	Title       string
	SubTitle    string
	MainImage   string
	GalleryImages []string
	DetailHTML  string
	Status      int32
}

// Sku 是商品库存单位的业务领域模型。
type Sku struct {
	ID            uint64
	SpuID         uint64
	Title         string
	Price         uint64
	OriginalPrice uint64
	Stock         uint32
	Image         string
	Specs         map[string]string
	Status        int32
}

// ProductClient 定义了管理员服务依赖的商品服务客户端接口。
type ProductClient interface {
	CreateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error)
	UpdateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error)
	GetSpuDetail(ctx context.Context, spuID uint64) (*Spu, []*Sku, error)
}

// ProductUsecase 封装了商品相关的业务逻辑。
type ProductUsecase struct {
	client ProductClient
}

// NewProductUsecase 是 ProductUsecase 的构造函数。
func NewProductUsecase(client ProductClient) *ProductUsecase {
	return &ProductUsecase{client: client}
}

// CreateProduct 创建商品。
func (uc *ProductUsecase) CreateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error) {
	return uc.client.CreateProduct(ctx, spu, skus)
}

// UpdateProduct 更新商品。
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error) {
	return uc.client.UpdateProduct(ctx, spu, skus)
}

// GetSpuDetail 获取商品详情。
func (uc *ProductUsecase) GetSpuDetail(ctx context.Context, spuID uint64) (*Spu, []*Sku, error) {
	return uc.client.GetSpuDetail(ctx, spuID)
}
