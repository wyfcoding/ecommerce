package service

import (
	"context"

	v1 "ecommerce/api/data_lake_ingestion/v1"
	"ecommerce/internal/data_lake_ingestion/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataLakeIngestionService is the gRPC service implementation for data lake ingestion.
type DataLakeIngestionService struct {
	v1.UnimplementedDataLakeIngestionServiceServer
	uc *biz.DataLakeIngestionUsecase
}

// NewDataLakeIngestionService creates a new DataLakeIngestionService.
func NewDataLakeIngestionService(uc *biz.DataLakeIngestionUsecase) *DataLakeIngestionService {
	return &DataLakeIngestionService{uc: uc}
}

// IngestData implements the IngestData RPC.
func (s *DataLakeIngestionService) IngestData(ctx context.Context, req *v1.IngestDataRequest) (*v1.IngestDataResponse, error) {
	if req.TableName == "" || req.RecordId == "" || req.Data == nil || req.EventTime == nil {
		return nil, status.Error(codes.InvalidArgument, "table_name, record_id, data, and event_time are required")
	}

	data := make(map[string]string)
	for k, v := range req.Data {
		data[k] = v
	}

	eventTime := req.EventTime.AsTime()

	err := s.uc.IngestData(ctx, req.TableName, req.RecordId, data, eventTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ingest data: %v", err)
	}

	return &v1.IngestDataResponse{Success: true, Message: "Data ingested successfully"}, nil
}
