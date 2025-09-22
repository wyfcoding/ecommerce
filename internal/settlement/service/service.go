package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	v1 "ecommerce/api/settlement/v1"
	"ecommerce/internal/settlement/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SettlementService is the gRPC service implementation for settlement.
type SettlementService struct {
	v1.UnimplementedSettlementServiceServer
	uc *biz.SettlementUsecase
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(uc *biz.SettlementUsecase) *SettlementService {
	return &SettlementService{uc: uc}
}

// bizSettlementRecordToProto converts biz.SettlementRecord to v1.SettlementRecord.
func bizSettlementRecordToProto(record *biz.SettlementRecord) *v1.SettlementRecord {
	if record == nil {
		return nil
	}
	protoRecord := &v1.SettlementRecord{
		RecordId:         strconv.FormatUint(uint64(record.ID), 10), // Convert uint to string for proto
		OrderId:          record.OrderID,
		MerchantId:       record.MerchantID,
		TotalAmount:      record.TotalAmount,
		PlatformFee:      record.PlatformFee,
		SettlementAmount: record.SettlementAmount,
		Status:           record.Status,
		CreatedAt:        timestamppb.New(record.CreatedAt),
	}
	if record.SettledAt != nil {
		protoRecord.SettledAt = timestamppb.New(*record.SettledAt)
	}
	return protoRecord
}

// ProcessOrderSettlement implements the ProcessOrderSettlement RPC.
func (s *SettlementService) ProcessOrderSettlement(ctx context.Context, req *v1.ProcessOrderSettlementRequest) (*v1.ProcessOrderSettlementResponse, error) {
	if req.OrderId == 0 || req.MerchantId == 0 || req.TotalAmount == 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id, merchant_id, and total_amount are required")
	}

	record, err := s.uc.ProcessOrderSettlement(ctx, req.OrderId, req.MerchantId, req.TotalAmount)
	if err != nil {
		if errors.Is(err, biz.ErrOrderAlreadySettled) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to process order settlement: %v", err)
	}

	return &v1.ProcessOrderSettlementResponse{
		RecordId: record.RecordID,
		Status:   record.Status,
		Message:  "Settlement processed successfully",
	}, nil
}

// GetSettlementRecord implements the GetSettlementRecord RPC.
func (s *SettlementService) GetSettlementRecord(ctx context.Context, req *v1.GetSettlementRecordRequest) (*v1.SettlementRecord, error) {
	if req.RecordId == 0 { // Assuming record_id is uint64 in proto
		return nil, status.Error(codes.InvalidArgument, "record_id is required")
	}

	record, err := s.uc.GetSettlementRecord(ctx, strconv.FormatUint(req.RecordId, 10)) // Convert uint64 to string for biz layer
	if err != nil {
		if errors.Is(err, biz.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to get settlement record: %v", err)
	}

	return bizSettlementRecordToProto(record), nil
}

// ListSettlementRecords implements the ListSettlementRecords RPC.
func (s *SettlementService) ListSettlementRecords(ctx context.Context, req *v1.ListSettlementRecordsRequest) (*v1.ListSettlementRecordsResponse, error) {
	records, total, err := s.uc.ListSettlementRecords(ctx, req.MerchantId, req.Status, req.PageSize, req.PageNum)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list settlement records: %v", err)
	}

	protoRecords := make([]*v1.SettlementRecord, len(records))
	for i, rec := range records {
		protoRecords[i] = bizSettlementRecordToProto(rec)
	}

	return &v1.ListSettlementRecordsResponse{
		Records:    protoRecords,
		TotalCount: total,
	}, nil
}
