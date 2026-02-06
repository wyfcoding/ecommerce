package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"gorm.io/gorm"
)

// MetricModel 指标写模型。
type MetricModel struct {
	gorm.Model
	MetricType   domain.MetricType      `gorm:"column:metric_type;type:varchar(32);not null;index;comment:指标类型"`
	MetricName   string                 `gorm:"column:metric_name;type:varchar(128);not null;comment:指标名称"`
	Value        float64                `gorm:"column:value;type:decimal(16,4);not null;comment:数值"`
	Timestamp    time.Time              `gorm:"column:timestamp;not null;index;comment:时间戳"`
	Granularity  domain.TimeGranularity `gorm:"column:granularity;type:varchar(32);not null;comment:时间粒度"`
	Dimension    string                 `gorm:"column:dimension;type:varchar(64);comment:维度"`
	DimensionVal string                 `gorm:"column:dimension_val;type:varchar(128);comment:维度值"`
}

// DashboardModel 仪表板写模型。
type DashboardModel struct {
	gorm.Model
	Name        string `gorm:"column:name;type:varchar(128);not null;comment:名称"`
	Description string `gorm:"column:description;type:text;comment:描述"`
	UserID      uint64 `gorm:"column:user_id;not null;index;comment:用户ID"`
	IsPublic    bool   `gorm:"column:is_public;default:false;comment:是否公开"`

	Metrics []DashboardMetricModel `gorm:"foreignKey:DashboardID"`
	Filters []DashboardFilterModel `gorm:"foreignKey:DashboardID"`
}

// DashboardMetricModel 仪表板指标写模型。
type DashboardMetricModel struct {
	gorm.Model
	DashboardID uint64           `gorm:"column:dashboard_id;not null;index;comment:仪表板ID"`
	MetricType  domain.MetricType `gorm:"column:metric_type;type:varchar(32);not null;comment:指标类型"`
	Title       string            `gorm:"column:title;type:varchar(128);not null;comment:标题"`
	ChartType   string            `gorm:"column:chart_type;type:varchar(32);not null;comment:图表类型"`
	Position    int32             `gorm:"column:position;not null;comment:位置"`
}

// DashboardFilterModel 仪表板过滤器写模型。
type DashboardFilterModel struct {
	gorm.Model
	DashboardID uint64 `gorm:"column:dashboard_id;not null;index;comment:仪表板ID"`
	FilterName  string `gorm:"column:filter_name;type:varchar(64);not null;comment:过滤器名称"`
	FilterType  string `gorm:"column:filter_type;type:varchar(32);not null;comment:过滤器类型"`
	FilterValue string `gorm:"column:filter_value;type:varchar(255);not null;comment:过滤器值"`
}

// ReportModel 报告写模型。
type ReportModel struct {
	gorm.Model
	ReportNo    string     `gorm:"column:report_no;type:varchar(64);uniqueIndex;not null;comment:报告编号"`
	Title       string     `gorm:"column:title;type:varchar(128);not null;comment:标题"`
	Description string     `gorm:"column:description;type:text;comment:描述"`
	UserID      uint64     `gorm:"column:user_id;not null;index;comment:用户ID"`
	ReportType  string     `gorm:"column:report_type;type:varchar(32);not null;comment:报告类型"`
	StartDate   time.Time  `gorm:"column:start_date;comment:开始日期"`
	EndDate     time.Time  `gorm:"column:end_date;comment:结束日期"`
	Status      string     `gorm:"column:status;type:varchar(32);not null;default:'draft';comment:状态"`
	Content     string     `gorm:"column:content;type:longtext;comment:内容"`
	PublishedAt *time.Time `gorm:"column:published_at;comment:发布时间"`

	Metrics []ReportMetricModel `gorm:"foreignKey:ReportID"`
}

// ReportMetricModel 报告指标写模型。
type ReportMetricModel struct {
	gorm.Model
	ReportID uint64  `gorm:"column:report_id;not null;index;comment:报告ID"`
	Metric   string  `gorm:"column:metric;type:varchar(128);not null;comment:指标"`
	Value    float64 `gorm:"column:value;type:decimal(16,4);not null;comment:数值"`
	Change   float64 `gorm:"column:change;type:decimal(10,4);comment:变化率"`
	Trend    string  `gorm:"column:trend;type:varchar(32);comment:趋势"`
}

