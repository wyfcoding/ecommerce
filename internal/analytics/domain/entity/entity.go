package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// MetricType 定义了可用的指标类型。
type MetricType string

const (
	MetricTypeSales       MetricType = "sales"        // 销售额指标。
	MetricTypeOrders      MetricType = "orders"       // 订单数指标。
	MetricTypeUsers       MetricType = "users"        // 用户数指标。
	MetricTypeConversion  MetricType = "conversion"   // 转化率指标。
	MetricTypeRevenue     MetricType = "revenue"      // 收入指标。
	MetricTypePageViews   MetricType = "page_views"   // 页面浏览量指标。
	MetricTypeClickRate   MetricType = "click_rate"   // 点击率指标。
	MetricTypeAvgOrderVal MetricType = "avg_order"    // 平均订单值指标。
	MetricTypeActiveUsers MetricType = "active_users" // 活跃用户数指标。
)

// TimeGranularity 定义了时间粒度。
type TimeGranularity string

const (
	GranularityHourly  TimeGranularity = "hourly"  // 按小时统计。
	GranularityDaily   TimeGranularity = "daily"   // 按天统计。
	GranularityWeekly  TimeGranularity = "weekly"  // 按周统计。
	GranularityMonthly TimeGranularity = "monthly" // 按月统计。
	GranularityYearly  TimeGranularity = "yearly"  // 按年统计。
)

// Metric 实体代表一个聚合根，用于存储具体的业务指标数据。
// 这些指标可以是销售额、订单数等，支持按时间粒度和维度进行记录。
type Metric struct {
	gorm.Model                   // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	MetricType   MetricType      `gorm:"type:varchar(32);not null;index;comment:指标类型" json:"metric_type"` // 指标类型，索引字段，不允许为空。
	MetricName   string          `gorm:"type:varchar(128);not null;comment:指标名称" json:"metric_name"`      // 指标名称。
	Value        float64         `gorm:"type:decimal(16,4);not null;comment:数值" json:"value"`             // 指标的具体数值。
	Timestamp    time.Time       `gorm:"not null;index;comment:时间戳" json:"timestamp"`                     // 指标记录的时间戳，索引字段。
	Granularity  TimeGranularity `gorm:"type:varchar(32);not null;comment:时间粒度" json:"granularity"`       // 指标的时间粒度。
	Dimension    string          `gorm:"type:varchar(64);comment:维度" json:"dimension"`                    // 指标的维度名称，例如“地区”、“渠道”。
	DimensionVal string          `gorm:"type:varchar(128);comment:维度值" json:"dimension_val"`              // 指标的维度值，例如“华东”、“App”。
}

// Dashboard 实体代表一个聚合根，用于组织和展示多个指标的图表。
// 用户可以创建和定制自己的仪表板来监控关键业务数据。
type Dashboard struct {
	gorm.Model                     // 嵌入gorm.Model。
	Name        string             `gorm:"type:varchar(128);not null;comment:名称" json:"name"` // 仪表板名称。
	Description string             `gorm:"type:text;comment:描述" json:"description"`           // 仪表板描述。
	UserID      uint64             `gorm:"not null;index;comment:用户ID" json:"user_id"`        // 创建仪表板的用户ID，索引字段。
	IsPublic    bool               `gorm:"default:false;comment:是否公开" json:"is_public"`       // 仪表板是否公开，默认为不公开。
	Metrics     []*DashboardMetric `gorm:"foreignKey:DashboardID" json:"metrics"`             // 仪表板上包含的指标图表列表，一对多关系。
	Filters     []*DashboardFilter `gorm:"foreignKey:DashboardID" json:"filters"`             // 仪表板上应用的过滤器列表，一对多关系。
}

// DashboardMetric 实体代表仪表板上显示的一个指标图表。
type DashboardMetric struct {
	gorm.Model             // 嵌入gorm.Model。
	DashboardID uint64     `gorm:"not null;index;comment:仪表板ID" json:"dashboard_id"`          // 关联的仪表板ID，索引字段。
	MetricType  MetricType `gorm:"type:varchar(32);not null;comment:指标类型" json:"metric_type"` // 显示的指标类型。
	Title       string     `gorm:"type:varchar(128);not null;comment:标题" json:"title"`        // 图表的标题。
	ChartType   string     `gorm:"type:varchar(32);not null;comment:图表类型" json:"chart_type"`  // 图表的类型，例如“折线图”、“柱状图”。
	Position    int32      `gorm:"not null;comment:位置" json:"position"`                       // 图表在仪表板中的显示位置或顺序。
}

// DashboardFilter 实体代表仪表板上的一个过滤器。
type DashboardFilter struct {
	gorm.Model         // 嵌入gorm.Model。
	DashboardID uint64 `gorm:"not null;index;comment:仪表板ID" json:"dashboard_id"`            // 关联的仪表板ID，索引字段。
	FilterName  string `gorm:"type:varchar(64);not null;comment:过滤器名称" json:"filter_name"`  // 过滤器名称。
	FilterType  string `gorm:"type:varchar(32);not null;comment:过滤器类型" json:"filter_type"`  // 过滤器类型，例如“时间范围”、“地区选择”。
	FilterValue string `gorm:"type:varchar(255);not null;comment:过滤器值" json:"filter_value"` // 过滤器的值。
}

