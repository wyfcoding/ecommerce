package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/report/model"
	"ecommerce/internal/report/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrReportNotFound = errors.New("报表不存在")
	ErrInvalidPeriod  = errors.New("无效的报表周期")
	ErrInvalidDateRange = errors.New("无效的日期范围")
)

// ReportService 报表服务接口
type ReportService interface {
	// 报表管理
	CreateReport(ctx context.Context, report *model.Report) (*model.Report, error)
	GetReport(ctx context.Context, id uint64) (*model.Report, error)
	ListReports(ctx context.Context, reportType string, period string, startDate, endDate time.Time, pageSize, pageNum int32) ([]*model.Report, int64, error)
	DeleteReport(ctx context.Context, id uint64) error
	
	// 报表生成
	GenerateSalesReport(ctx context.Context, startDate, endDate time.Time) (*model.SalesReport, error)
	GenerateOrderReport(ctx context.Context, startDate, endDate time.Time) (*model.OrderReport, error)
	GenerateUserReport(ctx context.Context, startDate, endDate time.Time) (*model.UserReport, error)
	GenerateProductReport(ctx context.Context, startDate, endDate time.Time) (*model.ProductReport, error)
	GenerateInventoryReport(ctx context.Context, date time.Time) (*model.InventoryReport, error)
	GenerateFinanceReport(ctx context.Context, startDate, endDate time.Time) (*model.FinanceReport, error)
	GenerateMarketingReport(ctx context.Context, startDate, endDate time.Time) (*model.MarketingReport, error)
	
	// 定时报表
	GenerateDailyReport(ctx context.Context, date time.Time) error
	GenerateWeeklyReport(ctx context.Context, startDate time.Time) error
	GenerateMonthlyReport(ctx context.Context, year int, month int) error
	
	// 报表导出
	ExportReport(ctx context.Context, reportID uint64, format string) ([]byte, error)
}

type reportService struct {
	repo   repository.ReportRepo
	logger *zap.Logger
}

// NewReportService 创建报表服务实例
func NewReportService(
	repo repository.ReportRepo,
	logger *zap.Logger,
) ReportService {
	return &reportService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReport 创建报表
func (s *reportService) CreateReport(ctx context.Context, report *model.Report) (*model.Report, error) {
	report.ReportNo = fmt.Sprintf("RPT%d", idgen.GenID())
	report.Status = model.ReportStatusPending
	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.Error("创建报表失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建报表成功", zap.Uint64("reportID", report.ID))
	return report, nil
}

// GetReport 获取报表详情
func (s *reportService) GetReport(ctx context.Context, id uint64) (*model.Report, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, ErrReportNotFound
	}
	return report, nil
}

// ListReports 获取报表列表
func (s *reportService) ListReports(ctx context.Context, reportType string, period string, startDate, endDate time.Time, pageSize, pageNum int32) ([]*model.Report, int64, error) {
	reports, total, err := s.repo.ListReports(ctx, reportType, period, startDate, endDate, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取报表列表失败", zap.Error(err))
		return nil, 0, err
	}
	return reports, total, nil
}

// DeleteReport 删除报表
func (s *reportService) DeleteReport(ctx context.Context, id uint64) error {
	if err := s.repo.DeleteReport(ctx, id); err != nil {
		s.logger.Error("删除报表失败", zap.Error(err))
		return err
	}
	return nil
}

