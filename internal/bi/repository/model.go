package data

// SalesOverview represents aggregated sales data for BI reports.
type SalesOverview struct {
	TotalSalesAmount uint64  `ch:"total_sales_amount"`
	TotalOrders      uint32  `ch:"total_orders"`
	TotalUsers       uint32  `ch:"total_users"`
	ConversionRate   float64 `ch:"conversion_rate"`
}

// ProductSalesData represents sales data for a specific product.
type ProductSalesData struct {
	ProductID     uint64 `ch:"product_id"`
	ProductName   string `ch:"product_name"`
	SalesQuantity uint32 `ch:"sales_quantity"`
	SalesAmount   uint64 `ch:"sales_amount"`
}
