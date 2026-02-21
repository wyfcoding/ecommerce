package domain

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// AnalyticsPeriod 分析周期
type AnalyticsPeriod string

const (
	PeriodRealTime  AnalyticsPeriod = "REAL_TIME" // 实时
	PeriodHourly    AnalyticsPeriod = "HOURLY"    // 每小时
	PeriodDaily     AnalyticsPeriod = "DAILY"     // 每日
	PeriodWeekly    AnalyticsPeriod = "WEEKLY"    // 每周
	PeriodMonthly   AnalyticsPeriod = "MONTHLY"   // 每月
	PeriodQuarterly AnalyticsPeriod = "QUARTERLY" // 每季度
	PeriodYearly    AnalyticsPeriod = "YEARLY"    // 每年
)

// AnalyticsMetric 分析指标
type AnalyticsMetric string

const (
	MetricDeliveryRate      AnalyticsMetric = "DELIVERY_RATE"       // 送达率
	MetricOpenRate          AnalyticsMetric = "OPEN_RATE"           // 打开率
	MetricClickRate         AnalyticsMetric = "CLICK_RATE"          // 点击率
	MetricConversionRate    AnalyticsMetric = "CONVERSION_RATE"     // 转化率
	MetricBounceRate        AnalyticsMetric = "BOUNCE_RATE"         // 退订率
	MetricEngagementRate    AnalyticsMetric = "ENGAGEMENT_RATE"     // 参与率
	MetricRevenuePerMessage AnalyticsMetric = "REVENUE_PER_MESSAGE" // 每消息收入
)

// AnalyticsDimension 分析维度
type AnalyticsDimension string

const (
	DimensionChannel  AnalyticsDimension = "CHANNEL"  // 渠道
	DimensionType     AnalyticsDimension = "TYPE"     // 类型
	DimensionAudience AnalyticsDimension = "AUDIENCE" // 受众
	DimensionCampaign AnalyticsDimension = "CAMPAIGN" // 活动
	DimensionTime     AnalyticsDimension = "TIME"     // 时间
	DimensionRegion   AnalyticsDimension = "REGION"   // 区域
	DimensionDevice   AnalyticsDimension = "DEVICE"   // 设备
)

// AnalyticsReport 分析报告
type AnalyticsReport struct {
	ID          string          `json:"id"`
	ReportNo    string          `json:"report_no"`
	ReportType  string          `json:"report_type"`
	Period      AnalyticsPeriod `json:"period"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	GeneratedAt time.Time       `json:"generated_at"`

	// 报告内容
	Summary         *ReportSummary    `json:"summary"`
	Metrics         []*MetricData     `json:"metrics"`
	Trends          []*TrendData      `json:"trends"`
	Breakdowns      []*BreakdownData  `json:"breakdowns"`
	Insights        []*Insight        `json:"insights"`
	Recommendations []*Recommendation `json:"recommendations"`

	// 元数据
	Format   string                 `json:"format"` // PDF, HTML, CSV, JSON
	Status   string                 `json:"status"` // GENERATING, COMPLETED, FAILED
	Error    string                 `json:"error"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ReportSummary 报告摘要
type ReportSummary struct {
	TotalMessages int `json:"total_messages"`
	Delivered     int `json:"delivered"`
	Opened        int `json:"opened"`
	Clicked       int `json:"clicked"`
	Converted     int `json:"converted"`
	Bounced       int `json:"bounced"`

	// 率指标
	DeliveryRate   float64 `json:"delivery_rate"`
	OpenRate       float64 `json:"open_rate"`
	ClickRate      float64 `json:"click_rate"`
	ConversionRate float64 `json:"conversion_rate"`
	BounceRate     float64 `json:"bounce_rate"`

	// 收入指标
	TotalRevenue      float64 `json:"total_revenue"`
	RevenuePerMessage float64 `json:"revenue_per_message"`

	// 时间指标
	AvgDeliveryTime time.Duration `json:"avg_delivery_time"`
	AvgOpenTime     time.Duration `json:"avg_open_time"`
	AvgClickTime    time.Duration `json:"avg_click_time"`
}

// MetricData 指标数据
type MetricData struct {
	Metric           AnalyticsMetric `json:"metric"`
	Value            float64         `json:"value"`
	PreviousValue    float64         `json:"previous_value"`
	Change           float64         `json:"change"`            // 变化量
	ChangePercentage float64         `json:"change_percentage"` // 变化百分比
	Target           float64         `json:"target"`            // 目标值
	AchievementRate  float64         `json:"achievement_rate"`  // 达成率
	Unit             string          `json:"unit"`              // 单位
}

// TrendData 趋势数据
type TrendData struct {
	Dimension      AnalyticsDimension `json:"dimension"`
	DimensionValue string             `json:"dimension_value"`
	Metric         AnalyticsMetric    `json:"metric"`
	DataPoints     []*TrendPoint      `json:"data_points"`
	TrendDirection string             `json:"trend_direction"` // UP, DOWN, STABLE
	TrendStrength  float64            `json:"trend_strength"`  // 趋势强度
}

// TrendPoint 趋势点
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label"`
}

