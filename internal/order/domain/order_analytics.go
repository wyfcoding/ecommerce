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
	PeriodDaily     AnalyticsPeriod = "DAILY"     // 每日
	PeriodWeekly    AnalyticsPeriod = "WEEKLY"    // 每周
	PeriodMonthly   AnalyticsPeriod = "MONTHLY"   // 每月
	PeriodQuarterly AnalyticsPeriod = "QUARTERLY" // 每季度
	PeriodYearly    AnalyticsPeriod = "YEARLY"    // 每年
)

// AnalyticsMetric 分析指标
type AnalyticsMetric string

const (
	MetricOrderCount       AnalyticsMetric = "ORDER_COUNT"       // 订单数量
	MetricOrderValue       AnalyticsMetric = "ORDER_VALUE"       // 订单金额
	MetricAvgOrderValue    AnalyticsMetric = "AVG_ORDER_VALUE"   // 平均订单金额
	MetricConversionRate   AnalyticsMetric = "CONVERSION_RATE"   // 转化率
	MetricCancellationRate AnalyticsMetric = "CANCELLATION_RATE" // 取消率
	MetricRefundRate       AnalyticsMetric = "REFUND_RATE"       // 退款率
	MetricCustomerLTV      AnalyticsMetric = "CUSTOMER_LTV"      // 客户生命周期价值
	MetricRepeatRate       AnalyticsMetric = "REPEAT_RATE"       // 复购率
)

// AnalyticsDimension 分析维度
type AnalyticsDimension string