func (MetricModel) TableName() string         { return "metrics" }
func (DashboardModel) TableName() string      { return "dashboards" }
func (DashboardMetricModel) TableName() string { return "dashboard_metrics" }
func (DashboardFilterModel) TableName() string { return "dashboard_filters" }
func (ReportModel) TableName() string          { return "reports" }
func (ReportMetricModel) TableName() string    { return "report_metrics" }

func toMetricModel(metric *domain.Metric) *MetricModel {
	if metric == nil {
		return nil
	}
	return &MetricModel{
		Model: gorm.Model{
			ID:        metric.ID,
			CreatedAt: metric.CreatedAt,
			UpdatedAt: metric.UpdatedAt,
		},
		MetricType:   metric.MetricType,
		MetricName:   metric.MetricName,
		Value:        metric.Value,
		Timestamp:    metric.Timestamp,
		Granularity:  metric.Granularity,
		Dimension:    metric.Dimension,
		DimensionVal: metric.DimensionVal,
	}
}

func toMetric(model *MetricModel) *domain.Metric {
	if model == nil {
		return nil
	}
	return &domain.Metric{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		MetricType:   model.MetricType,
		MetricName:   model.MetricName,
		Value:        model.Value,
		Timestamp:    model.Timestamp,
		Granularity:  model.Granularity,
		Dimension:    model.Dimension,
		DimensionVal: model.DimensionVal,
	}
}

func toDashboardModel(dashboard *domain.Dashboard) *DashboardModel {
	if dashboard == nil {
		return nil
	}
	model := &DashboardModel{
		Model: gorm.Model{
			ID:        dashboard.ID,
			CreatedAt: dashboard.CreatedAt,
			UpdatedAt: dashboard.UpdatedAt,
		},
		Name:        dashboard.Name,
		Description: dashboard.Description,
		UserID:      dashboard.UserID,
		IsPublic:    dashboard.IsPublic,
	}

	if len(dashboard.Metrics) > 0 {
		model.Metrics = make([]DashboardMetricModel, len(dashboard.Metrics))
		for i, m := range dashboard.Metrics {
			item := toDashboardMetricModel(m)
			if item != nil {
				if item.DashboardID == 0 {
					item.DashboardID = uint64(dashboard.ID)
				}
				model.Metrics[i] = *item
			}
		}
	}
	if len(dashboard.Filters) > 0 {
		model.Filters = make([]DashboardFilterModel, len(dashboard.Filters))
		for i, f := range dashboard.Filters {
			item := toDashboardFilterModel(f)
			if item != nil {
				if item.DashboardID == 0 {
					item.DashboardID = uint64(dashboard.ID)
				}
				model.Filters[i] = *item
			}
		}
	}
	return model
}

func toDashboard(model *DashboardModel) *domain.Dashboard {
	if model == nil {
		return nil
	}
	d := &domain.Dashboard{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Description: model.Description,
		UserID:      model.UserID,
		IsPublic:    model.IsPublic,
	}
	if len(model.Metrics) > 0 {
		metrics := make([]*domain.DashboardMetric, len(model.Metrics))
		for i := range model.Metrics {
			metrics[i] = toDashboardMetric(&model.Metrics[i])
		}
		d.Metrics = metrics
	}
	if len(model.Filters) > 0 {
		filters := make([]*domain.DashboardFilter, len(model.Filters))
		for i := range model.Filters {
			filters[i] = toDashboardFilter(&model.Filters[i])
		}
		d.Filters = filters
	}
	return d
}

func toDashboardMetricModel(metric *domain.DashboardMetric) *DashboardMetricModel {
	if metric == nil {
		return nil
	}
	return &DashboardMetricModel{
		Model: gorm.Model{
			ID:        metric.ID,
			CreatedAt: metric.CreatedAt,
			UpdatedAt: metric.UpdatedAt,
		},
		DashboardID: metric.DashboardID,
		MetricType:  metric.MetricType,
		Title:       metric.Title,
		ChartType:   metric.ChartType,
		Position:    metric.Position,
	}
}

