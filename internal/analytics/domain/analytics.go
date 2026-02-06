package domain

import "time"

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
	ID           uint            `json:"id"`            // 主键ID
	CreatedAt    time.Time       `json:"created_at"`    // 创建时间
	UpdatedAt    time.Time       `json:"updated_at"`    // 更新时间
	MetricType   MetricType      `json:"metric_type"`   // 指标类型
	MetricName   string          `json:"metric_name"`   // 指标名称
	Value        float64         `json:"value"`         // 指标的具体数值
	Timestamp    time.Time       `json:"timestamp"`     // 指标记录的时间戳
	Granularity  TimeGranularity `json:"granularity"`   // 时间粒度
	Dimension    string          `json:"dimension"`     // 维度名称
	DimensionVal string          `json:"dimension_val"` // 维度值
}

// Dashboard 实体代表一个聚合根，用于组织和展示多个指标的图表。
// 用户可以创建和定制自己的仪表板来监控关键业务数据。
type Dashboard struct {
	ID          uint               `json:"id"`          // 主键ID
	CreatedAt   time.Time          `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time          `json:"updated_at"`  // 更新时间
	Name        string             `json:"name"`        // 仪表板名称
	Description string             `json:"description"` // 仪表板描述
	UserID      uint64             `json:"user_id"`     // 创建仪表板的用户ID
	IsPublic    bool               `json:"is_public"`   // 是否公开
	Metrics     []*DashboardMetric `json:"metrics"`     // 指标图表列表
	Filters     []*DashboardFilter `json:"filters"`     // 过滤器列表
}

// DashboardMetric 实体代表仪表板上显示的一个指标图表。
type DashboardMetric struct {
	ID          uint       `json:"id"`           // 主键ID
	CreatedAt   time.Time  `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time  `json:"updated_at"`   // 更新时间
	DashboardID uint64     `json:"dashboard_id"` // 关联的仪表板ID
	MetricType  MetricType `json:"metric_type"`  // 指标类型
	Title       string     `json:"title"`        // 图表标题
	ChartType   string     `json:"chart_type"`   // 图表类型
	Position    int32      `json:"position"`     // 显示位置或顺序
}

// DashboardFilter 实体代表仪表板上的一个过滤器。
type DashboardFilter struct {
	ID          uint      `json:"id"`           // 主键ID
	CreatedAt   time.Time `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`   // 更新时间
	DashboardID uint64    `json:"dashboard_id"` // 关联的仪表板ID
	FilterName  string    `json:"filter_name"`  // 过滤器名称
	FilterType  string    `json:"filter_type"`  // 过滤器类型
	FilterValue string    `json:"filter_value"` // 过滤器值
}

// Report 实体代表一个聚合根，用于存储生成的数据分析报告。
// 报告可以是定期生成的，也可以是按需生成的，通常包含对业务数据的深入分析。
type Report struct {
	ID          uint            `json:"id"`           // 主键ID
	CreatedAt   time.Time       `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time       `json:"updated_at"`   // 更新时间
	ReportNo    string          `json:"report_no"`    // 报告编号
	Title       string          `json:"title"`        // 报告标题
	Description string          `json:"description"`  // 报告描述
	UserID      uint64          `json:"user_id"`      // 创建报告的用户ID
	ReportType  string          `json:"report_type"`  // 报告类型
	StartDate   time.Time       `json:"start_date"`   // 报告开始日期
	EndDate     time.Time       `json:"end_date"`     // 报告结束日期
	Status      string          `json:"status"`       // 报告状态
	Content     string          `json:"content"`      // 报告内容
	PublishedAt *time.Time      `json:"published_at"` // 发布时间
	Metrics     []*ReportMetric `json:"metrics"`      // 报告指标列表
}

// ReportMetric 实体代表报告中的一个指标数据。
type ReportMetric struct {
	ID        uint      `json:"id"`         // 主键ID
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	ReportID  uint64    `json:"report_id"`  // 关联的报告ID
	Metric    string    `json:"metric"`     // 指标名称
	Value     float64   `json:"value"`      // 指标值
	Change    float64   `json:"change"`     // 变化率
	Trend     string    `json:"trend"`      // 趋势
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
		Metrics:     []*DashboardMetric{},
		Filters:     []*DashboardFilter{},
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
		Metrics:     []*ReportMetric{},
	}
}

// Publish 发布报告，将其状态设置为“published”，并记录发布时间。
func (r *Report) Publish() {
	r.Status = "published"
	now := time.Now()
	r.PublishedAt = &now
}