const (
	DimensionTime          AnalyticsDimension = "TIME"           // 时间
	DimensionProduct       AnalyticsDimension = "PRODUCT"        // 产品
	DimensionCategory      AnalyticsDimension = "CATEGORY"       // 类别
	DimensionCustomer      AnalyticsDimension = "CUSTOMER"       // 客户
	DimensionRegion        AnalyticsDimension = "REGION"         // 区域
	DimensionChannel       AnalyticsDimension = "CHANNEL"        // 渠道
	DimensionPaymentMethod AnalyticsDimension = "PAYMENT_METHOD" // 支付方式
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
	TotalOrders      int     `json:"total_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgOrderValue    float64 `json:"avg_order_value"`
	TotalCustomers   int     `json:"total_customers"`
	NewCustomers     int     `json:"new_customers"`
	RepeatCustomers  int     `json:"repeat_customers"`
	CancellationRate float64 `json:"cancellation_rate"`
	RefundRate       float64 `json:"refund_rate"`
	ConversionRate   float64 `json:"conversion_rate"`
	TopProduct       string  `json:"top_product"`
	TopCategory      string  `json:"top_category"`
	TopRegion        string  `json:"top_region"`
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
	Category    string    `json:"category"` // SALES, CUSTOMER, OPERATIONAL, FINANCIAL
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
	Category       string    `json:"category"`
	Priority       string    `json:"priority"` // HIGH, MEDIUM, LOW
	Effort         string    `json:"effort"`   // HIGH, MEDIUM, LOW
	ExpectedImpact string    `json:"expected_impact"`
	ActionItems    []string  `json:"action_items"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// AnalyticsEngine 分析引擎
type AnalyticsEngine struct {
	orderRepo    AnalyticsOrderRepository
	customerRepo CustomerRepository
	productRepo  ProductRepository
	reportRepo   AnalyticsRepository
	mu           sync.RWMutex
	config       *AnalyticsConfig
	cache        *AnalyticsCache
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
func NewAnalyticsEngine(orderRepo AnalyticsOrderRepository, customerRepo CustomerRepository,
	productRepo ProductRepository, reportRepo AnalyticsRepository) *AnalyticsEngine {

	return &AnalyticsEngine{
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
		productRepo:  productRepo,
		reportRepo:   reportRepo,
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
	cacheKey := fmt.Sprintf("%s_%s_%s_%s", reportType, period, startDate.Format("20060102"), endDate.Format("20060102"))

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
	// 获取订单数据
	orders, err := ae.orderRepo.GetOrdersByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	// 获取客户数据
	customers, err := ae.customerRepo.GetCustomersByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	// 计算摘要
	summary := &ReportSummary{}

	// 订单统计
	for _, order := range orders {
		summary.TotalOrders++
		summary.TotalRevenue += float64(order.TotalAmount) / 100.0 // 转换为元

		// 统计取消和退款
		if order.Status.String() == "CANCELLED" {
			summary.CancellationRate++
		} else if order.Status.String() == "REFUNDED" {
			summary.RefundRate++
		}
	}

	// 计算平均值
	if summary.TotalOrders > 0 {
		summary.AvgOrderValue = summary.TotalRevenue / float64(summary.TotalOrders)
		summary.CancellationRate = summary.CancellationRate / float64(summary.TotalOrders) * 100
		summary.RefundRate = summary.RefundRate / float64(summary.TotalOrders) * 100
	}

	// 客户统计
	customerSet := make(map[string]bool)
	newCustomerSet := make(map[string]bool)
	repeatCustomerSet := make(map[string]bool)

	for _, customer := range customers {
		customerSet[customer.ID] = true

		// 检查是否为新客户
		if customer.CreatedAt.After(startDate) && customer.CreatedAt.Before(endDate) {
			newCustomerSet[customer.ID] = true
			summary.NewCustomers++
		}

		// 检查是否为复购客户
		orderCount, err := ae.orderRepo.GetOrderCountByCustomer(ctx, customer.ID, startDate, endDate)
		if err == nil && orderCount > 1 {
			repeatCustomerSet[customer.ID] = true
			summary.RepeatCustomers++
		}
	}

	summary.TotalCustomers = len(customerSet)

	// 计算转化率（简化）
	if len(customerSet) > 0 {
		summary.ConversionRate = float64(summary.TotalOrders) / float64(len(customerSet)) * 100
	}

	// 获取热门产品、类别、区域
	topProduct, topCategory, topRegion := ae.getTopItems(ctx, orders)
	summary.TopProduct = topProduct
	summary.TopCategory = topCategory
	summary.TopRegion = topRegion

	return summary, nil
}

// getTopItems 获取热门项
func (ae *AnalyticsEngine) getTopItems(ctx context.Context, orders []*Order) (string, string, string) {
	productCount := make(map[uint64]int)
	categoryCount := make(map[string]int)
	regionCount := make(map[string]int)

	for _, order := range orders {
		for _, item := range order.Items {
			productCount[item.ProductID]++

			// 获取产品类别
			product, err := ae.productRepo.GetProduct(ctx, fmt.Sprintf("%d", item.ProductID))
			if err == nil && product.Category != "" {
				categoryCount[product.Category]++
			}
		}

		if order.ShippingAddress != nil && order.ShippingAddress.Province != "" {
			regionCount[order.ShippingAddress.Province]++
		}
	}

	// 找出最大值
	topProduct := findMaxKeyUint64(productCount)
	topCategory := findMaxKeyString(categoryCount)
	topRegion := findMaxKeyString(regionCount)

	return topProduct, topCategory, topRegion
}

// findMaxKeyUint64 找出最大值对应的键（uint64类型）
func findMaxKeyUint64(m map[uint64]int) string {
	var maxKey string
	var maxValue int

	for key, value := range m {
		if value > maxValue {
			maxValue = value
			maxKey = fmt.Sprintf("产品%d", key)
		}
	}

	return maxKey
}

// findMaxKeyString 找出最大值对应的键（string类型）
func findMaxKeyString(m map[string]int) string {
	var maxKey string
	var maxValue int

	for key, value := range m {
		if value > maxValue {
			maxValue = value
			maxKey = key
		}
	}

	return maxKey
}

// generateMetrics 生成指标
func (ae *AnalyticsEngine) generateMetrics(ctx context.Context, startDate, endDate time.Time) ([]*MetricData, error) {
	var metrics []*MetricData

	// 获取当前周期数据
	currentOrders, err := ae.orderRepo.GetOrdersByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get current orders: %w", err)
	}

	// 获取上一周期数据
	previousStart := startDate.AddDate(0, 0, -int(endDate.Sub(startDate).Hours()/24))
	previousEnd := startDate
	previousOrders, err := ae.orderRepo.GetOrdersByDateRange(ctx, previousStart, previousEnd)
	if err != nil {
		// 如果获取上一周期数据失败，仍然生成当前周期指标
		previousOrders = []*Order{}
	}

	// 计算订单数量
	orderCount := len(currentOrders)
	previousOrderCount := len(previousOrders)
	orderCountChange := float64(orderCount - previousOrderCount)
	orderCountChangePercentage := 0.0
	if previousOrderCount > 0 {
		orderCountChangePercentage = orderCountChange / float64(previousOrderCount) * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricOrderCount,
		Value:            float64(orderCount),
		PreviousValue:    float64(previousOrderCount),
		Change:           orderCountChange,
		ChangePercentage: orderCountChangePercentage,
		Unit:             "个",
	})

	// 计算订单金额
	var orderValue, previousOrderValue float64
	for _, order := range currentOrders {
		orderValue += float64(order.TotalAmount) / 100.0 // 转换为元
	}
	for _, order := range previousOrders {
		previousOrderValue += float64(order.TotalAmount) / 100.0 // 转换为元
	}

	orderValueChange := orderValue - previousOrderValue
	orderValueChangePercentage := 0.0
	if previousOrderValue > 0 {
		orderValueChangePercentage = orderValueChange / previousOrderValue * 100
	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricOrderValue,
		Value:            orderValue,
		PreviousValue:    previousOrderValue,
		Change:           orderValueChange,
		ChangePercentage: orderValueChangePercentage,
		Unit:             "元",
	})

	// 计算平均订单金额
	avgOrderValue := 0.0
	if orderCount > 0 {
		avgOrderValue = orderValue / float64(orderCount)
	}

	previousAvgOrderValue := 0.0
	if previousOrderCount > 0 {
		previousAvgOrderValue = previousOrderValue / float64(previousOrderCount)
	}

	avgOrderValueChange := avgOrderValue - previousAvgOrderValue
	avgOrderValueChangePercentage := 0.0
	if previousAvgOrderValue > 0 {
		avgOrderValueChangePercentage = avgOrderValueChange / previousAvgOrderValue * 100

	}

	metrics = append(metrics, &MetricData{
		Metric:           MetricAvgOrderValue,
		Value:            avgOrderValue,
		PreviousValue:    previousAvgOrderValue,
		Change:           avgOrderValueChange,
		ChangePercentage: avgOrderValueChangePercentage,
		Unit:             "元",
	})

	// 可以添加更多指标...

	return metrics, nil
}

