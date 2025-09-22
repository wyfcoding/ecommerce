package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	v1 "ecommerce/api/data_warehouse_query/v1"
	"ecommerce/internal/data_warehouse_query/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataWarehouseQueryService is the gRPC service implementation for data warehouse queries.
type DataWarehouseQueryService struct {
	v1.UnimplementedDataWarehouseQueryServiceServer
	uc *biz.DataWarehouseQueryUsecase
}

// NewDataWarehouseQueryService creates a new DataWarehouseQueryService.
func NewDataWarehouseQueryService(uc *biz.DataWarehouseQueryUsecase) *DataWarehouseQueryService {
	return &DataWarehouseQueryService{uc: uc}
}

// ExecuteQuery implements the ExecuteQuery RPC.
func (s *DataWarehouseQueryService) ExecuteQuery(ctx context.Context, req *v1.ExecuteQueryRequest) (*v1.ExecuteQueryResponse, error) {
	if req.QuerySql == "" {
		return nil, status.Error(codes.InvalidArgument, "query_sql is required")
	}

	parameters := make(map[string]string)
	for k, v := range req.Parameters {
		parameters[k] = v
	}

	result, err := s.uc.ExecuteQuery(ctx, req.QuerySql, parameters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute query: %v", err)
	}

	// Convert rows to JSON strings for proto response
	protoRows := make([]string, len(result.Rows))
	for i, row := range result.Rows {
		rowBytes, err := json.Marshal(row)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal row to JSON: %v", err)
		}
		protoRows[i] = string(rowBytes)
	}

	return &v1.ExecuteQueryResponse{
		ColumnNames: result.ColumnNames,
		Rows:        protoRows,
		Message:     result.Message,
	}, nil
}
