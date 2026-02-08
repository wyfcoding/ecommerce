// Package infrastructure 发票服务基础设施层
package infrastructure

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/invoice/domain"
	"gorm.io/gorm"
)

// GormInvoiceRepository GORM 发票仓储实现
type GormInvoiceRepository struct {
	db *gorm.DB
}

// NewGormInvoiceRepository 创建仓储
func NewGormInvoiceRepository(db *gorm.DB) *GormInvoiceRepository {
	return &GormInvoiceRepository{db: db}
}

// Save 保存发票
func (r *GormInvoiceRepository) Save(ctx context.Context, inv *domain.Invoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

// Update 更新发票
func (r *GormInvoiceRepository) Update(ctx context.Context, inv *domain.Invoice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(inv).Error; err != nil {
			return err
		}
		// 一般明细更新较少，若有变动需重新保存 items
		return nil
	})
}

// FindByID 根据ID查找
func (r *GormInvoiceRepository) FindByID(ctx context.Context, id uint) (*domain.Invoice, error) {
	var inv domain.Invoice
	if err := r.db.WithContext(ctx).Preload("Items").First(&inv, id).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}
	return &inv, nil
}

// FindByApplicationNo 根据申请单号查找
func (r *GormInvoiceRepository) FindByApplicationNo(ctx context.Context, no string) (*domain.Invoice, error) {
	var inv domain.Invoice
	if err := r.db.WithContext(ctx).Preload("Items").Where("application_no = ?", no).First(&inv).Error; err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}
	return &inv, nil
}

// FindByOrderNo 根据订单号查找
func (r *GormInvoiceRepository) FindByOrderNo(ctx context.Context, orderNo string) ([]*domain.Invoice, error) {
	var invoices []*domain.Invoice
	if err := r.db.WithContext(ctx).Preload("Items").Where("order_no = ?", orderNo).Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// List 列表查询
func (r *GormInvoiceRepository) List(ctx context.Context, filter *domain.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.Invoice{})

	if filter.UserID > 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.MerchantID > 0 {
		query = query.Where("merchant_id = ?", filter.MerchantID)
	}
	if filter.OrderNo != "" {
		query = query.Where("order_no = ?", filter.OrderNo)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var invoices []*domain.Invoice
	offset := (filter.Page - 1) * filter.PageSize
	if offset < 0 {
		offset = 0
	}

	if err := query.Preload("Items").Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&invoices).Error; err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}