// Report 实体代表一个聚合根，用于存储生成的数据分析报告。
// 报告可以是定期生成的，也可以是按需生成的，通常包含对业务数据的深入分析。
type Report struct {
	gorm.Model                  // 嵌入gorm.Model。
	ReportNo    string          `gorm:"type:varchar(64);uniqueIndex;not null;comment:报告编号" json:"report_no"` // 报告的唯一编号，唯一索引，不允许为空。
	Title       string          `gorm:"type:varchar(128);not null;comment:标题" json:"title"`                  // 报告标题。
	Description string          `gorm:"type:text;comment:描述" json:"description"`                             // 报告描述。
	UserID      uint64          `gorm:"not null;index;comment:用户ID" json:"user_id"`                          // 创建报告的用户ID，索引字段。
	ReportType  string          `gorm:"type:varchar(32);not null;comment:报告类型" json:"report_type"`           // 报告类型，例如“周报”、“月报”、“专项分析报告”。
	StartDate   time.Time       `gorm:"comment:开始日期" json:"start_date"`                                      // 报告涵盖的开始日期。
	EndDate     time.Time       `gorm:"comment:结束日期" json:"end_date"`                                        // 报告涵盖的结束日期。
	Status      string          `gorm:"type:varchar(32);not null;default:'draft';comment:状态" json:"status"`  // 报告状态，例如“draft”（草稿）、“published”（已发布）。
	Content     string          `gorm:"type:longtext;comment:内容" json:"content"`                             // 报告的详细内容，可以存储HTML或Markdown等格式。
	PublishedAt *time.Time      `gorm:"comment:发布时间" json:"published_at"`                                    // 报告发布时间。
	Metrics     []*ReportMetric `gorm:"foreignKey:ReportID" json:"metrics"`                                  // 报告中包含的指标列表，一对多关系。
}

// ReportMetric 实体代表报告中的一个指标数据。
type ReportMetric struct {
	gorm.Model         // 嵌入gorm.Model。
	ReportID   uint64  `gorm:"not null;index;comment:报告ID" json:"report_id"`        // 关联的报告ID，索引字段。
	Metric     string  `gorm:"type:varchar(128);not null;comment:指标" json:"metric"` // 指标名称。
	Value      float64 `gorm:"type:decimal(16,4);not null;comment:数值" json:"value"` // 指标值。
	Change     float64 `gorm:"type:decimal(10,4);comment:变化率" json:"change"`        // 指标的变化率。
	Trend      string  `gorm:"type:varchar(32);comment:趋势" json:"trend"`            // 指标的趋势，例如“上升”、“下降”。
}

// NewMetric 创建并返回一个新的 Metric 实体实例。
// metricType: 指标类型。
// name: 指标名称。
// value: 指标值。
// granularity: 时间粒度。
func NewMetric(metricType MetricType, name string, value float64, granularity TimeGranularity) *Metric {
	return &Metric{
		MetricType:  metricType,
		MetricName:  name,
		Value:       value,
		Timestamp:   time.Now(), // 记录当前时间作为指标的时间戳。
		Granularity: granularity,
	}
}

// NewDashboard 创建并返回一个新的 Dashboard 实体实例。
// name: 仪表板名称。
// description: 仪表板描述。
// userID: 创建仪表板的用户ID。
func NewDashboard(name, description string, userID uint64) *Dashboard {
	return &Dashboard{
		Name:        name,
		Description: description,
		UserID:      userID,
		IsPublic:    false, // 默认不公开。
	}
}

// AddMetric 将一个 DashboardMetric 添加到仪表板中。
// metric: 待添加的 DashboardMetric 实体。
func (d *Dashboard) AddMetric(metric *DashboardMetric) {
	metric.Position = int32(len(d.Metrics) + 1) // 设置指标的显示位置。
	d.Metrics = append(d.Metrics, metric)       // 将指标添加到Metrics切片。
}

// Publish 发布仪表板，将其设置为公开状态。
func (d *Dashboard) Publish() {
	d.IsPublic = true
}

// NewReport 创建并返回一个新的 Report 实体实例。
// reportNo: 报告编号。
// title, description: 报告标题和描述。
// userID: 创建报告的用户ID。
// reportType: 报告类型。
func NewReport(reportNo, title, description string, userID uint64, reportType string) *Report {
	return &Report{
		ReportNo:    reportNo,
		Title:       title,
		Description: description,
		UserID:      userID,
		ReportType:  reportType,
		Status:      "draft", // 默认状态为草稿。
	}
}

// Publish 发布报告，将其状态设置为“published”，并记录发布时间。
func (r *Report) Publish() {
	r.Status = "published"
	now := time.Now()
	r.PublishedAt = &now
}
