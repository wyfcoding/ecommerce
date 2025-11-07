package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/analytics/model"
	"ecommerce/internal/analytics/repository"
)

var (
	// ErrInvalidTimeRange 表示时间范围无效。
	ErrInvalidTimeRange = errors.New("invalid time range")
)

// AnalyticsService 定义了分析服务的业务逻辑接口。
// 包含了事件处理、各种销售和用户行为分析查询功能。
type AnalyticsService interface {
	// ProcessEvent 路由并处理来自消息队列的事件。
	ProcessEvent(ctx context.Context, eventType string, payload []byte) error

	// GetTotalRevenue 查询指定时间范围内的总销售额。
	GetTotalRevenue(ctx context.Context, startTime, endTime *time.Time) (float64, error)
	// GetSalesByProductCategory 查询按商品分类划分的销售额。
	GetSalesByProductCategory(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error)
	// GetSalesByProductBrand 查询按商品品牌划分的销售额。
	GetSalesByProductBrand(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error)
	// GetUserActivityCount 查询指定时间范围内的活跃用户数。
	GetUserActivityCount(ctx context.Context, startTime, endTime *time.Time) (int64, error)
	// GetTopNProductsByRevenue 查询销售额最高的N个商品。
	GetTopNProductsByRevenue(ctx context.Context, n int, startTime, endTime *time.Time) ([]*model.ProductSales, error)
	// GetConversionRate 查询从页面浏览到下单的转化率。
	GetConversionRate(ctx context.Context, startTime, endTime *time.Time) (float64, error)
}

// analyticsService 是 AnalyticsService 接口的具体实现。
// 它负责事件的路由、批处理写入以及各种分析查询的业务逻辑。
type analyticsService struct {
	repo        repository.AnalyticsRepository
	logger      *zap.Logger

	// 用于批量写入的内部状态
	mutex       sync.Mutex
	pageViews   []*model.PageViewEvent
	salesFacts  []*model.SalesFact
	batchSize   int
	flushTicker *time.Ticker
}

// NewAnalyticsService 创建一个新的 analyticsService 实例。
// 接收 AnalyticsRepository 和 zap.Logger 实例，并启动后台批处理器。
func NewAnalyticsService(repo repository.AnalyticsRepository, logger *zap.Logger) AnalyticsService {
	s := &analyticsService{
		repo:        repo,
		logger:      logger,
		pageViews:   make([]*model.PageViewEvent, 0, 100),
		salesFacts:  make([]*model.SalesFact, 0, 100),
		batchSize:   100, // 每 100 个事件刷一次
		flushTicker: time.NewTicker(10 * time.Second), // 或每 10 秒刷一次
	}
	go s.runBatchProcessor() // 启动后台批处理器
	return s
}

// ProcessEvent 路由并处理来自消息队列的事件。
// 根据事件类型，将事件数据解析并添加到相应的批处理队列中。
func (s *analyticsService) ProcessEvent(ctx context.Context, eventType string, payload []byte) error {
	s.logger.Debug("Processing event", zap.String("event_type", eventType))

	// 简单的事件路由
	switch eventType {
	case "order.created":
		return s.handleOrderCreated(payload)
	case "user.viewed_page":
		return s.handlePageView(payload)
	default:
		s.logger.Debug("Ignoring unknown event type", zap.String("event_type", eventType))
		return nil
	}
}

