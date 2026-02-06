package mysql

import (
	"context"
	"errors"
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

// --- tx helpers ---

func (r *auditRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *auditRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *auditRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *auditRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- Log methods ---

func (r *auditRepository) SaveLog(ctx context.Context, log *domain.AuditLog) error {
	return r.saveLogWithTx(ctx, r.db, log)
}

func (r *auditRepository) SaveLogInTx(ctx context.Context, tx any, log *domain.AuditLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveLogWithTx(ctx, gormTx, log)
}

func (r *auditRepository) GetLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	var log AuditLogModel
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAuditLog(&log), nil
}

func (r *auditRepository) ListLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	var list []*AuditLogModel
	var total int64

	db := r.db.WithContext(ctx).Model(&AuditLogModel{})

	if query != nil {
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
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("timestamp desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AuditLog, len(list))
	for i, l := range list {
		items[i] = toAuditLog(l)
	}
	return items, total, nil
}

func (r *auditRepository) DeleteLog(ctx context.Context, id uint64) error {
	return r.deleteLogWithTx(ctx, r.db, id)
}

func (r *auditRepository) DeleteLogInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteLogWithTx(ctx, gormTx, id)
}

func (r *auditRepository) DeleteLogsBefore(ctx context.Context, beforeTime time.Time) error {
	return r.db.WithContext(ctx).Where("timestamp < ?", beforeTime).Delete(&AuditLogModel{}).Error
}

// --- Policy methods ---

func (r *auditRepository) SavePolicy(ctx context.Context, policy *domain.AuditPolicy) error {
	return r.savePolicyWithTx(ctx, r.db, policy)
}

func (r *auditRepository) SavePolicyInTx(ctx context.Context, tx any, policy *domain.AuditPolicy) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.savePolicyWithTx(ctx, gormTx, policy)
}

func (r *auditRepository) GetPolicy(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	var policy AuditPolicyModel
	if err := r.db.WithContext(ctx).First(&policy, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAuditPolicy(&policy), nil
}

func (r *auditRepository) ListPolicies(ctx context.Context, offset, limit int) ([]*domain.AuditPolicy, int64, error) {
	var list []*AuditPolicyModel
	var total int64

	db := r.db.WithContext(ctx).Model(&AuditPolicyModel{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AuditPolicy, len(list))
	for i, p := range list {
		items[i] = toAuditPolicy(p)
	}
	return items, total, nil
}

func (r *auditRepository) DeletePolicy(ctx context.Context, id uint64) error {
	return r.deletePolicyWithTx(ctx, r.db, id)
}

func (r *auditRepository) DeletePolicyInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deletePolicyWithTx(ctx, gormTx, id)
}

// --- Report methods ---

func (r *auditRepository) SaveReport(ctx context.Context, report *domain.AuditReport) error {
	return r.saveReportWithTx(ctx, r.db, report)
}

func (r *auditRepository) SaveReportInTx(ctx context.Context, tx any, report *domain.AuditReport) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveReportWithTx(ctx, gormTx, report)
}

func (r *auditRepository) GetReport(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	var report AuditReportModel
	if err := r.db.WithContext(ctx).First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAuditReport(&report), nil
}

func (r *auditRepository) ListReports(ctx context.Context, offset, limit int) ([]*domain.AuditReport, int64, error) {
	var list []*AuditReportModel
	var total int64

	db := r.db.WithContext(ctx).Model(&AuditReportModel{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AuditReport, len(list))
	for i, r := range list {
		items[i] = toAuditReport(r)
	}
	return items, total, nil
}

func (r *auditRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.deleteReportWithTx(ctx, r.db, id)
}

func (r *auditRepository) DeleteReportInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteReportWithTx(ctx, gormTx, id)
}

// --- internal helpers ---

func (r *auditRepository) saveLogWithTx(ctx context.Context, tx *gorm.DB, log *domain.AuditLog) error {
	if log == nil {
		return nil
	}
	data := toAuditLogModel(log)
	if err := tx.WithContext(ctx).Save(data).Error; err != nil {
		return err
	}
	if synced := toAuditLog(data); synced != nil {
		*log = *synced
	}
	return nil
}

func (r *auditRepository) deleteLogWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&AuditLogModel{}, id).Error
}

func (r *auditRepository) savePolicyWithTx(ctx context.Context, tx *gorm.DB, policy *domain.AuditPolicy) error {
	if policy == nil {
		return nil
	}
	data := toAuditPolicyModel(policy)
	if err := tx.WithContext(ctx).Save(data).Error; err != nil {
		return err
	}
	if synced := toAuditPolicy(data); synced != nil {
		*policy = *synced
	}
	return nil
}

func (r *auditRepository) deletePolicyWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&AuditPolicyModel{}, id).Error
}

func (r *auditRepository) saveReportWithTx(ctx context.Context, tx *gorm.DB, report *domain.AuditReport) error {
	if report == nil {
		return nil
	}
	data := toAuditReportModel(report)
	if err := tx.WithContext(ctx).Save(data).Error; err != nil {
		return err
	}
	if synced := toAuditReport(data); synced != nil {
		*report = *synced
	}
	return nil
}

func (r *auditRepository) deleteReportWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&AuditReportModel{}, id).Error
}
