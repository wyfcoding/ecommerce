package service

import (
	"context"
	"time"

	v1 "ecommerce/api/data_ingestion/v1"
	"ecommerce/internal/data_ingestion/biz"

	"github.com/google/uuid" // Added for uuid.New()
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DataIngestionService is the gRPC service implementation for data ingestion.
type DataIngestionService struct {
	v1.UnimplementedDataIngestionServiceServer
	uc *biz.DataIngestionUsecase
}

// NewDataIngestionService creates a new DataIngestionService.
func NewDataIngestionService(uc *biz.DataIngestionUsecase) *DataIngestionService {
	return &DataIngestionService{uc: uc}
}

// IngestEvent implements the IngestEvent RPC.
func (s *DataIngestionService) IngestEvent(ctx context.Context, req *v1.IngestEventRequest) (*v1.IngestEventResponse, error) {
	if req.EventType == "" || req.UserId == "" || req.EntityId == "" || req.Timestamp == "" {
		return nil, status.Error(codes.InvalidArgument, "event_type, user_id, entity_id, and timestamp are required")
	}

	eventTimestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid timestamp format: %v", err)
	}

	properties := make(map[string]string)
	for k, v := range req.Properties {
		properties[k] = v
	}

	err = s.uc.IngestEvent(ctx, req.EventType, req.UserId, req.EntityId, properties, eventTimestamp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to ingest event: %v", err)
	}

	// In a real system, you might generate a unique event ID here.
	return &v1.IngestEventResponse{EventId: uuid.New().String()}, nil
}
