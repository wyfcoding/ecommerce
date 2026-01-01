package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain" // 导入分析模块的领域定义。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// analyticsRepository 是 AnalyticsRepository 接口的GORM实现。
type analyticsRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewAnalyticsRepository 创建并返回一个新的 analyticsRepository 实例。
func NewAnalyticsRepository(db *gorm.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// --- Metric methods ---

// CreateMetric 在数据库中创建一个新的指标记录。
func (r *analyticsRepository) CreateMetric(ctx context.Context, metric *domain.Metric) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// GetMetric 根据ID从数据库获取指标记录。
func (r *analyticsRepository) GetMetric(ctx context.Context, id uint64) (*domain.Metric, error) {
	var metric domain.Metric
	if err := r.db.WithContext(ctx).First(&metric, id).Error; err != nil {
		return nil, err
	}
	return &metric, nil
}

// ListMetrics 从数据库列出所有指标记录，支持通过查询条件进行过滤和分页。
func (r *analyticsRepository) ListMetrics(ctx context.Context, query *domain.MetricQuery) ([]*domain.Metric, int64, error) {
	var list []*domain.Metric
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Metric{})

	// 根据查询条件构建WHERE子句。
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

// DeleteMetric 根据ID从数据库删除指标记录。
func (r *analyticsRepository) DeleteMetric(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Metric{}, id).Error
}

// --- Dashboard methods ---

// CreateDashboard 在数据库中创建一个新的仪表板记录。
func (r *analyticsRepository) CreateDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	// 创建仪表板时会同时创建其关联的Metrics和Filters。
	return r.db.WithContext(ctx).Create(dashboard).Error
}

// GetDashboard 根据ID从数据库获取仪表板记录，并预加载其关联的指标和过滤器。
func (r *analyticsRepository) GetDashboard(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	var dashboard domain.Dashboard
	if err := r.db.WithContext(ctx).Preload("Metrics").Preload("Filters").First(&dashboard, id).Error; err != nil {
		return nil, err
	}
	return &dashboard, nil
}

// ListDashboards 从数据库列出仪表板记录，支持根据用户ID或是否公开进行过滤，并支持分页。
func (r *analyticsRepository) ListDashboards(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Dashboard, int64, error) {
	var list []*domain.Dashboard
	var total int64

	// 查询条件：用户ID匹配，或者仪表板是公开的。
	db := r.db.WithContext(ctx).Model(&domain.Dashboard{}).Where("user_id = ? OR is_public = ?", userID, true)

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

// UpdateDashboard 更新数据库中的仪表板记录。
func (r *analyticsRepository) UpdateDashboard(ctx context.Context, dashboard *domain.Dashboard) error {
	return r.db.WithContext(ctx).Save(dashboard).Error
}

// DeleteDashboard 根据ID从数据库删除仪表板记录。
func (r *analyticsRepository) DeleteDashboard(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics", "Filters").Delete(&domain.Dashboard{}, id).Error
}

// --- Report methods ---

// CreateReport 在数据库中创建一个新的报告记录。
func (r *analyticsRepository) CreateReport(ctx context.Context, report *domain.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetReport 根据ID从数据库获取报告记录，并预加载其关联的指标。
func (r *analyticsRepository) GetReport(ctx context.Context, id uint64) (*domain.Report, error) {
	var report domain.Report
	if err := r.db.WithContext(ctx).Preload("Metrics").First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReports 从数据库列出报告记录，支持根据用户ID过滤，并支持分页。
func (r *analyticsRepository) ListReports(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Report, int64, error) {
	var list []*domain.Report
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Report{}).Where("user_id = ?", userID)

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

// UpdateReport 更新数据库中的报告记录。
func (r *analyticsRepository) UpdateReport(ctx context.Context, report *domain.Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// DeleteReport 根据ID从数据库删除报告记录。
func (r *analyticsRepository) DeleteReport(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Metrics").Delete(&domain.Report{}, id).Error
}

// GetActivePages 从指标数据中聚合最近活跃的页面名称。
func (r *analyticsRepository) GetActivePages(ctx context.Context, limit int) ([]string, error) {
	var pages []string
	// 聚合最近 24 小时的事件
	err := r.db.WithContext(ctx).Model(&domain.Metric{}).
		Select("name").
		Where("metric_type = ? AND timestamp >= ?", domain.MetricType("event"), time.Now().Add(-24*time.Hour)).
		Group("name").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("name", &pages).Error

	return pages, err
}
