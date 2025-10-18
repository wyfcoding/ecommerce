package service

import (
	"context"
	"time"

	"ecommerce/internal/data_warehouse_query/model"
	"ecommerce/internal/data_warehouse_query/repository"
)

// DataWarehouseQueryService is the business logic for data warehouse queries.
type DataWarehouseQueryService struct {
	repo repository.DataWarehouseQueryRepo
}

// NewDataWarehouseQueryService creates a new DataWarehouseQueryService.
func NewDataWarehouseQueryService(repo repository.DataWarehouseQueryRepo) *DataWarehouseQueryService {
	return &DataWarehouseQueryService{repo: repo}
}

// ExecuteQuery executes a query against the data warehouse.
func (s *DataWarehouseQueryService) ExecuteQuery(ctx context.Context, querySQL string, parameters map[string]string) (*model.QueryResult, error) {
	// Add any business logic here, e.g., query validation, permission checks
	return s.repo.ExecuteQuery(ctx, querySQL, parameters)
}
