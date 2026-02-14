package interfaces

import (
	"context"
	"strconv"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/merchantsettlement/v1"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/application"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MerchantSettlementGRPCServer 商家结算 gRPC 服务实现。
type MerchantSettlementGRPCServer struct {
	pb.UnimplementedMerchantSettlementServiceServer
	app *application.MerchantSettlementService
}

// NewMerchantSettlementGRPCServer 创建商家结算 gRPC 服务。
func NewMerchantSettlementGRPCServer(app *application.MerchantSettlementService) *MerchantSettlementGRPCServer {
	return &MerchantSettlementGRPCServer{app: app}
}

// GenerateSettlement 生成结算单。
func (s *MerchantSettlementGRPCServer) GenerateSettlement(ctx context.Context, req *pb.GenerateSettlementRequest) (*pb.GenerateSettlementResponse, error) {
	merchantID, err := strconv.ParseUint(req.MerchantId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid merchant_id")
	}

	periodStart := ""
	periodEnd := ""
	if req.StartDate != nil {
		periodStart = req.StartDate.AsTime().Format("2006-01-02")
	}
	if req.EndDate != nil {
		periodEnd = req.EndDate.AsTime().Format("2006-01-02")
	}

	dto, err := s.app.CreateSettlement(ctx, merchantID, domain.CycleMonthly, periodStart, periodEnd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate settlement: %v", err)
	}

	return &pb.GenerateSettlementResponse{
		SettlementId: dto.SettlementID,
		Amount:       dto.SettlementAmount.String(),
	}, nil
}

// GetSettlement 查询结算单。
func (s *MerchantSettlementGRPCServer) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.GetSettlementResponse, error) {
	dto, err := s.app.GetSettlement(ctx, req.SettlementId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get settlement: %v", err)
	}

	return &pb.GetSettlementResponse{Settlement: toProtoSettlement(dto)}, nil
}

// ListSettlements 查询结算单列表。
func (s *MerchantSettlementGRPCServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	var merchantID uint64
	var err error
	if req.MerchantId != "" {
		merchantID, err = strconv.ParseUint(req.MerchantId, 10, 64)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid merchant_id")
		}
	}

	result, err := s.app.ListSettlements(ctx, merchantID, domain.SettlementStatus(req.Status), 1, 100)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list settlements: %v", err)
	}

	resp := &pb.ListSettlementsResponse{Settlements: make([]*pb.SettlementDetail, 0, len(result.Settlements))}
	for _, settlement := range result.Settlements {
		resp.Settlements = append(resp.Settlements, toProtoSettlement(settlement))
	}
	return resp, nil
}

// MarkAsPaid 标记已支付。
func (s *MerchantSettlementGRPCServer) MarkAsPaid(ctx context.Context, req *pb.MarkAsPaidRequest) (*pb.MarkAsPaidResponse, error) {
	dto, err := s.app.GetSettlement(ctx, req.SettlementId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get settlement: %v", err)
	}
	if dto.BankAccountID == 0 {
		return nil, status.Error(codes.FailedPrecondition, "bank account is not configured")
	}
	if err := s.app.PaySettlement(ctx, req.SettlementId, dto.BankAccountID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark settlement as paid: %v", err)
	}
	return &pb.MarkAsPaidResponse{Success: true}, nil
}

func toProtoSettlement(dto *application.SettlementDTO) *pb.SettlementDetail {
	if dto == nil {
		return nil
	}
	result := &pb.SettlementDetail{
		SettlementId: dto.SettlementID,
		MerchantId:   strconv.FormatUint(dto.MerchantID, 10),
		Amount:       dto.SettlementAmount.String(),
		Status:       dto.Status,
	}
	if !dto.PeriodStart.IsZero() {
		result.PeriodStart = timestamppb.New(dto.PeriodStart)
	}
	if !dto.PeriodEnd.IsZero() {
		result.PeriodEnd = timestamppb.New(dto.PeriodEnd)
	}
	if !dto.CreatedAt.IsZero() {
		result.CreatedAt = timestamppb.New(dto.CreatedAt)
	} else {
		result.CreatedAt = timestamppb.New(time.Now())
	}
	return result
}

var _ pb.MerchantSettlementServiceServer = (*MerchantSettlementGRPCServer)(nil)
