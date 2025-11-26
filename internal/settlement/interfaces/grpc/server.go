package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedSettlementServiceServer
	app *application.SettlementService
}

func NewServer(app *application.SettlementService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	settlement, err := s.app.CreateSettlement(ctx, req.MerchantId, req.Cycle, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

func (s *Server) AddOrderToSettlement(ctx context.Context, req *pb.AddOrderToSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.AddOrderToSettlement(ctx, req.SettlementId, req.OrderId, req.OrderNo, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ProcessSettlement(ctx context.Context, req *pb.ProcessSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.ProcessSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) CompleteSettlement(ctx context.Context, req *pb.CompleteSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.CompleteSettlement(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
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
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) GetMerchantAccount(ctx context.Context, req *pb.GetMerchantAccountRequest) (*pb.GetMerchantAccountResponse, error) {
	account, err := s.app.GetMerchantAccount(ctx, req.MerchantId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetMerchantAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

func convertSettlementToProto(s *entity.Settlement) *pb.Settlement {
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

func convertAccountToProto(a *entity.MerchantAccount) *pb.MerchantAccount {
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
