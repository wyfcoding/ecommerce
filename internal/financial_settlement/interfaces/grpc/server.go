package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/financial_settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/application"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedFinancialSettlementServiceServer
	app *application.FinancialSettlementService
}

func NewServer(app *application.FinancialSettlementService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	settlement, err := s.app.CreateSettlement(ctx, req.SellerId, req.Period, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

func (s *Server) ApproveSettlement(ctx context.Context, req *pb.ApproveSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.ApproveSettlement(ctx, req.Id, req.ApprovedBy); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RejectSettlement(ctx context.Context, req *pb.RejectSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.RejectSettlement(ctx, req.Id, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.GetSettlementResponse, error) {
	settlement, err := s.app.GetSettlement(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	settlements, total, err := s.app.ListSettlements(ctx, req.SellerId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbSettlements := make([]*pb.Settlement, len(settlements))
	for i, st := range settlements {
		pbSettlements[i] = convertSettlementToProto(st)
	}

	return &pb.ListSettlementsResponse{
		Settlements: pbSettlements,
		TotalCount:  uint64(total),
	}, nil
}

func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, err := s.app.ProcessPayment(ctx, req.SettlementId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProcessPaymentResponse{
		Payment: convertPaymentToProto(payment),
	}, nil
}

func convertSettlementToProto(s *entity.Settlement) *pb.Settlement {
	if s == nil {
		return nil
	}
	resp := &pb.Settlement{
		Id:               uint64(s.ID),
		SellerId:         s.SellerID,
		Period:           s.Period,
		StartDate:        timestamppb.New(s.StartDate),
		EndDate:          timestamppb.New(s.EndDate),
		TotalSalesAmount: s.TotalSalesAmount,
		CommissionAmount: s.CommissionAmount,
		RebateAmount:     s.RebateAmount,
		OtherFees:        s.OtherFees,
		FinalAmount:      s.FinalAmount,
		Status:           string(s.Status),
		ApprovedBy:       s.ApprovedBy,
		RejectionReason:  s.RejectionReason,
		CreatedAt:        timestamppb.New(s.CreatedAt),
	}
	if s.ApprovedAt != nil {
		resp.ApprovedAt = timestamppb.New(*s.ApprovedAt)
	}
	return resp
}

func convertPaymentToProto(p *entity.SettlementPayment) *pb.SettlementPayment {
	if p == nil {
		return nil
	}
	resp := &pb.SettlementPayment{
		Id:            uint64(p.ID),
		SettlementId:  p.SettlementID,
		SellerId:      p.SellerID,
		Amount:        p.Amount,
		Status:        string(p.Status),
		TransactionId: p.TransactionID,
		CreatedAt:     timestamppb.New(p.CreatedAt),
	}
	if p.CompletedAt != nil {
		resp.CompletedAt = timestamppb.New(*p.CompletedAt)
	}
	return resp
}
