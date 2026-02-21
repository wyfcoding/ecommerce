package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
	pb "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SettlementGRPCServer 结算服务 gRPC 实现。
type SettlementGRPCServer struct {
	pb.UnimplementedSettlementServiceServer
	cmd *application.SettlementCommandService
}

// NewSettlementGRPCServer 创建结算 gRPC 服务。
func NewSettlementGRPCServer(cmd *application.SettlementCommandService) *SettlementGRPCServer {
	return &SettlementGRPCServer{cmd: cmd}
}

// CreateSettlement 生成结算单。
func (s *SettlementGRPCServer) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()

	dto, err := s.cmd.CreateSettlement(ctx, req.MerchantId, req.Cycle, startDate, endDate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create settlement: %v", err)
	}

	return &pb.CreateSettlementResponse{
		Settlement: toProtoSettlement(dto),
	}, nil
}

// AddOrderToSettlement 将已完成订单归入特定结算单。
func (s *SettlementGRPCServer) AddOrderToSettlement(ctx context.Context, req *pb.AddOrderToSettlementRequest) (*emptypb.Empty, error) {
	if err := s.cmd.AddOrderToSettlement(ctx, req.SettlementId, req.OrderId, req.OrderNo, req.Amount); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add order to settlement: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ProcessSettlement 执行金额核算。
func (s *SettlementGRPCServer) ProcessSettlement(ctx context.Context, req *pb.ProcessSettlementRequest) (*emptypb.Empty, error) {
	if err := s.cmd.ProcessSettlement(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to process settlement: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CompleteSettlement 确认资金划转完成。
func (s *SettlementGRPCServer) CompleteSettlement(ctx context.Context, req *pb.CompleteSettlementRequest) (*emptypb.Empty, error) {
	if err := s.cmd.CompleteSettlement(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to complete settlement: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListSettlements 查询历史结算记录。
func (s *SettlementGRPCServer) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	// TODO: 调用 QueryService 列表查询 (目前 cmd 还不支持)
	return nil, status.Error(codes.Unimplemented, "ListSettlements is not implemented in command service")
}

// RecordPaymentSuccess 记录订单实时入账。
func (s *SettlementGRPCServer) RecordPaymentSuccess(ctx context.Context, req *pb.RecordPaymentSuccessRequest) (*emptypb.Empty, error) {
	if err := s.cmd.RecordPaymentSuccess(ctx, req.OrderId, req.OrderNo, req.MerchantId, req.Amount, req.ChannelCost); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record payment: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// GetMerchantAccount 获取商家实时余额及费率。
func (s *SettlementGRPCServer) GetMerchantAccount(ctx context.Context, req *pb.GetMerchantAccountRequest) (*pb.GetMerchantAccountResponse, error) {
	// TODO: 调用 QueryService 或 Repo 获取细节
	return nil, status.Error(codes.Unimplemented, "GetMerchantAccount is not implemented in command service")
}

func toProtoSettlement(s *domain.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	res := &pb.Settlement{
		Id:               s.ID,
		SettlementNo:     s.SettlementNo,
		MerchantId:       s.MerchantID,
		Cycle:            string(s.Cycle),
		StartDate:        timestamppb.New(s.StartDate),
		EndDate:          timestamppb.New(s.EndDate),
		OrderCount:       s.OrderCount,
		TotalAmount:      uint64(s.GrossAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		PlatformFee:      uint64(s.PlatformCommission.Mul(decimal.NewFromInt(100)).IntPart()),
		SettlementAmount: uint64(s.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		Status:           int32(toProtoStatus(s.Status)),
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}
	if s.SettledAt != nil {
		res.SettledAt = timestamppb.New(*s.SettledAt)
	}
	return res
}

func toProtoStatus(status domain.SettlementStatus) int32 {
	switch status {
	case domain.StatusPending:
		return 0
	case domain.StatusCalculating:
		return 1
	case domain.StatusPendingApproval:
		return 2
	case domain.StatusApproved:
		return 3
	case domain.StatusPaying:
		return 4
	case domain.StatusPaid:
		return 5
	case domain.StatusFailed:
		return 6
	case domain.StatusCancelled:
		return 7
	default:
		return 0
	}
}

var _ pb.SettlementServiceServer = (*SettlementGRPCServer)(nil)