// generateTrends 生成趋势
func (ae *AnalyticsEngine) generateTrends(ctx context.Context, period AnalyticsPeriod, startDate, endDate time.Time) ([]*TrendData, error) {
	var trends []*TrendData

	// 生成时间趋势
	timeTrend, err := ae.generateTimeTrend(ctx, period, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate time trend: %w", err)
	}
	trends = append(trends, timeTrend)

	return trends, nil
}

// generateTimeTrend 生成时间趋势
func (ae *AnalyticsEngine) generateTimeTrend(ctx context.Context, period AnalyticsPeriod, startDate, endDate time.Time) (*TrendData, error) {
	// 根据周期确定时间间隔
	var interval time.Duration
	var dateFormat string

	switch period {
	case PeriodDaily:
		interval = 24 * time.Hour
		dateFormat = "01-02"
	case PeriodWeekly:
		interval = 7 * 24 * time.Hour
		dateFormat = "01-02"
	case PeriodMonthly:
		interval = 30 * 24 * time.Hour
		dateFormat = "2006-01"
	case PeriodQuarterly:
		interval = 90 * 24 * time.Hour
		dateFormat = "2006-Q1"
	case PeriodYearly:
		interval = 365 * 24 * time.Hour
		dateFormat = "2006"
	default:
		interval = 24 * time.Hour
		dateFormat = "01-02"
	}

	// 生成数据点
	var dataPoints []*TrendPoint
	currentDate := startDate

	for currentDate.Before(endDate) {
		nextDate := currentDate.Add(interval)
		if nextDate.After(endDate) {
			nextDate = endDate
		}

		// 获取该时间段的订单
		orders, err := ae.orderRepo.GetOrdersByDateRange(ctx, currentDate, nextDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get orders for date range: %w", err)
		}

		// 计算订单金额
		var orderValue float64
		for _, order := range orders {
			orderValue += float64(order.TotalAmount) / 100.0 // 转换为元
		}

		dataPoint := &TrendPoint{
			Timestamp: currentDate,
			Value:     orderValue,
			Label:     currentDate.Format(dateFormat),
		}
		dataPoints = append(dataPoints, dataPoint)

		currentDate = nextDate
	}

	// 计算趋势方向和强度
	direction, strength := ae.calculateTrend(dataPoints)

	trend := &TrendData{
		Dimension:      DimensionTime,
		DimensionValue: "时间",
		Metric:         MetricOrderValue,
		DataPoints:     dataPoints,
		TrendDirection: direction,
		TrendStrength:  strength,
	}

	return trend, nil
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

	// 产品细分
	productBreakdown, err := ae.generateProductBreakdown(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate product breakdown: %w", err)
	}
	breakdowns = append(breakdowns, productBreakdown)

	// 区域细分
	regionBreakdown, err := ae.generateRegionBreakdown(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate region breakdown: %w", err)
	}
	breakdowns = append(breakdowns, regionBreakdown)

	return breakdowns, nil
}

