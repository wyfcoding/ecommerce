package mysql

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/reportcenter/v1"
	"github.com/wyfcoding/ecommerce/internal/reportcenter/domain"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DailySalesReportModel 每日销售聚合数据.
type DailySalesReportModel struct {
	gorm.Model
	ReportDate   time.Time `gorm:"column:report_date;uniqueIndex;type:date;not null;comment:报表日期"`
	TotalRevenue float64   `gorm:"column:total_revenue;type:decimal(18,2);default:0;comment:总收入"`
	TotalOrders  int64     `gorm:"column:total_orders;default:0;comment:订单总数"`
}

func (DailySalesReportModel) TableName() string { return "report_daily_sales" }

// ProductPerformanceModel 商品表现聚合数据.
type ProductPerformanceModel struct {
	gorm.Model
	ProductID   string    `gorm:"column:product_id;uniqueIndex:idx_prod_date;type:varchar(64);not null"`
	ProductName string    `gorm:"column:product_name;type:varchar(255)"`
	ReportDate  time.Time `gorm:"column:report_date;uniqueIndex:idx_prod_date;type:date;not null"`
	UnitsSold   int64     `gorm:"column:units_sold;default:0"`
	Revenue     float64   `gorm:"column:revenue;type:decimal(18,2);default:0"`
}

func (ProductPerformanceModel) TableName() string { return "report_product_performance" }

// InventorySnapshotModel 库存当前快照.
type InventorySnapshotModel struct {
	gorm.Model
	SKUID        string `gorm:"column:sku_id;uniqueIndex;type:varchar(64);not null"`
	SKUName      string `gorm:"column:sku_name;type:varchar(255)"`
	CurrentStock int64  `gorm:"column:current_stock;default:0"`
}

func (InventorySnapshotModel) TableName() string { return "report_inventory_snapshots" }

