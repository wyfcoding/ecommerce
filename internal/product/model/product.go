package biz

import (
	"context"
	"ecommerce/pkg/snowflake"
)

// Transaction 定义了事务管理器接口。
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ProductUsecase 封装了商品相关的业务逻辑。
type ProductUsecase struct {
	repo ProductRepo
	tx   Transaction // 在更完善的设计中，会注入一个事务管理器
}

// NewProductUsecase 是 ProductUsecase 的构造函数。
func NewProductUsecase(repo ProductRepo, tx Transaction) *ProductUsecase {
	return &ProductUsecase{repo: repo, tx: tx}
}

// CreateProduct 负责原子性地创建SPU及其所有SKU。
func (uc *ProductUsecase) CreateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error) {
	var createdSpu *Spu
	var createdSkus []*Sku

	err := uc.tx.InTx(ctx, func(txCtx context.Context) error {
		// 1. 为 SPU 生成唯一 ID
		spu.SpuID = snowflake.GenID()
		var err error
		createdSpu, err = uc.repo.CreateSpu(txCtx, spu)
		if err != nil {
			return err
		}

		// 2. 为每个 SKU 生成唯一 ID 并关联 SPU ID，然后创建
		createdSkus = make([]*Sku, 0, len(skus))
		for _, sku := range skus {
			sku.SkuID = snowflake.GenID()
			sku.SpuID = createdSpu.SpuID
			createdSku, err := uc.repo.CreateSku(txCtx, sku)
			if err != nil {
				return err
			}
			createdSkus = append(createdSkus, createdSku)
		}
		return nil
	})

	return createdSpu, createdSkus, err
}

// UpdateProduct 负责原子性地更新SPU及其所有SKU。
func (uc *ProductUsecase) UpdateProduct(ctx context.Context, spu *Spu, skus []*Sku) (*Spu, []*Sku, error) {
	var updatedSpu *Spu
	var updatedSkus []*Sku

	err := uc.tx.InTx(ctx, func(txCtx context.Context) error {
		// 1. 更新 SPU 信息
		var err error
		updatedSpu, err = uc.repo.UpdateSpu(txCtx, spu)
		if err != nil {
			return err
		}

		// 2. 更新 SKU 信息
		updatedSkus = make([]*Sku, 0, len(skus))
		for _, sku := range skus {
			// 假设传入的 SKU 必须有 SkuID
			if sku.SkuID == 0 {
				continue // 或者返回错误
			}
			updatedSku, err := uc.repo.UpdateSku(txCtx, sku)
			if err != nil {
				return err
			}
			updatedSkus = append(updatedSkus, updatedSku)
		}
		return nil
	})

	return updatedSpu, updatedSkus, err
}

// DeleteProduct 负责原子性地删除SPU及其所有SKU。
func (uc *ProductUsecase) DeleteProduct(ctx context.Context, spuID uint64) error {
	err := uc.tx.InTx(ctx, func(txCtx context.Context) error {
		// 1. 删除所有 SKU
		if err := uc.repo.DeleteSkusBySpuID(txCtx, spuID); err != nil {
			return err
		}

		// 2. 删除 SPU
		return uc.repo.DeleteSpu(txCtx, spuID)
	})
	return err
}

// GetProductDetails 负责获取商品详情（包含SPU和所有SKU）。
func (uc *ProductUsecase) GetProductDetails(ctx context.Context, spuID uint64) (*Spu, []*Sku, error) {
	spu, err := uc.repo.GetSpu(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	skus, err := uc.repo.GetSkusBySpuID(ctx, spuID)
	if err != nil {
		return nil, nil, err
	}

	return spu, skus, nil
}

// ListProducts 负责获取商品SPU列表（分页）。
func (uc *ProductUsecase) ListProducts(ctx context.Context, page, pageSize int) ([]*Spu, int64, error) {
	return uc.repo.ListSpu(ctx, page, pageSize)
}