// generateProductBreakdown 生成产品细分
func (ae *AnalyticsEngine) generateProductBreakdown(ctx context.Context, startDate, endDate time.Time) (*BreakdownData, error) {
	// 获取订单
	orders, err := ae.orderRepo.GetOrdersByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	// 统计产品销售额
	productSales := make(map[uint64]float64)

	for _, order := range orders {
		for _, item := range order.Items {
			productSales[item.ProductID] += float64(item.Price * int64(item.Quantity))
		}
	}

	// 转换为细分项
	var breakdown []*BreakdownItem
	var totalSales float64

	for productID, sales := range productSales {
		totalSales += sales

		// 获取产品名称
		productName := fmt.Sprintf("产品%d", productID)
		product, err := ae.productRepo.GetProduct(ctx, fmt.Sprintf("%d", productID))
		if err == nil {
			productName = product.Name
		}

		breakdown = append(breakdown, &BreakdownItem{
			DimensionValue: productName,
			Value:          sales,
		})
	}

	// 计算百分比和排名
	for i, item := range breakdown {
		if totalSales > 0 {
			item.Percentage = item.Value / totalSales * 100
		}
		item.Rank = i + 1
	}

	// 按销售额排序
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
		Dimension: DimensionProduct,
		Metric:    MetricOrderValue,
		Breakdown: breakdown,
		Total:     totalSales,
	}

	return breakdownData, nil
}

// generateRegionBreakdown 生成区域细分
func (ae *AnalyticsEngine) generateRegionBreakdown(ctx context.Context, startDate, endDate time.Time) (*BreakdownData, error) {
	// 获取订单
	orders, err := ae.orderRepo.GetOrdersByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	// 统计区域销售额
	regionSales := make(map[string]float64)

	for _, order := range orders {
		if order.ShippingAddress != nil && order.ShippingAddress.Province != "" {
			regionSales[order.ShippingAddress.Province] += float64(order.TotalAmount)
		} else {
			regionSales["未知"] += float64(order.TotalAmount)
		}
	}

	// 转换为细分项
	var breakdown []*BreakdownItem
	var totalSales float64

	for region, sales := range regionSales {
		totalSales += sales
		breakdown = append(breakdown, &BreakdownItem{
			DimensionValue: region,
			Value:          sales,
		})
	}

	// 计算百分比和排名
	for i, item := range breakdown {
		if totalSales > 0 {
			item.Percentage = item.Value / totalSales * 100
		}
		item.Rank = i + 1
	}

	// 按销售额排序
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
		Dimension: DimensionRegion,
		Metric:    MetricOrderValue,
		Breakdown: breakdown,
		Total:     totalSales,
	}

	return breakdownData, nil
}

