// 生成摘要：
// - 实现 Fulfillment 服务的 MySQL 仓储层，遵循项目标准的 DDD 目录结构
// - 利用 GORM 进行履约单、商品项、包裹及异常信息的持久化
// - 支持预加载（Preload）关联实体，确保聚合根的完整性
// - 封装分页查询与过滤器逻辑

package mysql

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// FulfillmentRepositoryImpl GORM 履约单仓储实现
type FulfillmentRepositoryImpl struct {
	db *database.DB
}

// NewFulfillmentRepository 创建仓储实例
func NewFulfillmentRepository(db *database.DB) domain.FulfillmentRepository {
	return &FulfillmentRepositoryImpl{db: db}
}

// Save 保存履约单
func (r *FulfillmentRepositoryImpl) Save(ctx context.Context, f *domain.Fulfillment) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// Update 更新履约单及其关联实体
func (r *FulfillmentRepositoryImpl) Update(ctx context.Context, f *domain.Fulfillment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新主表
		if err := tx.Save(f).Error; err != nil {
			return err
		}
		// 更新商品项
		for i := range f.Items {
			f.Items[i].FulfillmentID = f.ID
			if err := tx.Save(&f.Items[i]).Error; err != nil {
				return err
			}
		}
		// 更新包裹
		for i := range f.Packages {
			f.Packages[i].FulfillmentID = f.ID
			if err := tx.Save(&f.Packages[i]).Error; err != nil {
				return err
			}
		}
		// 更新异常
		for i := range f.Exceptions {
			f.Exceptions[i].FulfillmentID = f.ID
			if err := tx.Save(&f.Exceptions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID 根据 ID 查找包含关联实体的履约单
func (r *FulfillmentRepositoryImpl) FindByID(ctx context.Context, id uint) (*domain.Fulfillment, error) {
	var f domain.Fulfillment
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Packages").
		Preload("Exceptions").
		First(&f, id).Error; err != nil {
		return nil, fmt.Errorf("fulfillment not found: %w", err)
	}
	return &f, nil
}

// FindByFulfillmentNo 根据履约单号查找
func (r *FulfillmentRepositoryImpl) FindByFulfillmentNo(ctx context.Context, no string) (*domain.Fulfillment, error) {
	var f domain.Fulfillment
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Packages").
		Where("fulfillment_no = ?", no).
		First(&f).Error; err != nil {
		return nil, fmt.Errorf("fulfillment not found: %w", err)
	}
	return &f, nil
}

// FindByOrderNo 根据订单号查找
func (r *FulfillmentRepositoryImpl) FindByOrderNo(ctx context.Context, orderNo string) ([]*domain.Fulfillment, error) {
	var fulfillments []*domain.Fulfillment
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("order_no = ?", orderNo).
		Find(&fulfillments).Error; err != nil {
		return nil, err
	}
	return fulfillments, nil
}

// List 列表查询与分页过滤
func (r *FulfillmentRepositoryImpl) List(ctx context.Context, filter *domain.FulfillmentFilter) ([]*domain.Fulfillment, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Fulfillment{})

	if filter.MerchantID > 0 {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.WarehouseID > 0 {
		query = query.Where("warehouse_id = ?", filter.WarehouseID)
	}
	if filter.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+filter.OrderNo+"%")
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var fulfillments []*domain.Fulfillment
	offset := (filter.Page - 1) * filter.PageSize
	if offset < 0 {
		offset = 0
	}

	if err := query.
		Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&fulfillments).Error; err != nil {
		return nil, 0, err
	}

	return fulfillments, total, nil
}

// Delete 删除履约单
func (r *FulfillmentRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Fulfillment{}, id).Error
}
