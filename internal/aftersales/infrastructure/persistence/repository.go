package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"     // 导入售后模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository" // 导入售后模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// afterSalesRepository 是 AfterSalesRepository 接口的GORM实现。
// 它负责将AfterSales模块的领域实体映射到数据库，并执行持久化操作。
type afterSalesRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAfterSalesRepository 创建并返回一个新的 afterSalesRepository 实例。
// db: GORM数据库连接实例。
func NewAfterSalesRepository(db *gorm.DB) repository.AfterSalesRepository {
	return &afterSalesRepository{db: db}
}

// Create 在数据库中创建一个新的售后申请记录。
func (r *afterSalesRepository) Create(ctx context.Context, afterSales *entity.AfterSales) error {
	return r.db.WithContext(ctx).Create(afterSales).Error
}

// GetByID 根据ID从数据库获取售后申请记录，并预加载其关联的商品项和操作日志。
func (r *afterSalesRepository) GetByID(ctx context.Context, id uint64) (*entity.AfterSales, error) {
	var afterSales entity.AfterSales
	// Preload "Items" 和 "Logs" 关联数据。
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").First(&afterSales, id).Error; err != nil {
		return nil, err
	}
	return &afterSales, nil
}

// GetByNo 根据售后单号从数据库获取售后申请记录，并预加载其关联的商品项和操作日志。
func (r *afterSalesRepository) GetByNo(ctx context.Context, no string) (*entity.AfterSales, error) {
	var afterSales entity.AfterSales
	// Preload "Items" 和 "Logs" 关联数据。
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").Where("after_sales_no = ?", no).First(&afterSales).Error; err != nil {
		return nil, err
	}
	return &afterSales, nil
}

// Update 更新数据库中的售后申请记录。
func (r *afterSalesRepository) Update(ctx context.Context, afterSales *entity.AfterSales) error {
	return r.db.WithContext(ctx).Save(afterSales).Error
}

// List 从数据库列出所有售后申请记录，支持通过查询条件进行过滤和分页。
func (r *afterSalesRepository) List(ctx context.Context, query *repository.AfterSalesQuery) ([]*entity.AfterSales, int64, error) {
	var list []*entity.AfterSales
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AfterSales{})

	// 根据查询条件构建WHERE子句。
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

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序，并预加载关联的商品项。
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at desc").Preload("Items").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateLog 在数据库中创建一条新的售后操作日志记录。
func (r *afterSalesRepository) CreateLog(ctx context.Context, log *entity.AfterSalesLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListLogs 列出指定售后申请的所有操作日志，按创建时间升序排列。
func (r *afterSalesRepository) ListLogs(ctx context.Context, afterSalesID uint64) ([]*entity.AfterSalesLog, error) {
	var logs []*entity.AfterSalesLog
	if err := r.db.WithContext(ctx).Where("after_sales_id = ?", afterSalesID).Order("created_at asc").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
