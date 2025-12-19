package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"

	"gorm.io/gorm"
)

type auditRepository struct {
	db *gorm.DB
}

// NewAuditRepository 创建并返回一个新的 auditRepository 实例。
func NewAuditRepository(db *gorm.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

// --- Log methods ---

// CreateLog 在数据库中创建一个新的审计日志记录。
func (r *auditRepository) CreateLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLog 根据ID从数据库获取审计日志记录。
func (r *auditRepository) GetLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	var log domain.AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// ListLogs 从数据库列出所有审计日志记录，支持通过查询条件进行过滤和分页。
func (r *auditRepository) ListLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	var list []*domain.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.AuditLog{})

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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("timestamp desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteLog 根据ID从数据库删除审计日志记录。
func (r *auditRepository) DeleteLog(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditLog{}, id).Error
}

// DeleteLogsBefore 删除指定时间点之前的审计日志记录，用于数据清理。
func (r *auditRepository) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", beforeTime).Delete(&domain.AuditLog{}).Error
}

// --- Policy methods ---

// CreatePolicy 在数据库中创建一个新的审计策略记录。
func (r *auditRepository) CreatePolicy(ctx context.Context, policy *domain.AuditPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// GetPolicy 根据ID从数据库获取审计策略记录。
func (r *auditRepository) GetPolicy(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	var policy domain.AuditPolicy
	if err := r.db.WithContext(ctx).First(&policy, id).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies 从数据库列出所有审计策略记录，支持分页。
func (r *auditRepository) ListPolicies(ctx context.Context, offset, limit int) ([]*domain.AuditPolicy, int64, error) {
	var list []*domain.AuditPolicy
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.AuditPolicy{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdatePolicy 更新数据库中的审计策略记录。
func (r *auditRepository) UpdatePolicy(ctx context.Context, policy *domain.AuditPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// DeletePolicy 根据ID从数据库删除审计策略记录。
func (r *auditRepository) DeletePolicy(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditPolicy{}, id).Error
}

// --- Report methods ---

// CreateReport 在数据库中创建一个新的审计报告记录。
func (r *auditRepository) CreateReport(ctx context.Context, report *domain.AuditReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetReport 根据ID从数据库获取审计报告记录。
func (r *auditRepository) GetReport(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	var report domain.AuditReport
	if err := r.db.WithContext(ctx).First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports 从数据库列出所有审计报告记录，支持分页。
func (r *auditRepository) ListReports(ctx context.Context, offset, limit int) ([]*domain.AuditReport, int64, error) {
	var list []*domain.AuditReport
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.AuditReport{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateReport 更新数据库中的审计报告记录。
func (r *auditRepository) UpdateReport(ctx context.Context, report *domain.AuditReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// DeleteReport 根据ID从数据库删除审计报告记录。
func (r *auditRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.AuditReport{}, id).Error
}