// GenerateSalesReport 生成销售报表
func (s *reportService) GenerateSalesReport(ctx context.Context, startDate, endDate time.Time) (*model.SalesReport, error) {
	s.logger.Info("开始生成销售报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取基础数据
	salesData, err := s.repo.GetSalesData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 获取用户数据
	userData, err := s.repo.GetUserData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 获取商品数据
	productData, err := s.repo.GetProductData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. 获取每日数据
	dailyData, err := s.repo.GetDailySalesData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. 获取分类数据
	categoryData, err := s.repo.GetCategorySalesData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 6. 获取商品排行
	productRanking, err := s.repo.GetProductRanking(ctx, startDate, endDate, 10)
	if err != nil {
		return nil, err
	}

	// 7. 获取地区数据
	regionData, err := s.repo.GetRegionSalesData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 8. 计算指标
	report := &model.SalesReport{
		TotalOrders:    salesData.TotalOrders,
		TotalAmount:    salesData.TotalAmount,
		TotalRefund:    salesData.TotalRefund,
		NetAmount:      salesData.TotalAmount - salesData.TotalRefund,
		TotalUsers:     userData.TotalUsers,
		NewUsers:       userData.NewUsers,
		ActiveUsers:    userData.ActiveUsers,
		PayingUsers:    userData.PayingUsers,
		TotalProducts:  productData.TotalProducts,
		SoldProducts:   productData.SoldProducts,
		TotalQuantity:  productData.TotalQuantity,
		DailyData:      dailyData,
		CategoryData:   categoryData,
		ProductRanking: productRanking,
		RegionData:     regionData,
	}

	// 计算平均订单金额
	if report.TotalOrders > 0 {
		report.AvgOrderAmount = report.TotalAmount / report.TotalOrders
	}

	// 计算复购率
	if report.PayingUsers > 0 {
		repeatUsers, _ := s.repo.GetRepeatPurchaseUsers(ctx, startDate, endDate)
		report.RepurchaseRate = float64(repeatUsers) / float64(report.PayingUsers)
	}

	// 计算转化率
	if report.ActiveUsers > 0 {
		report.ConversionRate = float64(report.PayingUsers) / float64(report.ActiveUsers)
	}

	// 计算支付成功率
	if report.TotalOrders > 0 {
		paidOrders, _ := s.repo.GetPaidOrderCount(ctx, startDate, endDate)
		report.PaymentRate = float64(paidOrders) / float64(report.TotalOrders)
	}

	// 计算退款率
	if report.TotalOrders > 0 {
		refundOrders, _ := s.repo.GetRefundOrderCount(ctx, startDate, endDate)
		report.RefundRate = float64(refundOrders) / float64(report.TotalOrders)
	}

	s.logger.Info("销售报表生成成功")
	return report, nil
}

// GenerateOrderReport 生成订单报表
func (s *reportService) GenerateOrderReport(ctx context.Context, startDate, endDate time.Time) (*model.OrderReport, error) {
	s.logger.Info("开始生成订单报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取订单统计数据
	orderStats, err := s.repo.GetOrderStatistics(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 获取每小时数据
	hourlyData, err := s.repo.GetHourlyOrderData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 获取支付方式数据
	paymentMethodData, err := s.repo.GetPaymentMethodData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. 计算时效指标
	avgPaymentTime, _ := s.repo.GetAvgPaymentTime(ctx, startDate, endDate)
	avgShippingTime, _ := s.repo.GetAvgShippingTime(ctx, startDate, endDate)
	avgDeliveryTime, _ := s.repo.GetAvgDeliveryTime(ctx, startDate, endDate)

	report := &model.OrderReport{
		TotalOrders:       orderStats.TotalOrders,
		PendingOrders:     orderStats.PendingOrders,
		PaidOrders:        orderStats.PaidOrders,
		ShippedOrders:     orderStats.ShippedOrders,
		CompletedOrders:   orderStats.CompletedOrders,
		CancelledOrders:   orderStats.CancelledOrders,
		TotalAmount:       orderStats.TotalAmount,
		PaidAmount:        orderStats.PaidAmount,
		RefundAmount:      orderStats.RefundAmount,
		AvgPaymentTime:    avgPaymentTime,
		AvgShippingTime:   avgShippingTime,
		AvgDeliveryTime:   avgDeliveryTime,
		HourlyData:        hourlyData,
		PaymentMethodData: paymentMethodData,
	}

	s.logger.Info("订单报表生成成功")
	return report, nil
}

// GenerateUserReport 生成用户报表
func (s *reportService) GenerateUserReport(ctx context.Context, startDate, endDate time.Time) (*model.UserReport, error) {
	s.logger.Info("开始生成用户报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取用户统计数据
	userStats, err := s.repo.GetUserStatistics(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 获取每日用户数据
	dailyUserData, err := s.repo.GetDailyUserData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 获取用户等级数据
	userLevelData, err := s.repo.GetUserLevelData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. 获取用户地区数据
	userRegionData, err := s.repo.GetUserRegionData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 5. 计算用户行为指标
	avgLoginFreq, _ := s.repo.GetAvgLoginFrequency(ctx, startDate, endDate)
	avgBrowseTime, _ := s.repo.GetAvgBrowseTime(ctx, startDate, endDate)
	avgOrderCount, _ := s.repo.GetAvgOrderCount(ctx, startDate, endDate)
	avgOrderAmount, _ := s.repo.GetAvgOrderAmount(ctx, startDate, endDate)

	// 6. 计算用户价值
	totalLTV, _ := s.repo.GetTotalLTV(ctx, startDate, endDate)
	avgLTV := int64(0)
	if userStats.TotalUsers > 0 {
		avgLTV = totalLTV / userStats.TotalUsers
	}

	// 7. 计算留存率
	retentionDay1, _ := s.repo.GetRetentionRate(ctx, startDate, 1)
	retentionDay7, _ := s.repo.GetRetentionRate(ctx, startDate, 7)
	retentionDay30, _ := s.repo.GetRetentionRate(ctx, startDate, 30)

	report := &model.UserReport{
		TotalUsers:      userStats.TotalUsers,
		NewUsers:        userStats.NewUsers,
		ActiveUsers:     userStats.ActiveUsers,
		PayingUsers:     userStats.PayingUsers,
		AvgLoginFreq:    avgLoginFreq,
		AvgBrowseTime:   avgBrowseTime,
		AvgOrderCount:   avgOrderCount,
		AvgOrderAmount:  avgOrderAmount,
		TotalLTV:        totalLTV,
		AvgLTV:          avgLTV,
		RetentionDay1:   retentionDay1,
		RetentionDay7:   retentionDay7,
		RetentionDay30:  retentionDay30,
		DailyUserData:   dailyUserData,
		UserLevelData:   userLevelData,
		UserRegionData:  userRegionData,
	}

	s.logger.Info("用户报表生成成功")
	return report, nil
}

// GenerateProductReport 生成商品报表
func (s *reportService) GenerateProductReport(ctx context.Context, startDate, endDate time.Time) (*model.ProductReport, error) {
	s.logger.Info("开始生成商品报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取商品统计数据
	productStats, err := s.repo.GetProductStatistics(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 获取分类数据
	categoryData, err := s.repo.GetCategoryProductData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 获取品牌数据
	brandData, err := s.repo.GetBrandProductData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 4. 获取热销商品
	topProducts, err := s.repo.GetTopProducts(ctx, startDate, endDate, 20)
	if err != nil {
		return nil, err
	}

	// 5. 获取新品数据
	newProducts, err := s.repo.GetNewProducts(ctx, startDate, endDate, 10)
	if err != nil {
		return nil, err
	}

	report := &model.ProductReport{
		TotalProducts:   productStats.TotalProducts,
		OnSaleProducts:  productStats.OnSaleProducts,
		SoldOutProducts: productStats.SoldOutProducts,
		TotalSales:      productStats.TotalSales,
		TotalAmount:     productStats.TotalAmount,
		AvgPrice:        productStats.AvgPrice,
		TotalStock:      productStats.TotalStock,
		LowStockCount:   productStats.LowStockCount,
		CategoryData:    categoryData,
		BrandData:       brandData,
		TopProducts:     topProducts,
		NewProducts:     newProducts,
	}

	s.logger.Info("商品报表生成成功")
	return report, nil
}

// GenerateInventoryReport 生成库存报表
func (s *reportService) GenerateInventoryReport(ctx context.Context, date time.Time) (*model.InventoryReport, error) {
	s.logger.Info("开始生成库存报表", zap.Time("date", date))

	// 1. 获取库存统计数据
	inventoryStats, err := s.repo.GetInventoryStatistics(ctx, date)
	if err != nil {
		return nil, err
	}

	// 2. 获取每日库存数据
	startDate := date.AddDate(0, 0, -30)
	dailyStockData, err := s.repo.GetDailyStockData(ctx, startDate, date)
	if err != nil {
		return nil, err
	}

	// 3. 获取分类库存数据
	categoryStockData, err := s.repo.GetCategoryStockData(ctx, date)
	if err != nil {
		return nil, err
	}

	// 4. 获取预警商品
	warningProducts, err := s.repo.GetWarningProducts(ctx, date)
	if err != nil {
		return nil, err
	}

	// 5. 计算库存周转率
	turnoverRate, _ := s.repo.GetInventoryTurnoverRate(ctx, startDate, date)
	avgTurnoverDays := float64(0)
	if turnoverRate > 0 {
		avgTurnoverDays = 365.0 / turnoverRate
	}

	report := &model.InventoryReport{
		TotalStock:        inventoryStats.TotalStock,
		TotalValue:        inventoryStats.TotalValue,
		LowStockCount:     inventoryStats.LowStockCount,
		OutOfStockCount:   inventoryStats.OutOfStockCount,
		TurnoverRate:      turnoverRate,
		AvgTurnoverDays:   avgTurnoverDays,
		DailyStockData:    dailyStockData,
		CategoryStockData: categoryStockData,
		WarningProducts:   warningProducts,
	}

	s.logger.Info("库存报表生成成功")
	return report, nil
}

// GenerateFinanceReport 生成财务报表
func (s *reportService) GenerateFinanceReport(ctx context.Context, startDate, endDate time.Time) (*model.FinanceReport, error) {
	s.logger.Info("开始生成财务报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取财务数据
	financeData, err := s.repo.GetFinanceData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 获取每日财务数据
	dailyFinanceData, err := s.repo.GetDailyFinanceData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 3. 计算利润
	grossProfit := financeData.TotalRevenue - financeData.TotalExpense
	netProfit := grossProfit // 简化处理，实际应扣除税费等
	profitMargin := float64(0)
	if financeData.TotalRevenue > 0 {
		profitMargin = float64(netProfit) / float64(financeData.TotalRevenue)
	}

	report := &model.FinanceReport{
		TotalRevenue:      financeData.TotalRevenue,
		ProductRevenue:    financeData.ProductRevenue,
		ServiceRevenue:    financeData.ServiceRevenue,
		OtherRevenue:      financeData.OtherRevenue,
		TotalExpense:      financeData.TotalExpense,
		RefundExpense:     financeData.RefundExpense,
		CommissionExpense: financeData.CommissionExpense,
		OtherExpense:      financeData.OtherExpense,
		GrossProfit:       grossProfit,
		NetProfit:         netProfit,
		ProfitMargin:      profitMargin,
		DailyFinanceData:  dailyFinanceData,
	}

	s.logger.Info("财务报表生成成功")
	return report, nil
}

// GenerateMarketingReport 生成营销报表
func (s *reportService) GenerateMarketingReport(ctx context.Context, startDate, endDate time.Time) (*model.MarketingReport, error) {
	s.logger.Info("开始生成营销报表",
		zap.Time("startDate", startDate),
		zap.Time("endDate", endDate))

	// 1. 获取营销数据
	marketingData, err := s.repo.GetMarketingData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// 2. 计算优惠券使用率
	couponUsageRate := float64(0)
	if marketingData.IssuedCoupons > 0 {
		couponUsageRate = float64(marketingData.UsedCoupons) / float64(marketingData.IssuedCoupons)
	}

	// 3. 计算ROI
	roi := float64(0)
	if marketingData.MarketingCost > 0 {
		roi = float64(marketingData.MarketingRevenue) / float64(marketingData.MarketingCost)
	}

	report := &model.MarketingReport{
		TotalCoupons:     marketingData.TotalCoupons,
		IssuedCoupons:    marketingData.IssuedCoupons,
		UsedCoupons:      marketingData.UsedCoupons,
		CouponUsageRate:  couponUsageRate,
		CouponDiscount:   marketingData.CouponDiscount,
		TotalPromotions:  marketingData.TotalPromotions,
		ActivePromotions: marketingData.ActivePromotions,
		PromotionOrders:  marketingData.PromotionOrders,
		PromotionAmount:  marketingData.PromotionAmount,
		TotalPoints:      marketingData.TotalPoints,
		UsedPoints:       marketingData.UsedPoints,
		PointsValue:      marketingData.PointsValue,
		MarketingCost:    marketingData.MarketingCost,
		MarketingRevenue: marketingData.MarketingRevenue,
		ROI:              roi,
	}

	s.logger.Info("营销报表生成成功")
	return report, nil
}

// GenerateDailyReport 生成日报
func (s *reportService) GenerateDailyReport(ctx context.Context, date time.Time) error {
	s.logger.Info("开始生成日报", zap.Time("date", date))

	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.AddDate(0, 0, 1)

	// 生成销售报表
	salesReport, err := s.GenerateSalesReport(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	// 保存报表
	data, _ := json.Marshal(salesReport)
	report := &model.Report{
		Type:        model.ReportTypeSales,
		Period:      model.ReportPeriodDaily,
		Title:       fmt.Sprintf("%s 销售日报", date.Format("2006-01-02")),
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.ReportStatusCompleted,
		Data:        string(data),
		GeneratedAt: &date,
	}

	_, err = s.CreateReport(ctx, report)
	if err != nil {
		return err
	}

	s.logger.Info("日报生成成功")
	return nil
}

// GenerateWeeklyReport 生成周报
func (s *reportService) GenerateWeeklyReport(ctx context.Context, startDate time.Time) error {
	s.logger.Info("开始生成周报", zap.Time("startDate", startDate))

	endDate := startDate.AddDate(0, 0, 7)

	// 生成销售报表
	salesReport, err := s.GenerateSalesReport(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	// 保存报表
	data, _ := json.Marshal(salesReport)
	now := time.Now()
	report := &model.Report{
		Type:        model.ReportTypeSales,
		Period:      model.ReportPeriodWeekly,
		Title:       fmt.Sprintf("%s - %s 销售周报", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.ReportStatusCompleted,
		Data:        string(data),
		GeneratedAt: &now,
	}

	_, err = s.CreateReport(ctx, report)
	if err != nil {
		return err
	}

	s.logger.Info("周报生成成功")
	return nil
}

// GenerateMonthlyReport 生成月报
func (s *reportService) GenerateMonthlyReport(ctx context.Context, year int, month int) error {
	s.logger.Info("开始生成月报", zap.Int("year", year), zap.Int("month", month))

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	// 生成销售报表
	salesReport, err := s.GenerateSalesReport(ctx, startDate, endDate)
	if err != nil {
		return err
	}

	// 保存报表
	data, _ := json.Marshal(salesReport)
	now := time.Now()
	report := &model.Report{
		Type:        model.ReportTypeSales,
		Period:      model.ReportPeriodMonthly,
		Title:       fmt.Sprintf("%d年%d月 销售月报", year, month),
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      model.ReportStatusCompleted,
		Data:        string(data),
		GeneratedAt: &now,
	}

	_, err = s.CreateReport(ctx, report)
	if err != nil {
		return err
	}

	s.logger.Info("月报生成成功")
	return nil
}

// ExportReport 导出报表
func (s *reportService) ExportReport(ctx context.Context, reportID uint64, format string) ([]byte, error) {
	report, err := s.GetReport(ctx, reportID)
	if err != nil {
		return nil, err
	}

	switch format {
	case "JSON":
		return []byte(report.Data), nil
	case "CSV":
		// TODO: 实现CSV导出
		return nil, fmt.Errorf("CSV导出功能待实现")
	case "EXCEL":
		// TODO: 实现Excel导出
		return nil, fmt.Errorf("Excel导出功能待实现")
	case "PDF":
		// TODO: 实现PDF导出
		return nil, fmt.Errorf("PDF导出功能待实现")
	default:
		return nil, fmt.Errorf("不支持的导出格式: %s", format)
	}
}
