package domain

import (
	"errors"
	"time"
)

var (
	ErrAnalyticsNotFound    = errors.New("analytics not found")
	ErrInvalidDateRange     = errors.New("invalid date range")
	ErrInsufficientData     = errors.New("insufficient data")
)

type MetricType string

const (
	MetricTypeImpression   MetricType = "IMPRESSION"
	MetricTypeClick        MetricType = "CLICK"
	MetricTypeConversion   MetricType = "CONVERSION"
	MetricTypeGMV          MetricType = "GMV"
	MetricTypeRevenue      MetricType = "REVENUE"
	MetricTypeCost         MetricType = "COST"
	MetricTypeROI          MetricType = "ROI"
	MetricTypeCTR          MetricType = "CTR"
	MetricTypeCVR          MetricType = "CVR"
	MetricTypeAOV          MetricType = "AOV"
	MetricTypeCAC          MetricType = "CAC"
	MetricTypeLTV          MetricType = "LTV"
)

type CampaignAnalytics struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	CampaignID       uint64    `json:"campaign_id"`
	CampaignName     string    `json:"campaign_name"`
	CampaignType     CampaignType `json:"campaign_type"`
	Date             time.Time `json:"date"`
	
	Impressions      int64     `json:"impressions"`
	Clicks           int64     `json:"clicks"`
	CTR              float64   `json:"ctr"`
	
	Visitors         int64     `json:"visitors"`
	UniqueVisitors   int64     `json:"unique_visitors"`
	
	Orders           int64     `json:"orders"`
	OrderAmount      int64     `json:"order_amount"`
	CVR              float64   `json:"cvr"`
	AOV              float64   `json:"aov"`
	
	GMV              int64     `json:"gmv"`
	Revenue          int64     `json:"revenue"`
	Discount         int64     `json:"discount"`
	Refund           int64     `json:"refund"`
	NetRevenue       int64     `json:"net_revenue"`
	
	Cost             int64     `json:"cost"`
	CPC              float64   `json:"cpc"`
	CPM              float64   `json:"cpm"`
	CPA              float64   `json:"cpa"`
	ROI              float64   `json:"roi"`
	ROAS             float64   `json:"roas"`
	
	NewUsers         int64     `json:"new_users"`
	NewUserCost      float64   `json:"new_user_cost"`
	RepeatUsers      int64     `json:"repeat_users"`
	RepeatRate       float64   `json:"repeat_rate"`
	
	Participations   int64     `json:"participations"`
	SuccessCount     int64     `json:"success_count"`
	SuccessRate      float64   `json:"success_rate"`
	
	ShareCount       int64     `json:"share_count"`
	LikeCount        int64     `json:"like_count"`
	CommentCount     int64     `json:"comment_count"`
	
	DeviceBreakdown  *DeviceBreakdown `json:"device_breakdown"`
	RegionBreakdown  *RegionBreakdown `json:"region_breakdown"`
	AgeBreakdown     *AgeBreakdown    `json:"age_breakdown"`
	GenderBreakdown  *GenderBreakdown `json:"gender_breakdown"`
	
	TopProducts      []*TopProduct    `json:"top_products"`
	TopChannels      []*TopChannel    `json:"top_channels"`
}

type DeviceBreakdown struct {
	Mobile  int64 `json:"mobile"`
	Desktop int64 `json:"desktop"`
	Tablet  int64 `json:"tablet"`
	Other   int64 `json:"other"`
}

type RegionBreakdown struct {
	Regions []*RegionMetric `json:"regions"`
}

type RegionMetric struct {
	Region string `json:"region"`
	Count  int64  `json:"count"`
	Amount int64  `json:"amount"`
}

type AgeBreakdown struct {
	Age18_24 int64 `json:"age_18_24"`
	Age25_34 int64 `json:"age_25_34"`
	Age35_44 int64 `json:"age_35_44"`
	Age45_54 int64 `json:"age_45_54"`
	Age55Plus int64 `json:"age_55_plus"`
}

type GenderBreakdown struct {
	Male   int64 `json:"male"`
	Female int64 `json:"female"`
	Other  int64 `json:"other"`
}

type TopProduct struct {
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Sales       int64  `json:"sales"`
	Amount      int64  `json:"amount"`
}

type TopChannel struct {
	Channel string `json:"channel"`
	Visits  int64  `json:"visits"`
	Orders  int64  `json:"orders"`
	Amount  int64  `json:"amount"`
}

type AnalyticsReport struct {
	ID           uint                   `json:"id"`
	CreatedAt    time.Time              `json:"created_at"`
	ReportNo     string                 `json:"report_no"`
	ReportType   string                 `json:"report_type"`
	StartDate    time.Time              `json:"start_date"`
	EndDate      time.Time              `json:"end_date"`
	
	Summary      *AnalyticsSummary      `json:"summary"`
	DailyData    []*CampaignAnalytics   `json:"daily_data"`
	Comparisons  []*MetricComparison    `json:"comparisons"`
	Trends       []*MetricTrend         `json:"trends"`
	Insights     []*AnalyticsInsight    `json:"insights"`
	
	Status       string                 `json:"status"`
	GeneratedAt  *time.Time             `json:"generated_at"`
}

