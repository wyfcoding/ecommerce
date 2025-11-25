package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/audit/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/audit/domain/repository"
	"time"

	"gorm.io/gorm"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) repository.AuditRepository {
	return &auditRepository{db: db}
}

// Log methods
func (r *auditRepository) CreateLog(ctx context.Context, log *entity.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) GetLog(ctx context.Context, id uint64) (*entity.AuditLog, error) {
	var log entity.AuditLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *auditRepository) ListLogs(ctx context.Context, query *repository.AuditLogQuery) ([]*entity.AuditLog, int64, error) {
	var list []*entity.AuditLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditLog{})

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

func (r *auditRepository) DeleteLog(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditLog{}, id).Error
}

func (r *auditRepository) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", beforeTime).Delete(&entity.AuditLog{}).Error
}

// Policy methods
func (r *auditRepository) CreatePolicy(ctx context.Context, policy *entity.AuditPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *auditRepository) GetPolicy(ctx context.Context, id uint64) (*entity.AuditPolicy, error) {
	var policy entity.AuditPolicy
	if err := r.db.WithContext(ctx).First(&policy, id).Error; err != nil {
		return nil, err
	}
	return &policy, nil
}

func (r *auditRepository) ListPolicies(ctx context.Context, offset, limit int) ([]*entity.AuditPolicy, int64, error) {
	var list []*entity.AuditPolicy
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditPolicy{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *auditRepository) UpdatePolicy(ctx context.Context, policy *entity.AuditPolicy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *auditRepository) DeletePolicy(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditPolicy{}, id).Error
}

// Report methods
func (r *auditRepository) CreateReport(ctx context.Context, report *entity.AuditReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *auditRepository) GetReport(ctx context.Context, id uint64) (*entity.AuditReport, error) {
	var report entity.AuditReport
	if err := r.db.WithContext(ctx).First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *auditRepository) ListReports(ctx context.Context, offset, limit int) ([]*entity.AuditReport, int64, error) {
	var list []*entity.AuditReport
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AuditReport{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *auditRepository) UpdateReport(ctx context.Context, report *entity.AuditReport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *auditRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.AuditReport{}, id).Error
}
