package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/analytics/model"
)

// AnalyticsRepository 定义了分析数据仓库的接口
type AnalyticsRepository interface {
	// 写入操作
	BatchInsertPageViewEvents(ctx context.Context, events []*model.PageViewEvent) error
	BatchInsertSalesFacts(ctx context.Context, facts []*model.SalesFact) error

	// 读取/查询操作 (示例)
	QueryTotalRevenue(ctx context.Context) (float64, error)
}

// analyticsRepository 是接口的具体实现
// 它使用 GORM，但实际的 ClickHouse 驱动可能需要更原生的交互方式以获得最佳性能
type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository 创建一个新的 analyticsRepository 实例
func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// BatchInsertPageViewEvents 批量插入页面浏览事件
func (r *analyticsRepository) BatchInsertPageViewEvents(ctx context.Context, events []*model.PageViewEvent) error {
	if len(events) == 0 {
		return nil
	}
	// GORM 的 CreateBatch 方法或原生驱动的批量插入 API 会非常高效
	if err := r.db.WithContext(ctx).Create(&events).Error; err != nil {
		return fmt.Errorf("数据库批量插入页面浏览事件失败: %w", err)
	}
	return nil
}

// BatchInsertSalesFacts 批量插入销售事实数据
func (r *analyticsRepository) BatchInsertSalesFacts(ctx context.Context, facts []*model.SalesFact) error {
	if len(facts) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&facts).Error; err != nil {
		return fmt.Errorf("数据库批量插入销售事实失败: %w", err)
	}
	return nil
}

// QueryTotalRevenue 查询总销售额的示例
func (r *analyticsRepository) QueryTotalRevenue(ctx context.Context) (float64, error) {
	var totalRevenue struct {
		Total float64
	}

	// 直接执行原生 SQL 查询通常是 OLAP 数据库的最佳实践
	result := r.db.WithContext(ctx).Raw("SELECT SUM(item_price * item_quantity) as total FROM sales_facts").Scan(&totalRevenue)
	if result.Error != nil {
		return 0, fmt.Errorf("数据库查询总销售额失败: %w", result.Error)
	}

	return totalRevenue.Total, nil
}
