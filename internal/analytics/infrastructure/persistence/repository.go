package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain/repository"

	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) repository.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// Metric methods
func (r *analyticsRepository) CreateMetric(ctx context.Context, metric *entity.Metric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

func (r *analyticsRepository) GetMetric(ctx context.Context, id uint64) (*entity.Metric, error) {
	var metric entity.Metric
	if err := r.db.WithContext(ctx).First(&metric, id).Error; err != nil {
		return nil, err
	}
	return &metric, nil
}

func (r *analyticsRepository) ListMetrics(ctx context.Context, query *repository.MetricQuery) ([]*entity.Metric, int64, error) {
	var list []*entity.Metric
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Metric{})

	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.Granularity != "" {
		db = db.Where("granularity = ?", query.Granularity)
	}
	if query.Dimension != "" {
		db = db.Where("dimension = ?", query.Dimension)
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

func (r *analyticsRepository) DeleteMetric(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Metric{}, id).Error
}

// Dashboard methods
func (r *analyticsRepository) CreateDashboard(ctx context.Context, dashboard *entity.Dashboard) error {
	return r.db.WithContext(ctx).Create(dashboard).Error
}

func (r *analyticsRepository) GetDashboard(ctx context.Context, id uint64) (*entity.Dashboard, error) {
	var dashboard entity.Dashboard
	if err := r.db.WithContext(ctx).Preload("Metrics").Preload("Filters").First(&dashboard, id).Error; err != nil {
		return nil, err
	}
	return &dashboard, nil
}

func (r *analyticsRepository) ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Dashboard, int64, error) {
	var list []*entity.Dashboard
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Dashboard{}).Where("user_id = ? OR is_public = ?", userID, true)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *analyticsRepository) UpdateDashboard(ctx context.Context, dashboard *entity.Dashboard) error {
	return r.db.WithContext(ctx).Save(dashboard).Error
}

func (r *analyticsRepository) DeleteDashboard(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics", "Filters").Delete(&entity.Dashboard{}, id).Error
}

// Report methods
func (r *analyticsRepository) CreateReport(ctx context.Context, report *entity.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *analyticsRepository) GetReport(ctx context.Context, id uint64) (*entity.Report, error) {
	var report entity.Report
	if err := r.db.WithContext(ctx).Preload("Metrics").First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *analyticsRepository) ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Report, int64, error) {
	var list []*entity.Report
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Report{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *analyticsRepository) UpdateReport(ctx context.Context, report *entity.Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *analyticsRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics").Delete(&entity.Report{}, id).Error
}
