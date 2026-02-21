package application

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/reportcenter/v1"
	"github.com/wyfcoding/ecommerce/internal/reportcenter/domain"
	"github.com/wyfcoding/pkg/idgen"
	"google.golang.org/protobuf/types/known/timestamppb"
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
func (s *ReportService) GetSalesReport(ctx context.Context, req *pb.GetSalesReportRequest) (*pb.SalesReport, error) {
	start, end, err := normalizeRange(req.GetStartTime(), req.GetEndTime())
	if err != nil {
		return nil, err
	}
	summaries, err := s.repo.GetSalesSummary(ctx, start, end)
	if err != nil {
		return nil, err
	}
	interval := normalizeInterval(req.GetInterval())

	type aggregate struct {
		revenue float64
		orders  int64
	}
	buckets := make(map[time.Time]*aggregate, len(summaries))
	for _, summary := range summaries {
		if summary == nil {
			continue
		}
		bucketAt := truncateByInterval(summary.Date, interval)
		item, exists := buckets[bucketAt]
		if !exists {
			item = &aggregate{}
			buckets[bucketAt] = item
		}
		item.revenue += summary.TotalRevenue
		item.orders += summary.TotalOrders
	}

	points := make([]*pb.SalesDataPoint, 0, len(buckets))
	ordered := make([]time.Time, 0, len(buckets))
	for ts := range buckets {
		ordered = append(ordered, ts)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Before(ordered[j]) })

	var totalRevenue float64
	var totalOrders int64
	for _, ts := range ordered {
		item := buckets[ts]
		totalRevenue += item.revenue
		totalOrders += item.orders
		points = append(points, &pb.SalesDataPoint{
			Timestamp: timestamppb.New(ts),
			Revenue:   item.revenue,
			Orders:    item.orders,
		})
	}

	aov := 0.0
	if totalOrders > 0 {
		aov = totalRevenue / float64(totalOrders)
	}

	return &pb.SalesReport{
		TotalRevenue:      totalRevenue,
		TotalOrders:       totalOrders,
		AverageOrderValue: aov,
		DataPoints:        points,
	}, nil
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

func (s *ReportService) GetLowStockAlerts(ctx context.Context) ([]*pb.LowStockAlert, error) {
	alerts, err := s.repo.GetLowStockAlerts(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]*pb.LowStockAlert, 0, len(alerts))
	for _, item := range alerts {
		if item == nil {
			continue
		}
		resp = append(resp, &pb.LowStockAlert{
			ProductId:    item.SKUID, // 当前模型未区分 product_id，先按 sku 透传
			SkuId:        item.SKUID,
			Name:         item.SKUName,
			CurrentStock: item.CurrentStock,
			Threshold:    10,
		})
	}
	return resp, nil
}

func (s *ReportService) GetFinancialSummary(ctx context.Context, req *pb.GetFinancialSummaryRequest) (*pb.FinancialSummary, error) {
	start, end, err := normalizeRange(req.GetStartTime(), req.GetEndTime())
	if err != nil {
		return nil, err
	}
	metric, err := s.repo.GetFinancialMetrics(ctx, start, end)
	if err != nil {
		return nil, err
	}
	if metric == nil {
		return &pb.FinancialSummary{}, nil
	}

	discounts := metric.GMV - metric.NetRevenue
	if discounts < 0 {
		discounts = 0
	}

	return &pb.FinancialSummary{
		GrossMerchandiseValue: metric.GMV,
		TotalDiscounts:        discounts,
		NetRevenue:            metric.NetRevenue,
		TotalFees:             metric.TotalFees,
		RefundAmount:          metric.Refunds,
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

	// 异步生成：先置处理中，再模拟生成完成并产出下载地址。
	go func(base *domain.CustomReport) {
		bg := context.Background()
		processing := *base
		processing.Status = pb.ReportStatus_REPORT_STATUS_PROCESSING
		if err := s.repo.SaveCustomReport(bg, &processing); err != nil {
			s.logger.Error("failed to mark custom report processing", "report_id", base.ID, "error", err)
			return
		}

		time.Sleep(100 * time.Millisecond)

		completed := processing
		completed.Status = pb.ReportStatus_REPORT_STATUS_COMPLETED
		completed.DownloadURL = fmt.Sprintf("/reports/download/%s", completed.ID)
		now := time.Now()
		completed.CompletedAt = &now
		if err := s.repo.SaveCustomReport(bg, &completed); err != nil {
			s.logger.Error("failed to complete custom report", "report_id", base.ID, "error", err)
		}
	}(report)

	return report, nil
}

func normalizeRange(startTS, endTS *timestamppb.Timestamp) (time.Time, time.Time, error) {
	end := time.Now()
	if endTS != nil {
		end = endTS.AsTime()
	}
	start := end.AddDate(0, 0, -30)
	if startTS != nil {
		start = startTS.AsTime()
	}
	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range: start must be before end")
	}
	return start, end, nil
}

func normalizeInterval(interval string) string {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "hour", "day", "week", "month":
		return strings.ToLower(strings.TrimSpace(interval))
	default:
		return "day"
	}
}

func truncateByInterval(t time.Time, interval string) time.Time {
	switch interval {
	case "hour":
		return t.Truncate(time.Hour)
	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		dayStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		return dayStart.AddDate(0, 0, -(weekday - 1))
	case "month":
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
}
