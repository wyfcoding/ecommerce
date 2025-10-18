package repository

import (
	"context"
	"time"

	"ecommerce/internal/data_warehouse_query/model"
)

// DataWarehouseQueryRepo defines the interface for data warehouse query data access.
type DataWarehouseQueryRepo interface {
	ExecuteQuery(ctx context.Context, querySQL string, parameters map[string]string) (*model.QueryResult, error)
}