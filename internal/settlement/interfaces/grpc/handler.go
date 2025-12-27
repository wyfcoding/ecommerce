package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 SettlementService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedSettlementServer
	app *application.SettlementService
}

// NewServer 创建并返回一个新的 Settlement gRPC 服务端实例。
func NewServer(app *application.SettlementService) *Server {
	return &Server{app: app}
}

// CreateSettlement 处理创建结算单的gRPC请求。
func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	settlement, err := s.app.CreateSettlement(ctx, req.MerchantId, req.Cycle, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create settlement: %v", err))
	}

	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// AddOrderToSettlement 处理添加订单到结算单的gRPC请求。
func (s *Server) AddOrderToSettlement(ctx context.Context, req *pb.AddOrderToSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.AddOrderToSettlement(ctx, req.SettlementId, req.OrderId, req.OrderNo, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add order to settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ProcessSettlement 处理结算单的gRPC请求。
func (s *Server) ProcessSettlement(ctx context.Context, req *pb.ProcessSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.ProcessSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// CompleteSettlement 处理完成结算单的gRPC请求。
func (s *Server) CompleteSettlement(ctx context.Context, req *pb.CompleteSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.CompleteSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to complete settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListSettlements 处理列出结算单的gRPC请求.
func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	var statusPtr *int
	if req.Status != -1 {
		st := int(req.Status)
		statusPtr = &st
	}

	settlements, total, err := s.app.ListSettlements(ctx, req.MerchantId, statusPtr, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list settlements: %v", err))
	}

	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, s := range settlements {
		pbSettlements[i] = convertSettlementToProto(s)
	}

	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		TotalCount:  total,
	}, nil
}

// GetMerchantAccount 处理获取商户账户信息的gRPC请求。
func (s *Server) GetMerchantAccount(ctx context.Context, req *pb.GetMerchantAccountRequest) (*pb.GetMerchantAccountResponse, error) {
	account, err := s.app.GetMerchantAccount(ctx, req.MerchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get merchant account: %v", err))
	}

	return &pb.GetMerchantAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

func convertSettlementToProto(s *domain.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	var settledAt *timestamppb.Timestamp
	if s.SettledAt != nil {
		settledAt = timestamppb.New(*s.SettledAt)
	}

	return &pb.Settlement{
		Id:               uint64(s.ID),
		SettlementNo:     s.SettlementNo,
		MerchantId:       s.MerchantID,
		Cycle:            string(s.Cycle),
		StartDate:        timestamppb.New(s.StartDate),
		EndDate:          timestamppb.New(s.EndDate),
		OrderCount:       s.OrderCount,
		TotalAmount:      s.TotalAmount,
		PlatformFee:      s.PlatformFee,
		SettlementAmount: s.SettlementAmount,
		Status:           int32(s.Status),
		SettledAt:        settledAt,
		FailReason:       s.FailReason,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}
}

func convertAccountToProto(a *domain.MerchantAccount) *pb.MerchantAccount {
	if a == nil {
		return nil
	}
	return &pb.MerchantAccount{
		Id:            uint64(a.ID),
		MerchantId:    a.MerchantID,
		Balance:       a.Balance,
		FrozenBalance: a.FrozenBalance,
		TotalIncome:   a.TotalIncome,
		TotalWithdraw: a.TotalWithdraw,
		FeeRate:       a.FeeRate,
		CreatedAt:     timestamppb.New(a.CreatedAt),
		UpdatedAt:     timestamppb.New(a.UpdatedAt),
	}
}