// CustomReportModel 自定义报表记录.
type CustomReportModel struct {
	gorm.Model
	ReportID    string     `gorm:"column:report_id;uniqueIndex;type:varchar(64);not null"`
	Name        string     `gorm:"column:name;type:varchar(255)"`
	Type        string     `gorm:"column:report_type;type:varchar(50)"`
	Status      int32      `gorm:"column:status;index"`
	DownloadURL string     `gorm:"column:download_url;type:varchar(512)"`
	ErrorMsg    string     `gorm:"column:error_msg;type:text"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
}

func (CustomReportModel) TableName() string { return "report_custom_jobs" }

type reportRepository struct {
	db     *database.DB
	logger *logging.Logger
}

func NewReportRepository(db *database.DB, logger *logging.Logger) domain.ReportRepository {
	return &reportRepository{db: db, logger: logger}
}

// domain.ReportRepository 实现
func (r *reportRepository) GetSalesSummary(ctx context.Context, start, end time.Time) ([]*domain.SalesSummary, error) {
	var models []*DailySalesReportModel
	if err := r.db.RawDB().WithContext(ctx).Where("report_date BETWEEN ? AND ?", start, end).Order("report_date ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.SalesSummary, len(models))
	for i, m := range models {
		aov := 0.0
		if m.TotalOrders > 0 {
			aov = m.TotalRevenue / float64(m.TotalOrders)
		}
		res[i] = &domain.SalesSummary{
			ID:                fmt.Sprintf("%d", m.ID),
			Date:              m.ReportDate,
			TotalRevenue:      m.TotalRevenue,
			TotalOrders:       m.TotalOrders,
			AverageOrderValue: aov,
		}
	}
	return res, nil
}

func (r *reportRepository) GetTopProducts(ctx context.Context, start, end time.Time, topN int) ([]*domain.ProductPerformance, error) {
	var results []struct {
		ProductID    string
		ProductName  string
		TotalUnits   int64
		TotalRevenue float64
	}
	err := r.db.RawDB().WithContext(ctx).Model(&ProductPerformanceModel{}).
		Select("product_id, product_name, SUM(units_sold) as total_units, SUM(revenue) as total_revenue").
		Where("report_date BETWEEN ? AND ?", start, end).
		Group("product_id, product_name").
		Order("total_revenue DESC").
		Limit(topN).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	res := make([]*domain.ProductPerformance, len(results))
	for i, m := range results {
		res[i] = &domain.ProductPerformance{
			ProductID:   m.ProductID,
			ProductName: m.ProductName,
			UnitsSold:   m.TotalUnits,
			Revenue:     m.TotalRevenue,
		}
	}
	return res, nil
}

func (r *reportRepository) GetFinancialMetrics(ctx context.Context, start, end time.Time) (*domain.FinancialMetric, error) {
	var res struct {
		GMV float64
	}
	// 简化：这里从销售摘要报表中获取 GMV
	err := r.db.RawDB().WithContext(ctx).Model(&DailySalesReportModel{}).
		Select("SUM(total_revenue) as gmv").
		Where("report_date BETWEEN ? AND ?", start, end).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}
	return &domain.FinancialMetric{
		GMV:        res.GMV,
		NetRevenue: res.GMV, // 示例简化
	}, nil
}

func (r *reportRepository) GetInventoryHealth(ctx context.Context) (total, low, out int64, val float64, err error) {
	r.db.RawDB().WithContext(ctx).Model(&InventorySnapshotModel{}).Count(&total)
	r.db.RawDB().WithContext(ctx).Model(&InventorySnapshotModel{}).Where("current_stock = 0").Count(&out)
	r.db.RawDB().WithContext(ctx).Model(&InventorySnapshotModel{}).Where("current_stock > 0 AND current_stock < 10").Count(&low)
	return
}

func (r *reportRepository) GetLowStockAlerts(ctx context.Context) ([]*domain.InventorySnapshot, error) {
	var models []*InventorySnapshotModel
	if err := r.db.RawDB().WithContext(ctx).Where("current_stock < 10").Find(&models).Error; err != nil {
		return nil, err
	}
	res := make([]*domain.InventorySnapshot, len(models))
	for i, m := range models {
		res[i] = &domain.InventorySnapshot{
			SKUID:        m.SKUID,
			SKUName:      m.SKUName,
			CurrentStock: m.CurrentStock,
		}
	}
	return res, nil
}

func (r *reportRepository) SaveCustomReport(ctx context.Context, report *domain.CustomReport) error {
	m := &CustomReportModel{
		ReportID:    report.ID,
		Name:        report.Name,
		Type:        report.Type,
		Status:      int32(report.Status),
		DownloadURL: report.DownloadURL,
		ErrorMsg:    report.ErrorMsg,
		CompletedAt: report.CompletedAt,
	}
	return r.db.RawDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "report_id"}},
		UpdateAll: true,
	}).Save(m).Error
}

func (r *reportRepository) GetCustomReport(ctx context.Context, id string) (*domain.CustomReport, error) {
	var m CustomReportModel
	if err := r.db.RawDB().WithContext(ctx).Where("report_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &domain.CustomReport{
		ID:          m.ReportID,
		Name:        m.Name,
		Type:        m.Type,
		Status:      pb.ReportStatus(m.Status),
		DownloadURL: m.DownloadURL,
		ErrorMsg:    m.ErrorMsg,
		CreatedAt:   m.CreatedAt,
		CompletedAt: m.CompletedAt,
	}, nil
}

// 事件驱动更新实现
func (r *reportRepository) UpdateSalesMetrics(ctx context.Context, amount float64, count int64, date time.Time) error {
	reportDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	return r.db.RawDB().WithContext(ctx).Model(&DailySalesReportModel{}).
		Where("report_date = ?", reportDate).
		Updates(map[string]interface{}{
			"total_revenue": gorm.Expr("total_revenue + ?", amount),
			"total_orders":  gorm.Expr("total_orders + ?", count),
		}).Error
}

func (r *reportRepository) UpdateProductPerformance(ctx context.Context, productID, name string, units int64, revenue float64, date time.Time) error {
	reportDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	m := &ProductPerformanceModel{
		ProductID:   productID,
		ProductName: name,
		ReportDate:  reportDate,
		UnitsSold:   units,
		Revenue:     revenue,
	}
	return r.db.RawDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "product_id"}, {Name: "report_date"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"units_sold": gorm.Expr("units_sold + ?", units),
			"revenue":    gorm.Expr("revenue + ?", revenue),
		}),
	}).Create(m).Error
}

func (r *reportRepository) UpdateInventorySnapshot(ctx context.Context, skuID, name string, stock int64) error {
	m := &InventorySnapshotModel{
		SKUID:        skuID,
		SKUName:      name,
		CurrentStock: stock,
	}
	return r.db.RawDB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku_id"}},
		UpdateAll: true,
	}).Save(m).Error
}