func toDashboardMetric(model *DashboardMetricModel) *domain.DashboardMetric {
	if model == nil {
		return nil
	}
	return &domain.DashboardMetric{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DashboardID: model.DashboardID,
		MetricType:  model.MetricType,
		Title:       model.Title,
		ChartType:   model.ChartType,
		Position:    model.Position,
	}
}

func toDashboardFilterModel(filter *domain.DashboardFilter) *DashboardFilterModel {
	if filter == nil {
		return nil
	}
	return &DashboardFilterModel{
		Model: gorm.Model{
			ID:        filter.ID,
			CreatedAt: filter.CreatedAt,
			UpdatedAt: filter.UpdatedAt,
		},
		DashboardID: filter.DashboardID,
		FilterName:  filter.FilterName,
		FilterType:  filter.FilterType,
		FilterValue: filter.FilterValue,
	}
}

func toDashboardFilter(model *DashboardFilterModel) *domain.DashboardFilter {
	if model == nil {
		return nil
	}
	return &domain.DashboardFilter{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DashboardID: model.DashboardID,
		FilterName:  model.FilterName,
		FilterType:  model.FilterType,
		FilterValue: model.FilterValue,
	}
}

func toReportModel(report *domain.Report) *ReportModel {
	if report == nil {
		return nil
	}
	model := &ReportModel{
		Model: gorm.Model{
			ID:        report.ID,
			CreatedAt: report.CreatedAt,
			UpdatedAt: report.UpdatedAt,
		},
		ReportNo:    report.ReportNo,
		Title:       report.Title,
		Description: report.Description,
		UserID:      report.UserID,
		ReportType:  report.ReportType,
		StartDate:   report.StartDate,
		EndDate:     report.EndDate,
		Status:      report.Status,
		Content:     report.Content,
		PublishedAt: report.PublishedAt,
	}
	if len(report.Metrics) > 0 {
		model.Metrics = make([]ReportMetricModel, len(report.Metrics))
		for i, m := range report.Metrics {
			item := toReportMetricModel(m)
			if item != nil {
				if item.ReportID == 0 {
					item.ReportID = uint64(report.ID)
				}
				model.Metrics[i] = *item
			}
		}
	}
	return model
}

func toReport(model *ReportModel) *domain.Report {
	if model == nil {
		return nil
	}
	r := &domain.Report{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		ReportNo:    model.ReportNo,
		Title:       model.Title,
		Description: model.Description,
		UserID:      model.UserID,
		ReportType:  model.ReportType,
		StartDate:   model.StartDate,
		EndDate:     model.EndDate,
		Status:      model.Status,
		Content:     model.Content,
		PublishedAt: model.PublishedAt,
	}
	if len(model.Metrics) > 0 {
		metrics := make([]*domain.ReportMetric, len(model.Metrics))
		for i := range model.Metrics {
			metrics[i] = toReportMetric(&model.Metrics[i])
		}
		r.Metrics = metrics
	}
	return r
}

func toReportMetricModel(metric *domain.ReportMetric) *ReportMetricModel {
	if metric == nil {
		return nil
	}
	return &ReportMetricModel{
		Model: gorm.Model{
			ID:        metric.ID,
			CreatedAt: metric.CreatedAt,
			UpdatedAt: metric.UpdatedAt,
		},
		ReportID: metric.ReportID,
		Metric:   metric.Metric,
		Value:    metric.Value,
		Change:   metric.Change,
		Trend:    metric.Trend,
	}
}

func toReportMetric(model *ReportMetricModel) *domain.ReportMetric {
	if model == nil {
		return nil
	}
	return &domain.ReportMetric{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		ReportID:  model.ReportID,
		Metric:    model.Metric,
		Value:     model.Value,
		Change:    model.Change,
		Trend:     model.Trend,
	}
}
