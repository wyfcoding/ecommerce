package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity"     // 导入审计模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/audit/domain/repository" // 导入审计模块的领域仓储接口。
	"time"                                                            // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// auditRepository 是 AuditRepository 接口的GORM实现。
// 它负责将审计模块的领域实体映射到数据库，并执行持久化操作。
type auditRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAuditRepository 创建并返回一个新的 auditRepository 实例。
// db: GORM数据库连接实例。
func NewAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &auditRepository{db: db}
}

// --- Log methods ---

// CreateLog 在数据库中创建一个新的审计日志记录。
func (r *auditRepository) CreateLog(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLog 根据ID从数据库获取审计日志记录。
func (r *auditRepository) GetLog(ctx context.Context, id uint64) (*entity.AuditLog, error) {
	var log entity.AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// ListLogs 从数据库列出所有审计日志记录，支持通过查询条件进行过滤和分页。
func (r *auditRepository) ListLogs(ctx context.Context, query *repository.AuditLogQuery) ([]*entity.AuditLog, int64, error) {
	var list []*entity.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})

	// 根据查询条件构建WHERE子句。
	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.EventType != "" {
		db = db.Where("event_type = ?", query.EventType)
	}
	if query.Module != "" {
		db = db.Where("module = ?", query.Module)
	}
	if query.ResourceType != "" {
		db = db.Where("resource_type = ?", query.ResourceType)
	}
	if !query.StartTime.IsZero() {
		db = db.Where("timestamp >= ?", query.StartTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("timestamp <= ?", query.EndTime)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("timestamp desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteLog 根据ID从数据库删除审计日志记录。
func (r *auditRepository) DeleteLog(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditLog{}, id).Error
}

// DeleteLogsBefore 删除指定时间点之前的审计日志记录，用于数据清理。
func (r *auditRepository) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", beforeTime).Delete(&entity.AuditLog{}).Error
}

// --- Policy methods ---

// CreatePolicy 在数据库中创建一个新的审计策略记录。
func (r *auditRepository) CreatePolicy(ctx context.Context, policy *entity.AuditPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// GetPolicy 根据ID从数据库获取审计策略记录。
func (r *auditRepository) GetPolicy(ctx context.Context, id uint64) (*entity.AuditPolicy, error) {
	var policy entity.AuditPolicy
	if err := r.db.WithContext(ctx).First(&policy, id).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies 从数据库列出所有审计策略记录，支持分页。
func (r *auditRepository) ListPolicies(ctx context.Context, offset, limit int) ([]*entity.AuditPolicy, int64, error) {
	var list []*entity.AuditPolicy
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditPolicy{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdatePolicy 更新数据库中的审计策略记录。
func (r *auditRepository) UpdatePolicy(ctx context.Context, policy *entity.AuditPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// DeletePolicy 根据ID从数据库删除审计策略记录。
func (r *auditRepository) DeletePolicy(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditPolicy{}, id).Error
}

// --- Report methods ---

// CreateReport 在数据库中创建一个新的审计报告记录。
func (r *auditRepository) CreateReport(ctx context.Context, report *entity.AuditReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetReport 根据ID从数据库获取审计报告记录。
func (r *auditRepository) GetReport(ctx context.Context, id uint64) (*entity.AuditReport, error) {
	var report entity.AuditReport
	if err := r.db.WithContext(ctx).First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports 从数据库列出所有审计报告记录，支持分页。
func (r *auditRepository) ListReports(ctx context.Context, offset, limit int) ([]*entity.AuditReport, int64, error) {
	var list []*entity.AuditReport
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditReport{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateReport 更新数据库中的审计报告记录。
func (r *auditRepository) UpdateReport(ctx context.Context, report *entity.AuditReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// DeleteReport 根据ID从数据库删除审计报告记录。
func (r *auditRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditReport{}, id).Error
}
