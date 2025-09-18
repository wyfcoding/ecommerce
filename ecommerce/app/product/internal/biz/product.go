package biz

import "context"

// --- Domain Models ---

// Category 领域模型
type Category struct {
	ID        uint64
	ParentID  uint64
	Name      string
	Level     uint8
	Icon      string
	SortOrder uint
	IsVisible bool
}

// Spu 领域模型
type Spu struct {
	SpuID         uint64
	CategoryID    uint64
	BrandID       uint64
	Title         string
	SubTitle      string
	MainImage     string
	GalleryImages []string
	DetailHTML    string
	Status        int8
}

// Sku 领域模型
type Sku struct {
	SkuID         uint64
	SpuID         uint64
	Title         string
	Price         uint64
	OriginalPrice uint64
	Stock         uint
	Image         string
	Specs         map[string]string
	Status        int8
}

// --- Repo Interface ---

// ProductRepo 定义了商品数据仓库的接口
type ProductRepo interface {
	ListCategories(ctx context.Context, parentID uint64) ([]*Category, error)
	GetSpu(ctx context.Context, spuID uint64) (*Spu, error)
	ListSkusBySpuID(ctx context.Context, spuID uint64) ([]*Sku, error)
}

// ProductUsecase 是商品业务逻辑的容器
type ProductUsecase struct {
	repo ProductRepo
	// log  *log.Helper // 通常会注入日志记录器
}

// NewProductUsecase 创建一个新的 ProductUsecase
func NewProductUsecase(repo ProductRepo /*, logger log.Logger*/) *ProductUsecase {
	return &ProductUsecase{
		repo: repo,
		// log:  log.NewHelper(logger),
	}
}

// ListCategories 实现了获取分类列表的业务逻辑
func (uc *ProductUsecase) ListCategories(ctx context.Context, parentID uint64) ([]*Category, error) {
	// 在这一层可以增加缓存逻辑。例如，先查询 Redis，如果未命中再查询数据库。
	// 目前，我们直接调用 repo。
	return uc.repo.ListCategories(ctx, parentID)
}

// GetSpuDetail 实现了获取 SPU 详情的业务逻辑
func (uc *ProductUsecase) GetSpuDetail(ctx context.Context, spuID uint64) (*Spu, []*Sku, error) {
	// 业务逻辑：获取 SPU 详情需要同时获取 SPU 信息和其下属的所有 SKU 信息。

	// 1. 获取 SPU 信息
	spu, err := uc.repo.GetSpu(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	// 2. 获取该 SPU 下的所有 SKU 信息
	skus, err := uc.repo.ListSkusBySpuID(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	// 对于更复杂的场景，比如还需要获取品牌信息、优惠信息等，
	// 可以使用 errgroup 并发执行，提升性能。

	// 3. 组装并返回结果
	return spu, skus, nil
}
