package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/reportcenter/v1"
	"github.com/wyfcoding/ecommerce/internal/reportcenter/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type ReportService struct {
	repo   domain.ReportRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

func NewReportService(repo domain.ReportRepository, idGen idgen.Generator, logger *slog.Logger) *ReportService {
	return &ReportService{
		repo:   repo,
		idGen:  idGen,
		logger: logger.With("service", "report_application"),
	}
}

// Queries
func (s *ReportService) GetSalesReport(ctx context.Context, req *pb.GetSalesReportRequest) (*domain.SalesSummary, error) {
	// start := req.StartTime.AsTime()
	// end := req.EndTime.AsTime()
	// summaries, err := s.repo.GetSalesSummary(ctx, start, end)
	// if err != nil {
	// 	return nil, err
	// }

	// 这里可以进一步聚合或处理维度（按天/按小时）
	// 为简化，直接返回汇总数据（实际场景需按 interval 分组）
	return nil, nil // TODO: 实现具体聚合逻辑
}

func (s *ReportService) GetInventoryHealth(ctx context.Context) (*pb.InventoryHealthReport, error) {
	total, low, out, val, err := s.repo.GetInventoryHealth(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.InventoryHealthReport{
		TotalSkus:           total,
		LowStockCount:       low,
		OutOfStockCount:     out,
		TotalInventoryValue: val,
	}, nil
}

// Event Handlers (Called by MQ Consumer)
func (s *ReportService) HandleOrderCreated(ctx context.Context, orderID string, amount float64, items []OrderItem) error {
	now := time.Now()
	// 更新整体销售额
	if err := s.repo.UpdateSalesMetrics(ctx, amount, 1, now); err != nil {
		return err
	}
	// 更新商品表现
	for _, item := range items {
		if err := s.repo.UpdateProductPerformance(ctx, item.ProductID, item.Name, item.Quantity, item.Price*float64(item.Quantity), now); err != nil {
			s.logger.Error("failed to update product performance", "error", err, "prod_id", item.ProductID)
		}
	}
	return nil
}

func (s *ReportService) HandleInventoryChanged(ctx context.Context, skuID string, name string, newStock int64) error {
	return s.repo.UpdateInventorySnapshot(ctx, skuID, name, newStock)
}

// Helper Types
type OrderItem struct {
	ProductID string
	Name      string
	Quantity  int64
	Price     float64
}

// Custom Report Job Placeholder
func (s *ReportService) CreateCustomReport(ctx context.Context, req *pb.CreateCustomReportRequest) (*domain.CustomReport, error) {
	id := s.idGen.Generate()
	report := &domain.CustomReport{
		ID:        fmt.Sprintf("rep_%d", id),
		Name:      req.Name,
		Type:      req.ReportType,
		Status:    pb.ReportStatus_REPORT_STATUS_PENDING,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveCustomReport(ctx, report); err != nil {
		return nil, err
	}
	// TODO: 异步触发报表生成任务（如使用 temporal, delay queue 或简单 goroutine）
	return report, nil
}
