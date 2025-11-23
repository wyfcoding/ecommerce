package entity

import (
	"time"

	"gorm.io/gorm"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeSales       MetricType = "sales"      // 销售额
	MetricTypeOrders      MetricType = "orders"     // 订单数
	MetricTypeUsers       MetricType = "users"      // 用户数
	MetricTypeConversion  MetricType = "conversion" // 转化率
	MetricTypeRevenue     MetricType = "revenue"    // 收入
	MetricTypePageViews   MetricType = "page_views" // 页面浏览
	MetricTypeClickRate   MetricType = "click_rate" // 点击率
	MetricTypeAvgOrderVal MetricType = "avg_order"  // 平均订单值
)

// TimeGranularity 时间粒度
type TimeGranularity string

const (
	GranularityHourly  TimeGranularity = "hourly"
	GranularityDaily   TimeGranularity = "daily"
	GranularityWeekly  TimeGranularity = "weekly"
	GranularityMonthly TimeGranularity = "monthly"
	GranularityYearly  TimeGranularity = "yearly"
)

// Metric 指标聚合根
type Metric struct {
	gorm.Model
	MetricType   MetricType      `gorm:"type:varchar(32);not null;index;comment:指标类型" json:"metric_type"`
	MetricName   string          `gorm:"type:varchar(128);not null;comment:指标名称" json:"metric_name"`
	Value        float64         `gorm:"type:decimal(16,4);not null;comment:数值" json:"value"`
	Timestamp    time.Time       `gorm:"not null;index;comment:时间戳" json:"timestamp"`
	Granularity  TimeGranularity `gorm:"type:varchar(32);not null;comment:时间粒度" json:"granularity"`
	Dimension    string          `gorm:"type:varchar(64);comment:维度" json:"dimension"`
	DimensionVal string          `gorm:"type:varchar(128);comment:维度值" json:"dimension_val"`
}

// Dashboard 仪表板聚合根
type Dashboard struct {
	gorm.Model
	Name        string             `gorm:"type:varchar(128);not null;comment:名称" json:"name"`
	Description string             `gorm:"type:text;comment:描述" json:"description"`
	UserID      uint64             `gorm:"not null;index;comment:用户ID" json:"user_id"`
	IsPublic    bool               `gorm:"default:false;comment:是否公开" json:"is_public"`
	Metrics     []*DashboardMetric `gorm:"foreignKey:DashboardID" json:"metrics"`
	Filters     []*DashboardFilter `gorm:"foreignKey:DashboardID" json:"filters"`
}

// DashboardMetric 仪表板指标
type DashboardMetric struct {
	gorm.Model
	DashboardID uint64     `gorm:"not null;index;comment:仪表板ID" json:"dashboard_id"`
	MetricType  MetricType `gorm:"type:varchar(32);not null;comment:指标类型" json:"metric_type"`
	Title       string     `gorm:"type:varchar(128);not null;comment:标题" json:"title"`
	ChartType   string     `gorm:"type:varchar(32);not null;comment:图表类型" json:"chart_type"`
	Position    int32      `gorm:"not null;comment:位置" json:"position"`
}

// DashboardFilter 仪表板过滤器
type DashboardFilter struct {
	gorm.Model
	DashboardID uint64 `gorm:"not null;index;comment:仪表板ID" json:"dashboard_id"`
	FilterName  string `gorm:"type:varchar(64);not null;comment:过滤器名称" json:"filter_name"`
	FilterType  string `gorm:"type:varchar(32);not null;comment:过滤器类型" json:"filter_type"`
	FilterValue string `gorm:"type:varchar(255);not null;comment:过滤器值" json:"filter_value"`
}

// Report 报告聚合根
type Report struct {
	gorm.Model
	ReportNo    string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:报告编号" json:"report_no"`
	Title       string          `gorm:"type:varchar(128);not null;comment:标题" json:"title"`
	Description string          `gorm:"type:text;comment:描述" json:"description"`
	UserID      uint64          `gorm:"not null;index;comment:用户ID" json:"user_id"`
	ReportType  string          `gorm:"type:varchar(32);not null;comment:报告类型" json:"report_type"`
	StartDate   time.Time       `gorm:"comment:开始日期" json:"start_date"`
	EndDate     time.Time       `gorm:"comment:结束日期" json:"end_date"`
	Status      string          `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"`
	Content     string          `gorm:"type:longtext;comment:内容" json:"content"`
	PublishedAt *time.Time      `gorm:"comment:发布时间" json:"published_at"`
	Metrics     []*ReportMetric `gorm:"foreignKey:ReportID" json:"metrics"`
}

// ReportMetric 报告指标
type ReportMetric struct {
	gorm.Model
	ReportID uint64  `gorm:"not null;index;comment:报告ID" json:"report_id"`
	Metric   string  `gorm:"type:varchar(128);not null;comment:指标" json:"metric"`
	Value    float64 `gorm:"type:decimal(16,4);not null;comment:数值" json:"value"`
	Change   float64 `gorm:"type:decimal(10,4);comment:变化率" json:"change"`
	Trend    string  `gorm:"type:varchar(32);comment:趋势" json:"trend"`
}

// NewMetric 创建指标
func NewMetric(metricType MetricType, name string, value float64, granularity TimeGranularity) *Metric {
	return &Metric{
		MetricType:  metricType,
		MetricName:  name,
		Value:       value,
		Timestamp:   time.Now(),
		Granularity: granularity,
	}
}

// NewDashboard 创建仪表板
func NewDashboard(name, description string, userID uint64) *Dashboard {
	return &Dashboard{
		Name:        name,
		Description: description,
		UserID:      userID,
		IsPublic:    false,
	}
}

// AddMetric 添加指标到仪表板
func (d *Dashboard) AddMetric(metric *DashboardMetric) {
	metric.Position = int32(len(d.Metrics) + 1)
	d.Metrics = append(d.Metrics, metric)
}

// Publish 发布仪表板
func (d *Dashboard) Publish() {
	d.IsPublic = true
}

// NewReport 创建报告
func NewReport(reportNo, title, description string, userID uint64, reportType string) *Report {
	return &Report{
		ReportNo:    reportNo,
		Title:       title,
		Description: description,
		UserID:      userID,
		ReportType:  reportType,
		Status:      "draft",
	}
}

// Publish 发布报告
func (r *Report) Publish() {
	r.Status = "published"
	now := time.Now()
	r.PublishedAt = &now
}
