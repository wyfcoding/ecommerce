package repository

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ecommerce/internal/analytics/model"
)

// --- 接口定义 ---

// AnalyticsRepository 定义了分析数据仓库的接口。
// 提供了对页面浏览事件和销售事实数据的批量写入，以及各种分析查询功能。
type AnalyticsRepository interface {
	// 写入操作
	// BatchInsertPageViewEvents 批量插入页面浏览事件。
	BatchInsertPageViewEvents(ctx context.Context, events []*model.PageViewEvent) error
	// BatchInsertSalesFacts 批量插入销售事实数据。
	BatchInsertSalesFacts(ctx context.Context, facts []*model.SalesFact) error

	// 读取/查询操作
	// QueryTotalRevenue 查询指定时间范围内的总销售额。
	QueryTotalRevenue(ctx context.Context, startTime, endTime *time.Time) (float64, error)
	// QuerySalesByProductCategory 查询按商品分类划分的销售额。
	QuerySalesByProductCategory(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error)
	// QuerySalesByProductBrand 查询按商品品牌划分的销售额。
	QuerySalesByProductBrand(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error)
	// QueryUserActivityCount 查询指定时间范围内的活跃用户数。
	QueryUserActivityCount(ctx context.Context, startTime, endTime *time.Time) (int64, error)
	// QueryTopNProductsByRevenue 查询销售额最高的N个商品。
	QueryTopNProductsByRevenue(ctx context.Context, n int, startTime, endTime *time.Time) ([]*model.ProductSales, error)
	// QueryConversionRate 查询从页面浏览到下单的转化率。
	QueryConversionRate(ctx context.Context, startTime, endTime *time.Time) (float64, error)
}

// --- 数据库模型 ---

// DBPageViewEvent 对应数据库中的页面浏览事件表。
type DBPageViewEvent struct {
	EventTime time.Time `gorm:"index;comment:事件发生时间"`
	UserID    uint      `gorm:"index;comment:用户ID"`
	SessionID string    `gorm:"index;type:varchar(100);comment:会话ID"`
	URL       string    `gorm:"type:text;comment:浏览的页面URL"`
	Referer   string    `gorm:"type:text;comment:来源页面"`
	UserAgent string    `gorm:"type:text;comment:浏览器User-Agent"`
	ClientIP  string    `gorm:"type:varchar(45);comment:客户端IP地址"`
}

// TableName 自定义 DBPageViewEvent 对应的表名。
func (DBPageViewEvent) TableName() string {
	return "page_view_events"
}

// DBSalesFact 对应数据库中的销售事实表。
type DBSalesFact struct {
	EventTime       time.Time `gorm:"index;comment:销售事件发生时间"`
	OrderID         uint      `gorm:"index;comment:订单ID"`
	OrderItemID     uint      `gorm:"primarykey;comment:订单项ID"`
	OrderSN         string    `gorm:"index;type:varchar(100);comment:订单号"`
	OrderTotal      float64   `gorm:"type:decimal(10,2);comment:订单总金额"`
	DiscountAmount  float64   `gorm:"type:decimal(10,2);comment:订单优惠金额"`
	ProductID       uint      `gorm:"index;comment:商品ID"`
	ProductSKU      string    `gorm:"type:varchar(100);comment:商品SKU"`
	ProductName     string    `gorm:"type:varchar(255);comment:商品名称"`
	ProductCategory string    `gorm:"index;type:varchar(100);comment:商品分类"`
	ProductBrand    string    `gorm:"index;type:varchar(100);comment:商品品牌"`
	ItemPrice       float64   `gorm:"type:decimal(10,2);comment:单个商品项价格"`
	ItemQuantity    int       `gorm:"comment:单个商品项数量"`
	UserID          uint      `gorm:"index;comment:用户ID"`
	UserCity        string    `gorm:"type:varchar(100);comment:用户所在城市"`
	UserCountry     string    `gorm:"index;type:varchar(100);comment:用户所在国家"`
}

// TableName 自定义 DBSalesFact 对应的表名。
func (DBSalesFact) TableName() string {
	return "sales_facts"
}

// --- 数据层核心 ---

