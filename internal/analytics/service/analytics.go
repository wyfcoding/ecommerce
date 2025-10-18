package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/analytics/model"
	"ecommerce/internal/analytics/repository"
)

// AnalyticsService 定义了分析服务的业务逻辑接口
type AnalyticsService interface {
	ProcessEvent(ctx context.Context, eventType string, payload []byte) error
	GetTotalRevenue(ctx context.Context) (float64, error)
}

// analyticsService 是接口的具体实现
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

// NewAnalyticsService 创建一个新的 analyticsService 实例
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

// ProcessEvent 路由并处理来自 MQ 的事件
func (s *analyticsService) ProcessEvent(ctx context.Context, eventType string, payload []byte) error {
	s.logger.Debug("Processing event", zap.String("eventType", eventType))

	// 简单的事件路由
	switch eventType {
	case "order.created":
		return s.handleOrderCreated(payload)
	case "user.viewed_page":
		return s.handlePageView(payload)
	default:
		// 忽略不关心的事件
		return nil
	}
}

// GetTotalRevenue ...
func (s *analyticsService) GetTotalRevenue(ctx context.Context) (float64, error) {
	return s.repo.QueryTotalRevenue(ctx)
}

// --- 内部事件处理方法 ---

func (s *analyticsService) handleOrderCreated(payload []byte) error {
	// 1. 解析事件
	var orderData map[string]interface{} // 实际应为强类型结构体
	if err := json.Unmarshal(payload, &orderData); err != nil {
		return err
	}

	// 2. 将事件转换为一个或多个 SalesFact
	// 在真实场景中，需要从 payload 中获取所有必要信息
	fact := &model.SalesFact{
		EventTime: time.Now(),
		// ... 从 orderData 中填充字段
	}

	// 3. 添加到批处理队列
	s.addSalesFact(fact)
	return nil
}

func (s *analyticsService) handlePageView(payload []byte) error {
	var pageViewData model.PageViewEvent
	if err := json.Unmarshal(payload, &pageViewData); err != nil {
		return err
	}
	s.addPageView(&pageViewData)
	return nil
}

// --- 批处理逻辑 ---

func (s *analyticsService) addSalesFact(fact *model.SalesFact) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.salesFacts = append(s.salesFacts, fact)
	if len(s.salesFacts) >= s.batchSize {
		s.flushSalesFacts()
	}
}

func (s *analyticsService) addPageView(event *model.PageViewEvent) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pageViews = append(s.pageViews, event)
	if len(s.pageViews) >= s.batchSize {
		s.flushPageViews()
	}
}

// flushSalesFacts 将销售事实数据刷入数据库
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

// flushPageViews 将页面浏览数据刷入数据库
func (s *analyticsService) flushPageViews() {
	// ... 与 flushSalesFacts 类似 ...
}

// runBatchProcessor 是一个后台 goroutine，用于按时间定期刷数据
func (s *analyticsService) runBatchProcessor() {
	for range s.flushTicker.C {
		s.mutex.Lock()
		s.flushSalesFacts()
		s.flushPageViews()
		s.mutex.Unlock()
	}
}