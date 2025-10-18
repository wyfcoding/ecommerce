package model

// SalesOverview represents aggregated sales data for BI reports in the business logic layer.
type SalesOverview struct {
	TotalSalesAmount uint64
	TotalOrders      uint32
	TotalUsers       uint32
	ConversionRate   float64
}

// ProductSalesData represents sales data for a specific product in the business logic layer.
type ProductSalesData struct {
	ProductID     uint64
	ProductName   string
	SalesQuantity uint32
	SalesAmount   uint64
}