// Data 封装了所有数据库操作的 GORM 客户端。
type Data struct {
	db *gorm.DB
}

// NewData 创建一个新的 Data 实例，并执行数据库迁移。
func NewData(db *gorm.DB) (*Data, func(), error) {
	d := &Data{
		db: db,
	}
	zap.S().Info("Running database migrations for analytics service...")
	// 自动迁移所有相关的数据库表
	if err := db.AutoMigrate(
		&DBPageViewEvent{},
		&DBSalesFact{},
	); err != nil {
		zap.S().Errorf("Failed to migrate analytics database: %v", err)
		return nil, nil, fmt.Errorf("failed to migrate analytics database: %w", err)
	}

	cleanup := func() {
		zap.S().Info("Closing analytics data layer...")
		// 可以在这里添加数据库连接关闭逻辑，如果 GORM 提供了的话
	}

	return d, cleanup, nil
}

// --- AnalyticsRepository 实现 ---

// analyticsRepository 是 AnalyticsRepository 接口的 GORM 实现。
type analyticsRepository struct {
	*Data
}

// NewAnalyticsRepository 创建一个新的 AnalyticsRepository 实例。
func NewAnalyticsRepository(data *Data) AnalyticsRepository {
	return &analyticsRepository{data}
}

// BatchInsertPageViewEvents 批量插入页面浏览事件。
// 使用 GORM 的 CreateInBatches 方法进行高效批量插入。
func (r *analyticsRepository) BatchInsertPageViewEvents(ctx context.Context, events []*model.PageViewEvent) error {
	if len(events) == 0 {
		return nil
	}
	dbEvents := make([]DBPageViewEvent, len(events))
	for i, event := range events {
		dbEvents[i] = *fromBizPageViewEvent(event)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(&dbEvents, 1000).Error; err != nil {
		zap.S().Errorf("Failed to batch insert page view events: %v", err)
		return fmt.Errorf("failed to batch insert page view events: %w", err)
	}
	zap.S().Infof("Successfully batch inserted %d page view events", len(events))
	return nil
}

// BatchInsertSalesFacts 批量插入销售事实数据。
// 使用 GORM 的 CreateInBatches 方法进行高效批量插入。
func (r *analyticsRepository) BatchInsertSalesFacts(ctx context.Context, facts []*model.SalesFact) error {
	if len(facts) == 0 {
		return nil
	}
	dbFacts := make([]DBSalesFact, len(facts))
	for i, fact := range facts {
		dbFacts[i] = *fromBizSalesFact(fact)
	}

	if err := r.db.WithContext(ctx).CreateInBatches(&dbFacts, 1000).Error; err != nil {
		zap.S().Errorf("Failed to batch insert sales facts: %v", err)
		return fmt.Errorf("failed to batch insert sales facts: %w", err)
	}
	zap.S().Infof("Successfully batch inserted %d sales facts", len(facts))
	return nil
}

// QueryTotalRevenue 查询指定时间范围内的总销售额。
// 直接执行原生 SQL 查询通常是 OLAP 数据库的最佳实践。
func (r *analyticsRepository) QueryTotalRevenue(ctx context.Context, startTime, endTime *time.Time) (float64, error) {
	var totalRevenue float64
	query := r.db.WithContext(ctx).Model(&DBSalesFact{}).Select("SUM(item_price * item_quantity)")

	if startTime != nil {
		query = query.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("event_time <= ?", *endTime)
	}

	if err := query.Scan(&totalRevenue).Error; err != nil {
		zap.S().Errorf("Failed to query total revenue: %v", err)
		return 0, fmt.Errorf("failed to query total revenue: %w", err)
	}
	zap.S().Debugf("Queried total revenue: %.2f from %v to %v", totalRevenue, startTime, endTime)
	return totalRevenue, nil
}

// QuerySalesByProductCategory 查询按商品分类划分的销售额。
func (r *analyticsRepository) QuerySalesByProductCategory(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error) {
	var results []struct {
		Category string
		Revenue  float64
	}

	query := r.db.WithContext(ctx).Model(&DBSalesFact{}).Select("product_category as Category, SUM(item_price * item_quantity) as Revenue").Group("product_category")

	if startTime != nil {
		query = query.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("event_time <= ?", *endTime)
	}

	if err := query.Scan(&results).Error; err != nil {
		zap.S().Errorf("Failed to query sales by product category: %v", err)
		return nil, fmt.Errorf("failed to query sales by product category: %w", err)
	}

	categorySales := make(map[string]float64)
	for _, res := range results {
		categorySales[res.Category] = res.Revenue
	}
	zap.S().Debugf("Queried sales by product category: %+v", categorySales)
	return categorySales, nil
}

// QuerySalesByProductBrand 查询按商品品牌划分的销售额。
func (r *analyticsRepository) QuerySalesByProductBrand(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error) {
	var results []struct {
		Brand   string
		Revenue float64
	}

	query := r.db.WithContext(ctx).Model(&DBSalesFact{}).Select("product_brand as Brand, SUM(item_price * item_quantity) as Revenue").Group("product_brand")

	if startTime != nil {
		query = query.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("event_time <= ?", *endTime)
	}

	if err := query.Scan(&results).Error; err != nil {
		zap.S().Errorf("Failed to query sales by product brand: %v", err)
		return nil, fmt.Errorf("failed to query sales by product brand: %w", err)
	}

	brandSales := make(map[string]float64)
	for _, res := range results {
		brandSales[res.Brand] = res.Revenue
	}
	zap.S().Debugf("Queried sales by product brand: %+v", brandSales)
	return brandSales, nil
}

// QueryUserActivityCount 查询指定时间范围内的活跃用户数。
// 活跃用户定义为在 PageViewEvent 中出现的用户。
func (r *analyticsRepository) QueryUserActivityCount(ctx context.Context, startTime, endTime *time.Time) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&DBPageViewEvent{}).Select("COUNT(DISTINCT user_id)")

	if startTime != nil {
		query = query.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("event_time <= ?", *endTime)
	}

	if err := query.Scan(&count).Error; err != nil {
		zap.S().Errorf("Failed to query user activity count: %v", err)
		return 0, fmt.Errorf("failed to query user activity count: %w", err)
	}
	zap.S().Debugf("Queried user activity count: %d from %v to %v", count, startTime, endTime)
	return count, nil
}

