package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository 创建并返回一个新的 analyticsRepository 实例。
func NewAnalyticsRepository(db *gorm.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *analyticsRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *analyticsRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *analyticsRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- Metric methods ---

func (r *analyticsRepository) SaveMetric(ctx context.Context, metric *domain.Metric) error {
	return r.saveMetricWithTx(ctx, r.db, metric)
}

func (r *analyticsRepository) SaveMetricInTx(ctx context.Context, tx any, metric *domain.Metric) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveMetricWithTx(ctx, gormTx, metric)
}

func (r *analyticsRepository) GetMetric(ctx context.Context, id uint64) (*domain.Metric, error) {
	var metric MetricModel
	if err := r.db.WithContext(ctx).First(&metric, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toMetric(&metric), nil
}

func (r *analyticsRepository) ListMetrics(ctx context.Context, query *domain.MetricQuery) ([]*domain.Metric, int64, error) {
	var list []*MetricModel
	var total int64

	db := r.db.WithContext(ctx).Model(&MetricModel{})

	if query != nil {
		if query.MetricType != "" {
			db = db.Where("metric_type = ?", query.MetricType)
		}
		if query.Granularity != "" {
			db = db.Where("granularity = ?", query.Granularity)
		}
		if query.Dimension != "" {
			db = db.Where("dimension = ?", query.Dimension)
		}
		if query.DimensionVal != "" {
			db = db.Where("dimension_val = ?", query.DimensionVal)
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
	pageSize := 10
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

	items := make([]*domain.Metric, len(list))
	for i, m := range list {
		items[i] = toMetric(m)
	}
	return items, total, nil
}

func (r *analyticsRepository) DeleteMetric(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&MetricModel{}, id).Error
}

// --- Dashboard methods ---

func (r *analyticsRepository) SaveDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	return r.saveDashboardWithTx(ctx, r.db, dashboard)
}

func (r *analyticsRepository) SaveDashboardInTx(ctx context.Context, tx any, dashboard *domain.Dashboard) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveDashboardWithTx(ctx, gormTx, dashboard)
}

func (r *analyticsRepository) GetDashboard(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	var dashboard DashboardModel
	if err := r.db.WithContext(ctx).Preload("Metrics").Preload("Filters").First(&dashboard, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDashboard(&dashboard), nil
}

func (r *analyticsRepository) ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Dashboard, int64, error) {
	var list []*DashboardModel
	var total int64

	db := r.db.WithContext(ctx).Model(&DashboardModel{}).Where("user_id = ? OR is_public = ?", userID, true)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Dashboard, len(list))
	for i, d := range list {
		items[i] = toDashboard(d)
	}
	return items, total, nil
}

func (r *analyticsRepository) DeleteDashboard(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics", "Filters").Delete(&DashboardModel{}, id).Error
}

// --- Report methods ---

func (r *analyticsRepository) SaveReport(ctx context.Context, report *domain.Report) error {
	return r.saveReportWithTx(ctx, r.db, report)
}

func (r *analyticsRepository) SaveReportInTx(ctx context.Context, tx any, report *domain.Report) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveReportWithTx(ctx, gormTx, report)
}

func (r *analyticsRepository) GetReport(ctx context.Context, id uint64) (*domain.Report, error) {
	var report ReportModel
	if err := r.db.WithContext(ctx).Preload("Metrics").First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toReport(&report), nil
}

func (r *analyticsRepository) ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Report, int64, error) {
	var list []*ReportModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ReportModel{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Report, len(list))
	for i, r := range list {
		items[i] = toReport(r)
	}
	return items, total, nil
}

func (r *analyticsRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics").Delete(&ReportModel{}, id).Error
}

// GetActivePages 从指标数据中聚合最近活跃的页面名称。
func (r *analyticsRepository) GetActivePages(ctx context.Context, limit int) ([]string, error) {
	var pages []string
	err := r.db.WithContext(ctx).Model(&MetricModel{}).
		Select("metric_name").
		Where("metric_type = ? AND timestamp >= ?", domain.MetricType("event"), time.Now().Add(-24*time.Hour)).
		Group("metric_name").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("metric_name", &pages).Error

	return pages, err
}

// --- internal helpers ---

func (r *analyticsRepository) saveMetricWithTx(ctx context.Context, tx *gorm.DB, metric *domain.Metric) error {
	if metric == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	model := toMetricModel(metric)
	if err := gormTx.Save(model).Error; err != nil {
		return err
	}
	if synced := toMetric(model); synced != nil {
		*metric = *synced
	}
	return nil
}

func (r *analyticsRepository) saveDashboardWithTx(ctx context.Context, tx *gorm.DB, dashboard *domain.Dashboard) error {
	if dashboard == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	model := toDashboardModel(dashboard)
	if err := gormTx.Session(&gorm.Session{FullSaveAssociations: true}).Save(model).Error; err != nil {
		return err
	}
	if synced := toDashboard(model); synced != nil {
		*dashboard = *synced
	}
	return nil
}

func (r *analyticsRepository) saveReportWithTx(ctx context.Context, tx *gorm.DB, report *domain.Report) error {
	if report == nil {
		return nil
	}
	gormTx := tx.WithContext(ctx)
	model := toReportModel(report)
	if err := gormTx.Session(&gorm.Session{FullSaveAssociations: true}).Save(model).Error; err != nil {
		return err
	}
	if synced := toReport(model); synced != nil {
		*report = *synced
	}
	return nil
}
