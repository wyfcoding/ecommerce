package service

import (
	"context"

	v1 "ecommerce/api/cdc/v1"
	"ecommerce/internal/cdc/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CdcService is the gRPC service implementation for CDC.
type CdcService struct {
	v1.UnimplementedCdcServiceServer
	uc *biz.CdcUsecase
}

// NewCdcService creates a new CdcService.
func NewCdcService(uc *biz.CdcUsecase) *CdcService {
	return &CdcService{uc: uc}
}

// CaptureChangeEvent implements the CaptureChangeEvent RPC.
func (s *CdcService) CaptureChangeEvent(ctx context.Context, req *v1.CaptureChangeEventRequest) (*v1.CaptureChangeEventResponse, error) {
	if req.TableName == "" || req.OperationType == "" || req.PrimaryKeyValue == "" {
		return nil, status.Error(codes.InvalidArgument, "table_name, operation_type, and primary_key_value are required")
	}

	event, err := s.uc.CaptureChangeEvent(ctx, req.TableName, req.OperationType, req.PrimaryKeyValue, req.OldData, req.NewData)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to capture change event: %v", err)
	}

	return &v1.CaptureChangeEventResponse{EventId: event.EventID}, nil
}