type AnalyticsSummary struct {
	TotalImpressions   int64   `json:"total_impressions"`
	TotalClicks        int64   `json:"total_clicks"`
	AvgCTR             float64 `json:"avg_ctr"`
	
	TotalOrders        int64   `json:"total_orders"`
	TotalGMV           int64   `json:"total_gmv"`
	TotalRevenue       int64   `json:"total_revenue"`
	AvgCVR             float64 `json:"avg_cvr"`
	AvgAOV             float64 `json:"avg_aov"`
	
	TotalCost          int64   `json:"total_cost"`
	TotalROI           float64 `json:"total_roi"`
	TotalROAS          float64 `json:"total_roas"`
	
	TotalNewUsers      int64   `json:"total_new_users"`
	AvgCAC             float64 `json:"avg_cac"`
	
	TotalParticipations int64  `json:"total_participations"`
	TotalSuccess       int64   `json:"total_success"`
	AvgSuccessRate     float64 `json:"avg_success_rate"`
	
	PeriodOverPeriod   *PeriodComparison `json:"period_over_period"`
}

type PeriodComparison struct {
	GMVChange         float64 `json:"gmv_change"`
	RevenueChange     float64 `json:"revenue_change"`
	OrdersChange      float64 `json:"orders_change"`
	ROIChange         float64 `json:"roi_change"`
	CVRChange         float64 `json:"cvr_change"`
	NewUsersChange    float64 `json:"new_users_change"`
}