// GetTotalRevenue 查询指定时间范围内的总销售额。
func (s *analyticsService) GetTotalRevenue(ctx context.Context, startTime, endTime *time.Time) (float64, error) {
	s.logger.Info("Querying total revenue", zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QueryTotalRevenue(ctx, startTime, endTime)
}

// GetSalesByProductCategory 查询按商品分类划分的销售额。
func (s *analyticsService) GetSalesByProductCategory(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error) {
	s.logger.Info("Querying sales by product category", zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QuerySalesByProductCategory(ctx, startTime, endTime)
}

// GetSalesByProductBrand 查询按商品品牌划分的销售额。
func (s *analyticsService) GetSalesByProductBrand(ctx context.Context, startTime, endTime *time.Time) (map[string]float64, error) {
	s.logger.Info("Querying sales by product brand", zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QuerySalesByProductBrand(ctx, startTime, endTime)
}

// GetUserActivityCount 查询指定时间范围内的活跃用户数。
func (s *analyticsService) GetUserActivityCount(ctx context.Context, startTime, endTime *time.Time) (int64, error) {
	s.logger.Info("Querying user activity count", zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QueryUserActivityCount(ctx, startTime, endTime)
}

// GetTopNProductsByRevenue 查询销售额最高的N个商品。
func (s *analyticsService) GetTopNProductsByRevenue(ctx context.Context, n int, startTime, endTime *time.Time) ([]*model.ProductSales, error) {
	s.logger.Info("Querying top N products by revenue", zap.Int("n", n), zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QueryTopNProductsByRevenue(ctx, n, startTime, endTime)
}

// GetConversionRate 查询从页面浏览到下单的转化率。
func (s *analyticsService) GetConversionRate(ctx context.Context, startTime, endTime *time.Time) (float64, error) {
	s.logger.Info("Querying conversion rate", zap.Any("start_time", startTime), zap.Any("end_time", endTime))
	return s.repo.QueryConversionRate(ctx, startTime, endTime)
}

// --- 内部事件处理方法 ---

// handleOrderCreated 处理订单创建事件。
// 它将订单数据解析为 SalesFact 并添加到批处理队列。
func (s *analyticsService) handleOrderCreated(payload []byte) error {
	// 1. 解析事件
	var orderData struct {
		OrderID        uint64    `json:"order_id"`
		OrderSN        string    `json:"order_sn"`
		UserID         uint64    `json:"user_id"`
		OrderTotal     float64   `json:"order_total"`
		DiscountAmount float64   `json:"discount_amount"`
		CreatedAt      time.Time `json:"created_at"`
		Items          []struct {
			OrderItemID     uint64  `json:"order_item_id"`
			ProductID       uint64  `json:"product_id"`
			ProductSKU      string  `json:"product_sku"`
			ProductName     string  `json:"product_name"`
			ProductCategory string  `json:"product_category"`
			ProductBrand    string  `json:"product_brand"`
			ItemPrice       float64 `json:"item_price"`
			ItemQuantity    int     `json:"item_quantity"`
		} `json:"items"`
	} // 实际应为强类型结构体
	if err := json.Unmarshal(payload, &orderData); err != nil {
		s.logger.Error("Failed to unmarshal order created event payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal order created event: %w", err)
	}

	// 2. 将事件转换为一个或多个 SalesFact
	for _, item := range orderData.Items {
		fact := &model.SalesFact{
			EventTime:       orderData.CreatedAt,
			OrderID:         uint(orderData.OrderID),
			OrderItemID:     uint(item.OrderItemID),
			OrderSN:         orderData.OrderSN,
			OrderTotal:      orderData.OrderTotal,
			DiscountAmount:  orderData.DiscountAmount,
			ProductID:       uint(item.ProductID),
			ProductSKU:      item.ProductSKU,
			ProductName:     item.ProductName,
			ProductCategory: item.ProductCategory,
			ProductBrand:    item.ProductBrand,
			ItemPrice:       item.ItemPrice,
			ItemQuantity:    item.ItemQuantity,
			UserID:          uint(orderData.UserID),
			// UserCity, UserCountry 需要从用户服务获取或从事件中带入
		}
		s.addSalesFact(fact)
	}
	return nil
}

// handlePageView 处理页面浏览事件。
// 它将页面浏览数据解析为 PageViewEvent 并添加到批处理队列。
func (s *analyticsService) handlePageView(payload []byte) error {
	var pageViewData model.PageViewEvent
	if err := json.Unmarshal(payload, &pageViewData); err != nil {
		s.logger.Error("Failed to unmarshal page view event payload", zap.Error(err))
		return fmt.Errorf("failed to unmarshal page view event: %w", err)
	}
	s.addPageView(&pageViewData)
	return nil
}

// --- 批处理逻辑 ---

// addSalesFact 将销售事实添加到批处理队列。
// 当队列达到批处理大小时，会触发一次刷新操作。
func (s *analyticsService) addSalesFact(fact *model.SalesFact) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.salesFacts = append(s.salesFacts, fact)
	if len(s.salesFacts) >= s.batchSize {
		s.flushSalesFacts()
	}
}

// addPageView 将页面浏览事件添加到批处理队列。
// 当队列达到批处理大小时，会触发一次刷新操作。
func (s *analyticsService) addPageView(event *model.PageViewEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pageViews = append(s.pageViews, event)
	if len(s.pageViews) >= s.batchSize {
		s.flushPageViews()
	}
}

// flushSalesFacts 将销售事实数据刷入数据库。
// 在一个单独的 goroutine 中执行，以避免阻塞。
func (s *analyticsService) flushSalesFacts() {
	if len(s.salesFacts) == 0 {
		return
	}
	batch := s.salesFacts
	s.salesFacts = make([]*model.SalesFact, 0, s.batchSize)

	go func(b []*model.SalesFact) {
		if err := s.repo.BatchInsertSalesFacts(context.Background(), b); err != nil {
			s.logger.Error("Failed to flush sales facts", zap.Error(err))
			// TODO: 添加重试或死信队列逻辑
		}
		s.logger.Info("Flushed sales facts to DB", zap.Int("count", len(b)))
	}(batch)
}

// flushPageViews 将页面浏览数据刷入数据库。
// 在一个单独的 goroutine 中执行，以避免阻塞。
func (s *analyticsService) flushPageViews() {
	if len(s.pageViews) == 0 {
		return
	}
	batch := s.pageViews
	s.pageViews = make([]*model.PageViewEvent, 0, s.batchSize)

	go func(b []*model.PageViewEvent) {
		if err := s.repo.BatchInsertPageViewEvents(context.Background(), b); err != nil {
			s.logger.Error("Failed to flush page views", zap.Error(err))
			// TODO: 添加重试或死信队列逻辑
		}
		s.logger.Info("Flushed page views to DB", zap.Int("count", len(b)))
	}(batch)
}

// runBatchProcessor 是一个后台 goroutine，用于按时间定期刷数据。
// 它会定期触发 flush 操作，确保数据不会长时间停留在内存中。
func (s *analyticsService) runBatchProcessor() {
	for range s.flushTicker.C {
		s.mutex.Lock()
		s.flushSalesFacts()
		s.flushPageViews()
		s.mutex.Unlock()
	}
}