// BreakdownData 细分数据
type BreakdownData struct {
	Dimension AnalyticsDimension `json:"dimension"`
	Metric    AnalyticsMetric    `json:"metric"`
	Breakdown []*BreakdownItem   `json:"breakdown"`
	Total     float64            `json:"total"`
}

// BreakdownItem 细分项
type BreakdownItem struct {
	DimensionValue string  `json:"dimension_value"`
	Value          float64 `json:"value"`
	Percentage     float64 `json:"percentage"`
	Rank           int     `json:"rank"`
}

// Insight 洞察
type Insight struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // PERFORMANCE, AUDIENCE, TIMING, CONTENT
	Impact      string    `json:"impact"`   // HIGH, MEDIUM, LOW
	Confidence  float64   `json:"confidence"`
	DataPoints  []string  `json:"data_points"`
	GeneratedAt time.Time `json:"generated_at"`
}

// Recommendation 建议
type Recommendation struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"` // OPTIMIZATION, STRATEGY, OPERATIONAL
	Priority       string    `json:"priority"` // HIGH, MEDIUM, LOW
	Effort         string    `json:"effort"`   // HIGH, MEDIUM, LOW
	ExpectedImpact string    `json:"expected_impact"`
	ActionItems    []string  `json:"action_items"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// AnalyticsEngine 分析引擎
type AnalyticsEngine struct {
	notificationRepo NotificationRepository
	reportRepo       AnalyticsRepository
	mu               sync.RWMutex
	config           *AnalyticsConfig
	cache            *AnalyticsCache
}

// AnalyticsConfig 分析配置
type AnalyticsConfig struct {
	CacheTTL         time.Duration `json:"cache_ttl"`
	RealTimeInterval time.Duration `json:"real_time_interval"`
	RetentionPeriod  time.Duration `json:"retention_period"`
	MinDataPoints    int           `json:"min_data_points"`
	ConfidenceLevel  float64       `json:"confidence_level"`
	AnomalyThreshold float64       `json:"anomaly_threshold"`
}

// AnalyticsCache 分析缓存
type AnalyticsCache struct {
	metricsCache map[string]*CachedMetric
	reportsCache map[string]*AnalyticsReport
	mu           sync.RWMutex
}

// CachedMetric 缓存指标
type CachedMetric struct {
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// NewAnalyticsEngine 创建分析引擎
func NewAnalyticsEngine(notificationRepo NotificationRepository,
	reportRepo AnalyticsRepository) *AnalyticsEngine {

	return &AnalyticsEngine{
		notificationRepo: notificationRepo,
		reportRepo:       reportRepo,
		config: &AnalyticsConfig{
			CacheTTL:         5 * time.Minute,
			RealTimeInterval: 1 * time.Minute,
			RetentionPeriod:  365 * 24 * time.Hour,
			MinDataPoints:    10,
			ConfidenceLevel:  0.95,
			AnomalyThreshold: 2.0,
		},
		cache: &AnalyticsCache{
			metricsCache: make(map[string]*CachedMetric),
			reportsCache: make(map[string]*AnalyticsReport),
		},
	}
}

// GenerateReport 生成报告
func (ae *AnalyticsEngine) GenerateReport(ctx context.Context, reportType string,
	period AnalyticsPeriod, startDate, endDate time.Time) (*AnalyticsReport, error) {

	// 检查缓存
	cacheKey := fmt.Sprintf("%s_%s_%s_%s", reportType, period,
		startDate.Format("20060102"), endDate.Format("20060102"))

	ae.cache.mu.RLock()
	cachedReport, exists := ae.cache.reportsCache[cacheKey]
	ae.cache.mu.RUnlock()

	if exists && time.Now().Before(cachedReport.GeneratedAt.Add(ae.config.CacheTTL)) {
		return cachedReport, nil
	}

	// 创建报告
	report := &AnalyticsReport{
		ID:          generateReportID(),
		ReportNo:    generateReportNo(),
		ReportType:  reportType,
		Period:      period,
		StartDate:   startDate,
		EndDate:     endDate,
		GeneratedAt: time.Now(),
		Status:      "GENERATING",
		Format:      "JSON",
		Metadata:    make(map[string]interface{}),
	}

	// 保存报告
	err := ae.reportRepo.SaveReport(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// 异步生成报告内容
	go ae.generateReportContent(ctx, report)

	return report, nil
}

// generateReportContent 生成报告内容
func (ae *AnalyticsEngine) generateReportContent(ctx context.Context, report *AnalyticsReport) {
	defer func() {
		if r := recover(); r != nil {
			report.Status = "FAILED"
			report.Error = fmt.Sprintf("panic: %v", r)
			report.GeneratedAt = time.Now()

			ae.reportRepo.UpdateReport(ctx, report)
		}
	}()

	// 生成摘要
	summary, err := ae.generateSummary(ctx, report.StartDate, report.EndDate)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Summary = summary

	// 生成指标
	metrics, err := ae.generateMetrics(ctx, report.StartDate, report.EndDate)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Metrics = metrics

	// 生成趋势
	trends, err := ae.generateTrends(ctx, report.Period, report.StartDate, report.EndDate)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Trends = trends

	// 生成细分
	breakdowns, err := ae.generateBreakdowns(ctx, report.StartDate, report.EndDate)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Breakdowns = breakdowns

	// 生成洞察
	insights, err := ae.generateInsights(ctx, report)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Insights = insights

	// 生成建议
	recommendations, err := ae.generateRecommendations(ctx, report)
	if err != nil {
		report.Status = "FAILED"
		report.Error = err.Error()
		report.GeneratedAt = time.Now()

		ae.reportRepo.UpdateReport(ctx, report)
		return
	}
	report.Recommendations = recommendations

	// 更新报告状态
	report.Status = "COMPLETED"
	report.GeneratedAt = time.Now()

	err = ae.reportRepo.UpdateReport(ctx, report)
	if err != nil {
		fmt.Printf("Failed to update report: %v\n", err)
	}

	// 更新缓存
	cacheKey := fmt.Sprintf("%s_%s_%s_%s", report.ReportType, report.Period,
		report.StartDate.Format("20060102"), report.EndDate.Format("20060102"))

	ae.cache.mu.Lock()
	ae.cache.reportsCache[cacheKey] = report
	ae.cache.mu.Unlock()
}

// generateSummary 生成摘要
func (ae *AnalyticsEngine) generateSummary(ctx context.Context, startDate, endDate time.Time) (*ReportSummary, error) {
	// 获取通知数据
	notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// 计算摘要
	summary := &ReportSummary{}

	for _, notification := range notifications {
		summary.TotalMessages++

		// 统计送达
		if notification.Status == NotificationStatusRead {
			summary.Delivered++
		}

		// 统计打开
		if notification.Status == NotificationStatusRead {
			summary.Opened++
		}

		// 统计点击
		if notification.Metadata != nil {
			if clicked, ok := notification.Metadata["clicked"].(bool); ok && clicked {
				summary.Clicked++
			}
		}

		// 统计转化
		if notification.Metadata != nil {
			if converted, ok := notification.Metadata["converted"].(bool); ok && converted {
				summary.Converted++
			}
		}

		// 统计退订
		if notification.Status == NotificationStatusDeleted {
			summary.Bounced++
		}

		// 统计收入
		if notification.Metadata != nil {
			if revenue, ok := notification.Metadata["revenue"].(float64); ok {
				summary.TotalRevenue += revenue
			}
		}

		// 统计时间
		if notification.DeliveredAt != nil {
			deliveryTime := notification.DeliveredAt.Sub(notification.CreatedAt)
			summary.AvgDeliveryTime += deliveryTime
		}

		if notification.ReadAt != nil {
			openTime := notification.ReadAt.Sub(*notification.DeliveredAt)
			summary.AvgOpenTime += openTime
		}
	}

	// 计算率指标
	if summary.TotalMessages > 0 {
		summary.DeliveryRate = float64(summary.Delivered) / float64(summary.TotalMessages) * 100
		summary.BounceRate = float64(summary.Bounced) / float64(summary.TotalMessages) * 100
		summary.RevenuePerMessage = summary.TotalRevenue / float64(summary.TotalMessages)
	}

	if summary.Delivered > 0 {
		summary.OpenRate = float64(summary.Opened) / float64(summary.Delivered) * 100
	}

	if summary.Opened > 0 {
		summary.ClickRate = float64(summary.Clicked) / float64(summary.Opened) * 100
	}

	if summary.Clicked > 0 {
		summary.ConversionRate = float64(summary.Converted) / float64(summary.Clicked) * 100
	}

	// 计算平均时间
	if summary.Delivered > 0 {
		summary.AvgDeliveryTime = summary.AvgDeliveryTime / time.Duration(summary.Delivered)
	}

	if summary.Opened > 0 {
		summary.AvgOpenTime = summary.AvgOpenTime / time.Duration(summary.Opened)
	}

	return summary, nil
}

// generateMetrics 生成指标
func (ae *AnalyticsEngine) generateMetrics(ctx context.Context, startDate, endDate time.Time) ([]*MetricData, error) {
	var metrics []*MetricData

	// 获取当前周期数据
	currentNotifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get current notifications: %w", err)
	}

	// 获取上一周期数据
	previousStart := startDate.AddDate(0, 0, -int(endDate.Sub(startDate).Hours()/24))
	previousEnd := startDate
	previousNotifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, previousStart, previousEnd)
	if err != nil {
		// 如果获取上一周期数据失败，仍然生成当前周期指标
		previousNotifications = []*Notification{}
	}

	// 计算送达率
	currentDeliveryRate := ae.calculateDeliveryRate(currentNotifications)
	previousDeliveryRate := ae.calculateDeliveryRate(previousNotifications)

	deliveryRateChange := currentDeliveryRate - previousDeliveryRate
	deliveryRateChangePercentage := 0.0
	if previousDeliveryRate > 0 {
		deliveryRateChangePercentage = deliveryRateChange / previousDeliveryRate * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricDeliveryRate,
		Value:            currentDeliveryRate,
		PreviousValue:    previousDeliveryRate,
		Change:           deliveryRateChange,
		ChangePercentage: deliveryRateChangePercentage,
		Unit:             "%",
	})

	// 计算打开率
	currentOpenRate := ae.calculateOpenRate(currentNotifications)
	previousOpenRate := ae.calculateOpenRate(previousNotifications)

	openRateChange := currentOpenRate - previousOpenRate
	openRateChangePercentage := 0.0
	if previousOpenRate > 0 {
		openRateChangePercentage = openRateChange / previousOpenRate * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricOpenRate,
		Value:            currentOpenRate,
		PreviousValue:    previousOpenRate,
		Change:           openRateChange,
		ChangePercentage: openRateChangePercentage,
		Unit:             "%",
	})

	// 计算点击率
	currentClickRate := ae.calculateClickRate(currentNotifications)
	previousClickRate := ae.calculateClickRate(previousNotifications)

	clickRateChange := currentClickRate - previousClickRate
	clickRateChangePercentage := 0.0
	if previousClickRate > 0 {
		clickRateChangePercentage = clickRateChange / previousClickRate * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricClickRate,
		Value:            currentClickRate,
		PreviousValue:    previousClickRate,
		Change:           clickRateChange,
		ChangePercentage: clickRateChangePercentage,
		Unit:             "%",
	})

	// 计算转化率
	currentConversionRate := ae.calculateConversionRate(currentNotifications)
	previousConversionRate := ae.calculateConversionRate(previousNotifications)

	conversionRateChange := currentConversionRate - previousConversionRate
	conversionRateChangePercentage := 0.0
	if previousConversionRate > 0 {
		conversionRateChangePercentage = conversionRateChange / previousConversionRate * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricConversionRate,
		Value:            currentConversionRate,
		PreviousValue:    previousConversionRate,
		Change:           conversionRateChange,
		ChangePercentage: conversionRateChangePercentage,
		Unit:             "%",
	})

	// 计算每消息收入
	currentRevenuePerMessage := ae.calculateRevenuePerMessage(currentNotifications)
	previousRevenuePerMessage := ae.calculateRevenuePerMessage(previousNotifications)

	revenueChange := currentRevenuePerMessage - previousRevenuePerMessage
	revenueChangePercentage := 0.0
	if previousRevenuePerMessage > 0 {
		revenueChangePercentage = revenueChange / previousRevenuePerMessage * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricRevenuePerMessage,
		Value:            currentRevenuePerMessage,
		PreviousValue:    previousRevenuePerMessage,
		Change:           revenueChange,
		ChangePercentage: revenueChangePercentage,
		Unit:             "元",
	})

	return metrics, nil
}

// calculateDeliveryRate 计算送达率
func (ae *AnalyticsEngine) calculateDeliveryRate(notifications []*Notification) float64 {
	if len(notifications) == 0 {
		return 0
	}

	delivered := 0
	for _, notification := range notifications {
		if notification.Status == NotificationStatusRead {
			delivered++
		}
	}

	return float64(delivered) / float64(len(notifications)) * 100
}

// calculateOpenRate 计算打开率
func (ae *AnalyticsEngine) calculateOpenRate(notifications []*Notification) float64 {
	delivered := 0
	opened := 0

	for _, notification := range notifications {
		if notification.Status == NotificationStatusRead {
			delivered++
		}
		if notification.Status == NotificationStatusRead {
			opened++
		}
	}

	if delivered == 0 {
		return 0
	}

	return float64(opened) / float64(delivered) * 100
}

// calculateClickRate 计算点击率
func (ae *AnalyticsEngine) calculateClickRate(notifications []*Notification) float64 {
	opened := 0
	clicked := 0

	for _, notification := range notifications {
		if notification.Status == NotificationStatusRead {
			opened++
		}
		if notification.Metadata != nil {
			if clickedFlag, ok := notification.Metadata["clicked"].(bool); ok && clickedFlag {
				clicked++
			}
		}
	}

	if opened == 0 {
		return 0
	}

	return float64(clicked) / float64(opened) * 100
}

// calculateConversionRate 计算转化率
func (ae *AnalyticsEngine) calculateConversionRate(notifications []*Notification) float64 {
	clicked := 0
	converted := 0

	for _, notification := range notifications {
		if notification.Metadata != nil {
			if clickedFlag, ok := notification.Metadata["clicked"].(bool); ok && clickedFlag {
				clicked++
			}
			if convertedFlag, ok := notification.Metadata["converted"].(bool); ok && convertedFlag {
				converted++
			}
		}
	}

	if clicked == 0 {
		return 0
	}

	return float64(converted) / float64(clicked) * 100
}

// calculateRevenuePerMessage 计算每消息收入
func (ae *AnalyticsEngine) calculateRevenuePerMessage(notifications []*Notification) float64 {
	if len(notifications) == 0 {
		return 0
	}

	totalRevenue := 0.0
	for _, notification := range notifications {
		if notification.Metadata != nil {
			if revenue, ok := notification.Metadata["revenue"].(float64); ok {
				totalRevenue += revenue
			}
		}
	}

	return totalRevenue / float64(len(notifications))
}

// generateTrends 生成趋势
func (ae *AnalyticsEngine) generateTrends(ctx context.Context, period AnalyticsPeriod, startDate, endDate time.Time) ([]*TrendData, error) {
	var trends []*TrendData

	// 生成送达率趋势
	deliveryTrend, err := ae.generateDeliveryTrend(ctx, period, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate delivery trend: %w", err)
	}
	trends = append(trends, deliveryTrend)

	// 生成打开率趋势
	openTrend, err := ae.generateOpenTrend(ctx, period, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate open trend: %w", err)
	}
	trends = append(trends, openTrend)

	return trends, nil
}

// generateDeliveryTrend 生成送达率趋势
func (ae *AnalyticsEngine) generateDeliveryTrend(ctx context.Context, period AnalyticsPeriod, startDate, endDate time.Time) (*TrendData, error) {
	// 根据周期确定时间间隔
	interval, dateFormat := ae.getTimeInterval(period)

	// 生成数据点
	var dataPoints []*TrendPoint
	currentDate := startDate

	for currentDate.Before(endDate) {
		nextDate := currentDate.Add(interval)
		if nextDate.After(endDate) {
			nextDate = endDate
		}

		// 获取该时间段的通知
		notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, currentDate, nextDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get notifications: %w", err)
		}

		// 计算送达率
		deliveryRate := ae.calculateDeliveryRate(notifications)

		dataPoint := &TrendPoint{
			Timestamp: currentDate,
			Value:     deliveryRate,
			Label:     currentDate.Format(dateFormat),
		}
		dataPoints = append(dataPoints, dataPoint)

		currentDate = nextDate
	}

	// 计算趋势
	direction, strength := ae.calculateTrend(dataPoints)

	trend := &TrendData{
		Dimension:      DimensionTime,
		DimensionValue: "时间",
		Metric:         MetricDeliveryRate,
		DataPoints:     dataPoints,
		TrendDirection: direction,
		TrendStrength:  strength,
	}

	return trend, nil
}

// generateOpenTrend 生成打开率趋势
func (ae *AnalyticsEngine) generateOpenTrend(ctx context.Context, period AnalyticsPeriod, startDate, endDate time.Time) (*TrendData, error) {
	// 根据周期确定时间间隔
	interval, dateFormat := ae.getTimeInterval(period)

	// 生成数据点
	var dataPoints []*TrendPoint
	currentDate := startDate

	for currentDate.Before(endDate) {
		nextDate := currentDate.Add(interval)
		if nextDate.After(endDate) {
			nextDate = endDate
		}

		// 获取该时间段的通知
		notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, currentDate, nextDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get notifications: %w", err)
		}

		// 计算打开率
		openRate := ae.calculateOpenRate(notifications)

		dataPoint := &TrendPoint{
			Timestamp: currentDate,
			Value:     openRate,
			Label:     currentDate.Format(dateFormat),
		}
		dataPoints = append(dataPoints, dataPoint)

		currentDate = nextDate
	}

	// 计算趋势
	direction, strength := ae.calculateTrend(dataPoints)

	trend := &TrendData{
		Dimension:      DimensionTime,
		DimensionValue: "时间",
		Metric:         MetricOpenRate,
		DataPoints:     dataPoints,
		TrendDirection: direction,
		TrendStrength:  strength,
	}

	return trend, nil
}

// getTimeInterval 获取时间间隔
func (ae *AnalyticsEngine) getTimeInterval(period AnalyticsPeriod) (time.Duration, string) {
	switch period {
	case PeriodHourly:
		return 1 * time.Hour, "15:04"
	case PeriodDaily:
		return 24 * time.Hour, "01-02"
	case PeriodWeekly:
		return 7 * 24 * time.Hour, "01-02"
	case PeriodMonthly:
		return 30 * 24 * time.Hour, "2006-01"
	case PeriodQuarterly:
		return 90 * 24 * time.Hour, "2006-Q1"
	case PeriodYearly:
		return 365 * 24 * time.Hour, "2006"
	default:
		return 24 * time.Hour, "01-02"
	}
}

// calculateTrend 计算趋势
func (ae *AnalyticsEngine) calculateTrend(dataPoints []*TrendPoint) (string, float64) {
	if len(dataPoints) < 2 {
		return "STABLE", 0
	}

	// 计算线性回归
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(dataPoints))

	for i, point := range dataPoints {
		x := float64(i)
		y := point.Value

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// 确定趋势方向
	var direction string
	if slope > 0.1 {
		direction = "UP"
	} else if slope < -0.1 {
		direction = "DOWN"
	} else {
		direction = "STABLE"
	}

	// 计算趋势强度（斜率的绝对值）
	strength := math.Abs(slope)

	return direction, strength
}

// generateBreakdowns 生成细分
func (ae *AnalyticsEngine) generateBreakdowns(ctx context.Context, startDate, endDate time.Time) ([]*BreakdownData, error) {
	var breakdowns []*BreakdownData

	// 按渠道细分
	channelBreakdown, err := ae.generateChannelBreakdown(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate channel breakdown: %w", err)
	}
	breakdowns = append(breakdowns, channelBreakdown)

	// 按类型细分
	typeBreakdown, err := ae.generateTypeBreakdown(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate type breakdown: %w", err)
	}
	breakdowns = append(breakdowns, typeBreakdown)

	return breakdowns, nil
}

// generateChannelBreakdown 生成渠道细分
func (ae *AnalyticsEngine) generateChannelBreakdown(ctx context.Context, startDate, endDate time.Time) (*BreakdownData, error) {
	// 获取通知数据
	notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// 按渠道统计
	channelStats := make(map[string]float64)

	for _, notification := range notifications {
		channel := notification.Channel
		if channel == "" {
			channel = "UNKNOWN"
		}

		// 统计送达率
		if notification.Status == NotificationStatusDelivered || notification.Status == NotificationStatusRead {
			channelStats[string(channel)]++
		}
	}

	// 转换为细分项
	var breakdown []*BreakdownItem
	total := float64(len(notifications))

	for channel, count := range channelStats {
		percentage := 0.0
		if total > 0 {
			percentage = count / total * 100
		}

		breakdown = append(breakdown, &BreakdownItem{
			DimensionValue: channel,
			Value:          count,
			Percentage:     percentage,
		})
	}

	// 按值排序
	for i := 0; i < len(breakdown)-1; i++ {
		for j := i + 1; j < len(breakdown); j++ {
			if breakdown[i].Value < breakdown[j].Value {
				breakdown[i], breakdown[j] = breakdown[j], breakdown[i]
			}
		}
	}

	// 更新排名
	for i, item := range breakdown {
		item.Rank = i + 1
	}

	breakdownData := &BreakdownData{
		Dimension: DimensionChannel,
		Metric:    MetricDeliveryRate,
		Breakdown: breakdown,
		Total:     total,
	}

	return breakdownData, nil
}

// generateTypeBreakdown 生成类型细分
func (ae *AnalyticsEngine) generateTypeBreakdown(ctx context.Context, startDate, endDate time.Time) (*BreakdownData, error) {
	// 获取通知数据
	notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// 按类型统计
	typeStats := make(map[string]float64)

	for _, notification := range notifications {
		notificationType := string(notification.Type)
		if notificationType == "" {
			notificationType = "UNKNOWN"
		}

		// 统计打开率
		if notification.Status == NotificationStatusRead {
			typeStats[notificationType]++
		}
	}

	// 转换为细分项
	var breakdown []*BreakdownItem
	total := float64(len(notifications))

	for notificationType, count := range typeStats {
		percentage := 0.0
		if total > 0 {
			percentage = count / total * 100
		}

		breakdown = append(breakdown, &BreakdownItem{
			DimensionValue: notificationType,
			Value:          count,
			Percentage:     percentage,
		})
	}

	// 按值排序
	for i := 0; i < len(breakdown)-1; i++ {
		for j := i + 1; j < len(breakdown); j++ {
			if breakdown[i].Value < breakdown[j].Value {
				breakdown[i], breakdown[j] = breakdown[j], breakdown[i]
			}
		}
	}

	// 更新排名
	for i, item := range breakdown {
		item.Rank = i + 1
	}

	breakdownData := &BreakdownData{
		Dimension: DimensionType,
		Metric:    MetricOpenRate,
		Breakdown: breakdown,
		Total:     total,
	}

	return breakdownData, nil
}

// generateInsights 生成洞察
func (ae *AnalyticsEngine) generateInsights(ctx context.Context, report *AnalyticsReport) ([]*Insight, error) {
	var insights []*Insight

	// 基于报告数据生成洞察
	if report.Summary != nil {
		// 送达率洞察
		if report.Summary.DeliveryRate < 90 {
			insights = append(insights, &Insight{
				ID:          generateInsightID(),
				Title:       "低送达率预警",
				Description: fmt.Sprintf("通知送达率仅为 %.1f%%，低于90%%的标准", report.Summary.DeliveryRate),
				Category:    "PERFORMANCE",
				Impact:      "HIGH",
				Confidence:  0.9,
				DataPoints:  []string{"delivery_rate"},
				GeneratedAt: time.Now(),
			})
		}

		// 打开率洞察
		if report.Summary.OpenRate < 20 {
			insights = append(insights, &Insight{
				ID:          generateInsightID(),
				Title:       "低打开率发现",
				Description: fmt.Sprintf("通知打开率仅为 %.1f%%，内容吸引力不足", report.Summary.OpenRate),
				Category:    "CONTENT",
				Impact:      "MEDIUM",
				Confidence:  0.8,
				DataPoints:  []string{"open_rate"},
				GeneratedAt: time.Now(),
			})
		}

		// 转化率洞察
		if report.Summary.ConversionRate > 10 {
			insights = append(insights, &Insight{
				ID:          generateInsightID(),
				Title:       "高转化率表现",
				Description: fmt.Sprintf("通知转化率达到 %.1f%%，表现优秀", report.Summary.ConversionRate),
				Category:    "PERFORMANCE",
				Impact:      "MEDIUM",
				Confidence:  0.85,
				DataPoints:  []string{"conversion_rate"},
				GeneratedAt: time.Now(),
			})
		}
	}

	return insights, nil
}

// generateRecommendations 生成建议
func (ae *AnalyticsEngine) generateRecommendations(ctx context.Context, report *AnalyticsReport) ([]*Recommendation, error) {
	var recommendations []*Recommendation

	// 基于洞察生成建议
	if report.Summary != nil && report.Summary.DeliveryRate < 90 {
		recommendations = append(recommendations, &Recommendation{
			ID:             generateRecommendationID(),
			Title:          "优化通知送达率",
			Description:    "分析低送达率的原因，优化发送策略和渠道配置",
			Category:       "OPTIMIZATION",
			Priority:       "HIGH",
			Effort:         "MEDIUM",
			ExpectedImpact: "提升送达率5-10%",
			ActionItems: []string{
				"检查发送服务配置",
				"分析退订和失败原因",
				"优化发送时间选择",
				"测试不同发送渠道",
			},
			GeneratedAt: time.Now(),
		})
	}

	if report.Summary != nil && report.Summary.OpenRate < 20 {
		recommendations = append(recommendations, &Recommendation{
			ID:             generateRecommendationID(),
			Title:          "提升通知内容吸引力",
			Description:    "优化通知标题和内容，提高用户打开率",
			Category:       "CONTENT",
			Priority:       "MEDIUM",
			Effort:         "LOW",
			ExpectedImpact: "提升打开率3-5%",
			ActionItems: []string{
				"A/B测试不同标题",
				"优化内容格式和长度",
				"添加个性化元素",
				"测试不同发送时间",
			},
			GeneratedAt: time.Now(),
		})
	}

	return recommendations, nil
}

// GetRealTimeMetrics 获取实时指标
func (ae *AnalyticsEngine) GetRealTimeMetrics(ctx context.Context) (map[string]interface{}, error) {
	// 检查缓存
	cacheKey := "realtime_metrics"

	ae.cache.mu.RLock()
	cachedMetric, exists := ae.cache.metricsCache[cacheKey]
	ae.cache.mu.RUnlock()

	if exists && time.Now().Before(cachedMetric.ExpiresAt) {
		return cachedMetric.Value.(map[string]interface{}), nil
	}

	// 计算实时指标
	metrics := make(map[string]interface{})

	// 获取最近1小时的通知
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	notifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, oneHourAgo, time.Now())
	if err == nil {
		delivered := 0
		opened := 0

		for _, notification := range notifications {
			if notification.Status == NotificationStatusDelivered || notification.Status == NotificationStatusRead {
				delivered++
			}
			if notification.Status == NotificationStatusRead {
				opened++
			}
		}

		deliveryRate := 0.0
		if len(notifications) > 0 {
			deliveryRate = float64(delivered) / float64(len(notifications)) * 100
		}

		openRate := 0.0
		if delivered > 0 {
			openRate = float64(opened) / float64(delivered) * 100
		}

		metrics["hourly_messages"] = len(notifications)
		metrics["hourly_delivery_rate"] = deliveryRate
		metrics["hourly_open_rate"] = openRate
	}

	// 获取今日通知
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayNotifications, err := ae.notificationRepo.GetNotificationsByDateRange(ctx, todayStart, time.Now())
	if err == nil {
		metrics["daily_messages"] = len(todayNotifications)
	}

	// 更新缓存
	ae.cache.mu.Lock()
	ae.cache.metricsCache[cacheKey] = &CachedMetric{
		Value:     metrics,
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(ae.config.RealTimeInterval),
	}
	ae.cache.mu.Unlock()

	return metrics, nil
}

// Helper functions

func generateReportID() string {
	return fmt.Sprintf("REPORT_%d", time.Now().UnixNano())
}

func generateReportNo() string {
	return fmt.Sprintf("RPT%d", time.Now().UnixNano())
}

func generateInsightID() string {
	return fmt.Sprintf("INSIGHT_%d", time.Now().UnixNano())
}

func generateRecommendationID() string {
	return fmt.Sprintf("RECOMMEND_%d", time.Now().UnixNano())
}

// Repository interfaces

type AnalyticsRepository interface {
	SaveReport(ctx context.Context, report *AnalyticsReport) error
	GetReport(ctx context.Context, reportID string) (*AnalyticsReport, error)
	GetReportsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*AnalyticsReport, error)
	UpdateReport(ctx context.Context, report *AnalyticsReport) error
	DeleteReport(ctx context.Context, reportID string) error
}
