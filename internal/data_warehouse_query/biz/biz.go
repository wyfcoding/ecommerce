package biz

import (
	"context"
	"time"
)

// QueryResult represents a result from a data warehouse query in the business logic layer.
type QueryResult struct {
	ColumnNames []string
	Rows        []map[string]string
	Message     string
	QueryTime   time.Time
}

// DataWarehouseQueryRepo defines the interface for data warehouse query data access.
type DataWarehouseQueryRepo interface {
	ExecuteQuery(ctx context.Context, querySQL string, parameters map[string]string) (*QueryResult, error)
}

// DataWarehouseQueryUsecase is the business logic for data warehouse queries.
type DataWarehouseQueryUsecase struct {
	repo DataWarehouseQueryRepo
}

// NewDataWarehouseQueryUsecase creates a new DataWarehouseQueryUsecase.
func NewDataWarehouseQueryUsecase(repo DataWarehouseQueryRepo) *DataWarehouseQueryUsecase {
	return &DataWarehouseQueryUsecase{repo: repo}
}

// ExecuteQuery executes a query against the data warehouse.
func (uc *DataWarehouseQueryUsecase) ExecuteQuery(ctx context.Context, querySQL string, parameters map[string]string) (*QueryResult, error) {
	// Add any business logic here, e.g., query validation, permission checks
	return uc.repo.ExecuteQuery(ctx, querySQL, parameters)
}
