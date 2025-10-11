package data

import (
	"context"
	"ecommerce/internal/bi/biz"
	"ecommerce/internal/bi/data/model"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type biRepo struct {
	data     *Data
	chClient clickhouse.Conn
}

// NewBiRepo creates a new BiRepo.
func NewBiRepo(data *Data, chClient clickhouse.Conn) biz.BiRepo {
	return &biRepo{data: data, chClient: chClient}
}

// GetSalesOverview retrieves sales overview data from ClickHouse.
func (r *biRepo) GetSalesOverview(ctx context.Context, startDate, endDate *time.Time) (*biz.SalesOverview, error) {
	query := `
		SELECT
			sum(total_amount) AS total_sales_amount,
			count(DISTINCT order_id) AS total_orders,
			count(DISTINCT user_id) AS total_users,
			if(count(DISTINCT user_id) > 0, count(DISTINCT order_id) / count(DISTINCT user_id), 0) AS conversion_rate
		FROM orders_table
		WHERE created_at >= ? AND created_at <= ?
	`
	// orders_table is a placeholder for a ClickHouse table that stores order data.
	// In a real system, this would be populated by Kafka consumers from order service.

	var overview model.SalesOverview
	err := r.chClient.QueryRow(ctx, query, startDate, endDate).Scan(
		&overview.TotalSalesAmount,
		&overview.TotalOrders,
		&overview.TotalUsers,
		&overview.ConversionRate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales overview from ClickHouse: %w", err)
	}

	return &biz.SalesOverview{
		TotalSalesAmount: overview.TotalSalesAmount,
		TotalOrders:      overview.TotalOrders,
		TotalUsers:       overview.TotalUsers,
		ConversionRate:   overview.ConversionRate,
	}, nil
}

// GetTopSellingProducts retrieves top selling products from ClickHouse.
func (r *biRepo) GetTopSellingProducts(ctx context.Context, limit uint32, startDate, endDate *time.Time) ([]*biz.ProductSalesData, error) {
	query := `
		SELECT
			product_id,
			product_name, -- Assuming product_name is available in order_items_table
			sum(quantity) AS sales_quantity,
			sum(price * quantity) AS sales_amount
		FROM order_items_table
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY product_id, product_name
		ORDER BY sales_quantity DESC
		LIMIT ?
	`
	// order_items_table is a placeholder for a ClickHouse table that stores order item data.

	rows := r.chClient.QueryRow(ctx, query, startDate, endDate, limit)

	var products []*model.ProductSalesData
	for rows.Next() {
		var product model.ProductSalesData
		err := rows.Scan(
			&product.ProductID,
			&product.ProductName,
			&product.SalesQuantity,
			&product.SalesAmount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan top selling product row: %w", err)
		}
		products = append(products, &biz.ProductSalesData{
			ProductID:     product.ProductID,
			ProductName:   product.ProductName,
			SalesQuantity: product.SalesQuantity,
			SalesAmount:   product.SalesAmount,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating top selling products rows: %w", err)
	}

	return products, nil
}