// QueryTopNProductsByRevenue 查询销售额最高的N个商品。
func (r *analyticsRepository) QueryTopNProductsByRevenue(ctx context.Context, n int, startTime, endTime *time.Time) ([]*model.ProductSales, error) {
	var results []*model.ProductSales

	query := r.db.WithContext(ctx).Model(&DBSalesFact{}).Select("product_id, product_name, SUM(item_price * item_quantity) as revenue").Group("product_id, product_name").Order("revenue DESC")

	if startTime != nil {
		query = query.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("event_time <= ?", *endTime)
	}

	if n > 0 {
		query = query.Limit(n)
	}

	if err := query.Scan(&results).Error; err != nil {
		zap.S().Errorf("Failed to query top N products by revenue: %v", err)
		return nil, fmt.Errorf("failed to query top N products by revenue: %w", err)
	}
	zap.S().Debugf("Queried top %d products by revenue: %+v", n, results)
	return results, nil
}

// QueryConversionRate 查询从页面浏览到下单的转化率。
// 这是一个简化的示例，实际转化率计算会更复杂。
func (r *analyticsRepository) QueryConversionRate(ctx context.Context, startTime, endTime *time.Time) (float64, error) {
	var pageViewsCount int64
	pageViewQuery := r.db.WithContext(ctx).Model(&DBPageViewEvent{}).Select("COUNT(DISTINCT user_id)")
	if startTime != nil {
		pageViewQuery = pageViewQuery.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		pageViewQuery = pageViewQuery.Where("event_time <= ?", *endTime)
	}
	if err := pageViewQuery.Scan(&pageViewsCount).Error; err != nil {
		zap.S().Errorf("Failed to count page views for conversion rate: %v", err)
		return 0, fmt.Errorf("failed to count page views: %w", err)
	}

	var ordersCount int64
	orderQuery := r.db.WithContext(ctx).Model(&DBSalesFact{}).Select("COUNT(DISTINCT user_id)")
	if startTime != nil {
		orderQuery = orderQuery.Where("event_time >= ?", *startTime)
	}
	if endTime != nil {
		orderQuery = orderQuery.Where("event_time <= ?", *endTime)
	}
	if err := orderQuery.Scan(&ordersCount).Error; err != nil {
		zap.S().Errorf("Failed to count orders for conversion rate: %v", err)
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	if pageViewsCount == 0 {
		return 0, nil // 避免除以零
	}

	conversionRate := float64(ordersCount) / float64(pageViewsCount)
	zap.S().Debugf("Calculated conversion rate: %.2f (orders: %d, page views: %d)", conversionRate, ordersCount, pageViewsCount)
	return conversionRate, nil
}

// --- 模型转换辅助函数 ---

// toBizPageViewEvent 将 DBPageViewEvent 数据库模型转换为 model.PageViewEvent 业务领域模型。
func toBizPageViewEvent(dbEvent *DBPageViewEvent) *model.PageViewEvent {
	if dbEvent == nil {
		return nil
	}
	return &model.PageViewEvent{
		EventTime: dbEvent.EventTime,
		UserID:    dbEvent.UserID,
		SessionID: dbEvent.SessionID,
		URL:       dbEvent.URL,
		Referer:   dbEvent.Referer,
		UserAgent: dbEvent.UserAgent,
		ClientIP:  dbEvent.ClientIP,
	}
}

// fromBizPageViewEvent 将 model.PageViewEvent 业务领域模型转换为 DBPageViewEvent 数据库模型。
func fromBizPageViewEvent(bizEvent *model.PageViewEvent) *DBPageViewEvent {
	if bizEvent == nil {
		return nil
	}
	return &DBPageViewEvent{
		EventTime: bizEvent.EventTime,
		UserID:    bizEvent.UserID,
		SessionID: bizEvent.SessionID,
		URL:       bizEvent.URL,
		Referer:   bizEvent.Referer,
		UserAgent: bizEvent.UserAgent,
		ClientIP:  bizEvent.ClientIP,
	}
}

// toBizSalesFact 将 DBSalesFact 数据库模型转换为 model.SalesFact 业务领域模型。
func toBizSalesFact(dbFact *DBSalesFact) *model.SalesFact {
	if dbFact == nil {
		return nil
	}
	return &model.SalesFact{
		EventTime:       dbFact.EventTime,
		OrderID:         dbFact.OrderID,
		OrderItemID:     dbFact.OrderItemID,
		OrderSN:         dbFact.OrderSN,
		OrderTotal:      dbFact.OrderTotal,
		DiscountAmount:  dbFact.DiscountAmount,
		ProductID:       dbFact.ProductID,
		ProductSKU:      dbFact.ProductSKU,
		ProductName:     dbFact.ProductName,
		ProductCategory: dbFact.ProductCategory,
		ProductBrand:    dbFact.ProductBrand,
		ItemPrice:       dbFact.ItemPrice,
		ItemQuantity:    dbFact.ItemQuantity,
		UserID:          dbFact.UserID,
		UserCity:        dbFact.UserCity,
		UserCountry:     dbFact.UserCountry,
	}
}

// fromBizSalesFact 将 model.SalesFact 业务领域模型转换为 DBSalesFact 数据库模型。
func fromBizSalesFact(bizFact *model.SalesFact) *DBSalesFact {
	if bizFact == nil {
		return nil
	}
	return &DBSalesFact{
		EventTime:       bizFact.EventTime,
		OrderID:         bizFact.OrderID,
		OrderItemID:     bizFact.OrderItemID,
		OrderSN:         bizFact.OrderSN,
		OrderTotal:      bizFact.OrderTotal,
		DiscountAmount:  bizFact.DiscountAmount,
		ProductID:       bizFact.ProductID,
		ProductSKU:      bizFact.ProductSKU,
		ProductName:     bizFact.ProductName,
		ProductCategory: bizFact.ProductCategory,
		ProductBrand:    bizFact.ProductBrand,
		ItemPrice:       bizFact.ItemPrice,
		ItemQuantity:    bizFact.ItemQuantity,
		UserID:          bizFact.UserID,
		UserCity:        bizFact.UserCity,
		UserCountry:     bizFact.UserCountry,
	}
}