// generateInsights 生成洞察
func (ae *AnalyticsEngine) generateInsights(ctx context.Context, report *AnalyticsReport) ([]*Insight, error) {
	var insights []*Insight

	// 基于报告数据生成洞察
	// 这里可以添加复杂的分析逻辑

	// 示例洞察
	if report.Summary != nil {
		if report.Summary.CancellationRate > 10 {
			insights = append(insights, &Insight{
				ID:          generateInsightID(),
				Title:       "高取消率预警",
				Description: fmt.Sprintf("订单取消率达到 %.1f%%，超过10%%的警戒线", report.Summary.CancellationRate),
				Category:    "OPERATIONAL",
				Impact:      "HIGH",
				Confidence:  0.9,
				DataPoints:  []string{"cancellation_rate"},
				GeneratedAt: time.Now(),
			})
		}

		if report.Summary.AvgOrderValue > 1000 {
			insights = append(insights, &Insight{
				ID:          generateInsightID(),
				Title:       "高客单价表现",
				Description: fmt.Sprintf("平均订单金额达到 %.2f元，表现优秀", report.Summary.AvgOrderValue),
				Category:    "SALES",
				Impact:      "MEDIUM",
				Confidence:  0.8,
				DataPoints:  []string{"avg_order_value"},
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
	// 这里可以添加复杂的建议生成逻辑

	// 示例建议
	if report.Summary != nil && report.Summary.CancellationRate > 10 {
		recommendations = append(recommendations, &Recommendation{
			ID:             generateRecommendationID(),
			Title:          "优化订单取消流程",
			Description:    "分析高取消率的原因，优化购物流程和客户服务",
			Category:       "OPERATIONAL",
			Priority:       "HIGH",
			Effort:         "MEDIUM",
			ExpectedImpact: "降低取消率5-10%",
			ActionItems: []string{
				"分析取消订单的原因分类",
				"优化购物车和结账流程",
				"加强客户服务响应",
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

	// 获取最近1小时的订单
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	orders, err := ae.orderRepo.GetOrdersByDateRange(ctx, oneHourAgo, time.Now())
	if err == nil {
		var hourRevenue float64
		for _, order := range orders {
			hourRevenue += float64(order.TotalAmount) / 100.0 // 转换为元
		}

		metrics["hourly_orders"] = len(orders)
		metrics["hourly_revenue"] = hourRevenue
	}

	// 获取今日订单
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayOrders, err := ae.orderRepo.GetOrdersByDateRange(ctx, todayStart, time.Now())
	if err == nil {
		var todayRevenue float64
		for _, order := range todayOrders {
			todayRevenue += float64(order.TotalAmount) / 100.0 // 转换为元
		}

		metrics["daily_orders"] = len(todayOrders)
		metrics["daily_revenue"] = todayRevenue
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

// 注意：这个接口已经在 order_repository.go 中定义，这里只添加分析专用方法
type AnalyticsOrderRepository interface {
	// 分析专用方法
	GetOrdersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Order, error)
	GetOrderCountByCustomer(ctx context.Context, customerID string, startDate, endDate time.Time) (int, error)
}

// Repository interfaces

type AnalyticsRepository interface {
	SaveReport(ctx context.Context, report *AnalyticsReport) error
	GetReport(ctx context.Context, reportID string) (*AnalyticsReport, error)
	GetReportsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*AnalyticsReport, error)
	UpdateReport(ctx context.Context, report *AnalyticsReport) error
	DeleteReport(ctx context.Context, reportID string) error
}

// Customer 客户实体（简化版本）
type Customer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CustomerRepository interface {
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
	GetCustomersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*Customer, error)
	SaveCustomer(ctx context.Context, customer *Customer) error
	UpdateCustomer(ctx context.Context, customer *Customer) error
}

// Product 产品实体（简化版本）
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

type ProductRepository interface {
	GetProduct(ctx context.Context, productID string) (*Product, error)
	GetProductsByCategory(ctx context.Context, category string) ([]*Product, error)
	SaveProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, product *Product) error
}
