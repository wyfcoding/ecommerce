// Package infrastructure 履约服务基础设施层
package infrastructure

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/fulfillment/domain"
	"gorm.io/gorm"
)

// GormFulfillmentRepository GORM 履约单仓储实现
type GormFulfillmentRepository struct {
	db *gorm.DB
}

// NewGormFulfillmentRepository 创建仓储
func NewGormFulfillmentRepository(db *gorm.DB) *GormFulfillmentRepository {
	return &GormFulfillmentRepository{db: db}
}

// Save 保存履约单
func (r *GormFulfillmentRepository) Save(ctx context.Context, f *domain.Fulfillment) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// Update 更新履约单
func (r *GormFulfillmentRepository) Update(ctx context.Context, f *domain.Fulfillment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新主表
		if err := tx.Save(f).Error; err != nil {
			return err
		}
		// 更新商品项
		for _, item := range f.Items {
			item.FulfillmentID = f.ID
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
		}
		// 更新包裹
		for _, pkg := range f.Packages {
			pkg.FulfillmentID = f.ID
			if err := tx.Save(&pkg).Error; err != nil {
				return err
			}
		}
		// 更新异常
		for _, exc := range f.Exceptions {
			exc.FulfillmentID = f.ID
			if err := tx.Save(&exc).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID 根据ID查找
func (r *GormFulfillmentRepository) FindByID(ctx context.Context, id uint) (*domain.Fulfillment, error) {
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
func (r *GormFulfillmentRepository) FindByFulfillmentNo(ctx context.Context, no string) (*domain.Fulfillment, error) {
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
func (r *GormFulfillmentRepository) FindByOrderNo(ctx context.Context, orderNo string) ([]*domain.Fulfillment, error) {
	var fulfillments []*domain.Fulfillment
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Where("order_no = ?", orderNo).
		Find(&fulfillments).Error; err != nil {
		return nil, err
	}
	return fulfillments, nil
}

// List 列表查询
func (r *GormFulfillmentRepository) List(ctx context.Context, filter *domain.FulfillmentFilter) ([]*domain.Fulfillment, int64, error) {
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
	offset := max((filter.Page-1)*filter.PageSize, 0)

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
func (r *GormFulfillmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Fulfillment{}, id).Error
}
