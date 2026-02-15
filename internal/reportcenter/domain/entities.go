package domain

import (
	"context"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/reportcenter/v1"
)

// SalesSummary 销售汇总模型.
type SalesSummary struct {
	ID                string
	Date              time.Time
	TotalRevenue      float64
	TotalOrders       int64
	AverageOrderValue float64
}

// ProductPerformance 核心商品性能指标.
type ProductPerformance struct {
	ProductID   string
	ProductName string
	UnitsSold   int64
	Revenue     float64
}

// InventorySnapshot 库存快照模型.
type InventorySnapshot struct {
	SKUID         string
	SKUName       string
	CurrentStock  int64
	LastUpdatedAt time.Time
}

// FinancialMetric 财务指标模型.
type FinancialMetric struct {
	GMV        float64
	NetRevenue float64
	TotalFees  float64
	Refunds    float64
	MetricDate time.Time
}

// CustomReport 异步生成的自定义报表。
type CustomReport struct {
	ID          string
	Name        string
	Type        string
	Status      pb.ReportStatus
	DownloadURL string
	ErrorMsg    string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// ReportRepository 定义报表数据访问接口。
type ReportRepository interface {
	// Sales & Financial
	GetSalesSummary(ctx context.Context, start, end time.Time) ([]*SalesSummary, error)
	GetTopProducts(ctx context.Context, start, end time.Time, topN int) ([]*ProductPerformance, error)
	GetFinancialMetrics(ctx context.Context, start, end time.Time) (*FinancialMetric, error)

	// Inventory
	GetInventoryHealth(ctx context.Context) (total, low, out int64, val float64, err error)
	GetLowStockAlerts(ctx context.Context) ([]*InventorySnapshot, error)

	// Custom
	SaveCustomReport(ctx context.Context, report *CustomReport) error
	GetCustomReport(ctx context.Context, id string) (*CustomReport, error)

	// Update logic (for event handlers)
	UpdateSalesMetrics(ctx context.Context, amount float64, count int64, date time.Time) error
	UpdateProductPerformance(ctx context.Context, productID, name string, units int64, revenue float64, date time.Time) error
	UpdateInventorySnapshot(ctx context.Context, skuID, name string, stock int64) error
}
