package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository"

	"gorm.io/gorm"
)

type afterSalesRepository struct {
	db *gorm.DB
}

func NewAfterSalesRepository(db *gorm.DB) repository.AfterSalesRepository {
	return &afterSalesRepository{db: db}
}

func (r *afterSalesRepository) Create(ctx context.Context, afterSales *entity.AfterSales) error {
	return r.db.WithContext(ctx).Create(afterSales).Error
}

func (r *afterSalesRepository) GetByID(ctx context.Context, id uint64) (*entity.AfterSales, error) {
	var afterSales entity.AfterSales
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").First(&afterSales, id).Error; err != nil {
		return nil, err
	}
	return &afterSales, nil
}

func (r *afterSalesRepository) GetByNo(ctx context.Context, no string) (*entity.AfterSales, error) {
	var afterSales entity.AfterSales
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").Where("after_sales_no = ?", no).First(&afterSales).Error; err != nil {
		return nil, err
	}
	return &afterSales, nil
}

func (r *afterSalesRepository) Update(ctx context.Context, afterSales *entity.AfterSales) error {
	return r.db.WithContext(ctx).Save(afterSales).Error
}

func (r *afterSalesRepository) List(ctx context.Context, query *repository.AfterSalesQuery) ([]*entity.AfterSales, int64, error) {
	var list []*entity.AfterSales
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AfterSales{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrderID > 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	if query.Type > 0 {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.AfterSalesNo != "" {
		db = db.Where("after_sales_no = ?", query.AfterSalesNo)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at desc").Preload("Items").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *afterSalesRepository) CreateLog(ctx context.Context, log *entity.AfterSalesLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *afterSalesRepository) ListLogs(ctx context.Context, afterSalesID uint64) ([]*entity.AfterSalesLog, error) {
	var logs []*entity.AfterSalesLog
	if err := r.db.WithContext(ctx).Where("after_sales_id = ?", afterSalesID).Order("created_at asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
