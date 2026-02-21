package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/shopspring/decimal"

	pb "github.com/wyfcoding/ecommerce/go-api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 SettlementService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedSettlementServiceServer
	cmd   *application.SettlementCommandService
	query *application.SettlementQueryService
}

// NewServer 创建并返回一个新的 Settlement gRPC 服务端实例。
func NewServer(cmd *application.SettlementCommandService, query *application.SettlementQueryService) *Server {
	return &Server{cmd: cmd, query: query}
}

// CreateSettlement 处理创建结算单的gRPC请求。
func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateSettlement received", "merchant_id", req.MerchantId, "cycle", req.Cycle)

	settlement, err := s.cmd.CreateSettlement(ctx, req.MerchantId, req.Cycle, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		slog.Error("gRPC CreateSettlement failed", "merchant_id", req.MerchantId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create settlement: %v", err))
	}

	slog.Info("gRPC CreateSettlement successful", "settlement_id", settlement.ID, "duration", time.Since(start))
	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// AddOrderToSettlement 处理添加订单到结算单的gRPC请求。
func (s *Server) AddOrderToSettlement(ctx context.Context, req *pb.AddOrderToSettlementRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC AddOrderToSettlement received", "settlement_id", req.SettlementId, "order_id", req.OrderId)

	if err := s.cmd.AddOrderToSettlement(ctx, req.SettlementId, req.OrderId, req.OrderNo, req.Amount); err != nil {
		slog.Error("gRPC AddOrderToSettlement failed", "settlement_id", req.SettlementId, "order_id", req.OrderId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add order to settlement: %v", err))
	}

	slog.Info("gRPC AddOrderToSettlement successful", "settlement_id", req.SettlementId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// ProcessSettlement 处理结算单的gRPC请求。
func (s *Server) ProcessSettlement(ctx context.Context, req *pb.ProcessSettlementRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC ProcessSettlement received", "id", req.Id)

	if err := s.cmd.ProcessSettlement(ctx, req.Id); err != nil {
		slog.Error("gRPC ProcessSettlement failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process settlement: %v", err))
	}

	slog.Info("gRPC ProcessSettlement successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// CompleteSettlement 处理完成结算单的gRPC请求。
func (s *Server) CompleteSettlement(ctx context.Context, req *pb.CompleteSettlementRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC CompleteSettlement received", "id", req.Id)

	if err := s.cmd.CompleteSettlement(ctx, req.Id); err != nil {
		slog.Error("gRPC CompleteSettlement failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to complete settlement: %v", err))
	}

	slog.Info("gRPC CompleteSettlement successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

// ListSettlements 处理列出结算单的gRPC请求.
func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListSettlements received", "merchant_id", req.MerchantId)

	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *domain.SettlementStatus
	if req.Status != -1 {
		st := domain.SettlementStatus(req.Status)
		statusPtr = &st
	}

	settlements, total, err := s.query.ListSettlements(ctx, req.MerchantId, statusPtr, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListSettlements failed", "merchant_id", req.MerchantId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list settlements: %v", err))
	}

	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, s := range settlements {
		pbSettlements[i] = convertSettlementToProto(s)
	}

	slog.Debug("gRPC ListSettlements successful", "merchant_id", req.MerchantId, "count", len(pbSettlements), "duration", time.Since(start))
	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		TotalCount:  total,
	}, nil
}

// GetMerchantAccount 处理获取商户账户信息的gRPC请求。
func (s *Server) GetMerchantAccount(ctx context.Context, req *pb.GetMerchantAccountRequest) (*pb.GetMerchantAccountResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetMerchantAccount received", "merchant_id", req.MerchantId)

	account, err := s.query.GetMerchantAccount(ctx, req.MerchantId)
	if err != nil {
		slog.Error("gRPC GetMerchantAccount failed", "merchant_id", req.MerchantId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get merchant account: %v", err))
	}

	slog.Debug("gRPC GetMerchantAccount successful", "merchant_id", req.MerchantId, "duration", time.Since(start))
	return &pb.GetMerchantAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

func convertSettlementToProto(s *domain.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	var settledAt *timestamppb.Timestamp
	if s.PaidAt != nil {
		settledAt = timestamppb.New(*s.PaidAt)
	}

	return &pb.Settlement{
		Id:               uint64(s.ID),
		SettlementNo:     s.SettlementID,
		MerchantId:       s.MerchantID,
		Cycle:            string(s.Cycle),
		StartDate:        timestamppb.New(s.PeriodStart),
		EndDate:          timestamppb.New(s.PeriodEnd),
		OrderCount:       s.OrderCount,
		TotalAmount:      uint64(s.GrossAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		PlatformFee:      uint64(s.PlatformCommission.Mul(decimal.NewFromInt(100)).IntPart()),
		SettlementAmount: uint64(s.SettlementAmount.Mul(decimal.NewFromInt(100)).IntPart()),
		Status:           convertStatusToProto(s.Status),
		SettledAt:        settledAt,
		FailReason:       s.FailReason,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}
}

func convertStatusToProto(status domain.SettlementStatus) int32 {
	// 假设 pb.SettlementStatus 枚举值对应顺序，或简单返回 int 映射
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

func convertAccountToProto(a *domain.MerchantAccount) *pb.MerchantAccount {
	if a == nil {
		return nil
	}
	return &pb.MerchantAccount{
		Id:            uint64(a.ID),
		MerchantId:    a.MerchantID,
		Balance:       uint64(a.Balance.Mul(decimal.NewFromInt(100)).IntPart()),
		FrozenBalance: uint64(a.FrozenBalance.Mul(decimal.NewFromInt(100)).IntPart()),
		TotalIncome:   uint64(a.TotalIncome.Mul(decimal.NewFromInt(100)).IntPart()),
		TotalWithdraw: uint64(a.TotalWithdraw.Mul(decimal.NewFromInt(100)).IntPart()),
		FeeRate:       a.FeeRate.InexactFloat64(),
		CreatedAt:     timestamppb.New(a.CreatedAt),
		UpdatedAt:     timestamppb.New(a.UpdatedAt),
	}
}