type MetricComparison struct {
	MetricName    string  `json:"metric_name"`
	CurrentValue  float64 `json:"current_value"`
	PreviousValue float64 `json:"previous_value"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

type MetricTrend struct {
	MetricName string    `json:"metric_name"`
	Values     []*TrendValue `json:"values"`
	Trend      string    `json:"trend"`
}

type TrendValue struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

type AnalyticsInsight struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Change      float64   `json:"change"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

type AnalyticsAlert struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	CampaignID  uint64    `json:"campaign_id"`
	MetricType  MetricType `json:"metric_type"`
	AlertType   string    `json:"alert_type"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	AcknowledgedBy uint64 `json:"acknowledged_by"`
}

func NewCampaignAnalytics(campaignID uint64, campaignName string, campaignType CampaignType, date time.Time) *CampaignAnalytics {
	return &CampaignAnalytics{
		CampaignID:   campaignID,
		CampaignName: campaignName,
		CampaignType: campaignType,
		Date:         date,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (a *CampaignAnalytics) RecordImpression(count int64) {
	a.Impressions += count
	a.calculateCTR()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordClick(count int64) {
	a.Clicks += count
	a.calculateCTR()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordOrder(orderAmount, discount int64) {
	a.Orders++
	a.OrderAmount += orderAmount
	a.Discount += discount
	a.calculateCVR()
	a.calculateAOV()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordGMV(gmv int64) {
	a.GMV += gmv
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordRevenue(revenue int64) {
	a.Revenue += revenue
	a.NetRevenue = a.Revenue - a.Refund
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordCost(cost int64) {
	a.Cost += cost
	a.calculateROI()
	a.calculateCPM()
	a.calculateCPC()
	a.calculateCPA()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordRefund(refund int64) {
	a.Refund += refund
	a.NetRevenue = a.Revenue - a.Refund
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordNewUser(count int64) {
	a.NewUsers += count
	if a.NewUsers > 0 && a.Cost > 0 {
		a.NewUserCost = float64(a.Cost) / float64(a.NewUsers)
	}
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordRepeatUser(count int64) {
	a.RepeatUsers += count
	a.calculateRepeatRate()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordParticipation(count int64) {
	a.Participations += count
	a.calculateSuccessRate()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) RecordSuccess(count int64) {
	a.SuccessCount += count
	a.calculateSuccessRate()
	a.UpdatedAt = time.Now()
}

func (a *CampaignAnalytics) calculateCTR() {
	if a.Impressions > 0 {
		a.CTR = float64(a.Clicks) / float64(a.Impressions) * 100
	}
}

func (a *CampaignAnalytics) calculateCVR() {
	if a.Visitors > 0 {
		a.CVR = float64(a.Orders) / float64(a.Visitors) * 100
	}
}

func (a *CampaignAnalytics) calculateAOV() {
	if a.Orders > 0 {
		a.AOV = float64(a.OrderAmount) / float64(a.Orders)
	}
}

func (a *CampaignAnalytics) calculateROI() {
	if a.Cost > 0 {
		a.ROI = float64(a.Revenue-a.Cost) / float64(a.Cost) * 100
		a.ROAS = float64(a.Revenue) / float64(a.Cost)
	}
}

func (a *CampaignAnalytics) calculateCPM() {
	if a.Impressions > 0 {
		a.CPM = float64(a.Cost) / float64(a.Impressions) * 1000
	}
}

func (a *CampaignAnalytics) calculateCPC() {
	if a.Clicks > 0 {
		a.CPC = float64(a.Cost) / float64(a.Clicks)
	}
}

func (a *CampaignAnalytics) calculateCPA() {
	if a.Orders > 0 {
		a.CPA = float64(a.Cost) / float64(a.Orders)
	}
}

func (a *CampaignAnalytics) calculateRepeatRate() {
	total := a.NewUsers + a.RepeatUsers
	if total > 0 {
		a.RepeatRate = float64(a.RepeatUsers) / float64(total) * 100
	}
}

func (a *CampaignAnalytics) calculateSuccessRate() {
	if a.Participations > 0 {
		a.SuccessRate = float64(a.SuccessCount) / float64(a.Participations) * 100
	}
}

func NewAnalyticsReport(reportType string, startDate, endDate time.Time) *AnalyticsReport {
	return &AnalyticsReport{
		ReportNo:    generateReportNo(),
		ReportType:  reportType,
		StartDate:   startDate,
		EndDate:     endDate,
		DailyData:   make([]*CampaignAnalytics, 0),
		Comparisons: make([]*MetricComparison, 0),
		Trends:      make([]*MetricTrend, 0),
		Insights:    make([]*AnalyticsInsight, 0),
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}
}

func generateReportNo() string {
	return "RPT" + time.Now().Format("20060102150405")
}

func (r *AnalyticsReport) SetSummary(summary *AnalyticsSummary) {
	r.Summary = summary
}

func (r *AnalyticsReport) AddDailyData(data *CampaignAnalytics) {
	r.DailyData = append(r.DailyData, data)
}

func (r *AnalyticsReport) AddComparison(comparison *MetricComparison) {
	r.Comparisons = append(r.Comparisons, comparison)
}

func (r *AnalyticsReport) AddTrend(trend *MetricTrend) {
	r.Trends = append(r.Trends, trend)
}

func (r *AnalyticsReport) AddInsight(insight *AnalyticsInsight) {
	r.Insights = append(r.Insights, insight)
}

func (r *AnalyticsReport) Complete() {
	r.Status = "COMPLETED"
	now := time.Now()
	r.GeneratedAt = &now
}

func NewAnalyticsAlert(campaignID uint64, metricType MetricType, alertType string, threshold, currentValue float64, message string) *AnalyticsAlert {
	return &AnalyticsAlert{
		CampaignID:   campaignID,
		MetricType:   metricType,
		AlertType:    alertType,
		Threshold:    threshold,
		CurrentValue: currentValue,
		Message:      message,
		Status:       "ACTIVE",
		CreatedAt:    time.Now(),
	}
}

func (a *AnalyticsAlert) Acknowledge(by uint64) {
	a.Status = "ACKNOWLEDGED"
	now := time.Now()
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = by
}

type AnalyticsRepository interface {
	Save(ctx interface{}, analytics *CampaignAnalytics) error
	FindByID(ctx interface{}, id uint) (*CampaignAnalytics, error)
	FindByCampaignID(ctx interface{}, campaignID uint64, startDate, endDate time.Time) ([]*CampaignAnalytics, error)
	FindByDate(ctx interface{}, date time.Time) ([]*CampaignAnalytics, error)
	FindByType(ctx interface{}, campaignType CampaignType, startDate, endDate time.Time) ([]*CampaignAnalytics, error)
	AggregateByCampaign(ctx interface{}, campaignID uint64, startDate, endDate time.Time) (*AnalyticsSummary, error)
	AggregateByType(ctx interface{}, campaignType CampaignType, startDate, endDate time.Time) (*AnalyticsSummary, error)
	
	SaveReport(ctx interface{}, report *AnalyticsReport) error
	FindReportByID(ctx interface{}, id uint) (*AnalyticsReport, error)
	FindReportsByDateRange(ctx interface{}, startDate, endDate time.Time) ([]*AnalyticsReport, error)
	
	SaveAlert(ctx interface{}, alert *AnalyticsAlert) error
	FindAlertByID(ctx interface{}, id uint) (*AnalyticsAlert, error)
	FindActiveAlerts(ctx interface{}) ([]*AnalyticsAlert, error)
	FindAlertsByCampaign(ctx interface{}, campaignID uint64) ([]*AnalyticsAlert, error)
}

type AnalyticsService interface {
	RecordMetric(ctx interface{}, campaignID uint64, metricType MetricType, value float64) error
	GetCampaignAnalytics(ctx interface{}, campaignID uint64, startDate, endDate time.Time) (*AnalyticsSummary, error)
	GenerateReport(ctx interface{}, reportType string, startDate, endDate time.Time) (*AnalyticsReport, error)
	GetInsights(ctx interface{}, campaignID uint64) ([]*AnalyticsInsight, error)
	CheckAlerts(ctx interface{}) error
}
