package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/financial_settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/application"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedFinancialSettlementServiceServer
	app *application.SettlementService
}

// NewServer 创建并返回一个新的 FinancialSettlement gRPC 服务端实例。
func NewServer(app *application.SettlementService) *Server {
	return &Server{app: app}
}

// CreateSettlement 处理创建结算单的gRPC请求。
func (s *Server) CreateSettlement(ctx context.Context, req *pb.CreateSettlementRequest) (*pb.CreateSettlementResponse, error) {
	settlement, err := s.app.CreateSettlement(ctx, req.SellerId, req.Period, req.StartDate.AsTime(), req.EndDate.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create settlement: %v", err))
	}

	return &pb.CreateSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// ApproveSettlement 处理批准结算单的gRPC请求。
func (s *Server) ApproveSettlement(ctx context.Context, req *pb.ApproveSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.ApproveSettlement(ctx, req.Id, req.ApprovedBy); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to approve settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RejectSettlement 处理拒绝结算单的gRPC请求。
func (s *Server) RejectSettlement(ctx context.Context, req *pb.RejectSettlementRequest) (*emptypb.Empty, error) {
	if err := s.app.RejectSettlement(ctx, req.Id, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to reject settlement: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetSettlement 处理获取结算单详情的gRPC请求。
func (s *Server) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.GetSettlementResponse, error) {
	settlement, err := s.app.GetSettlement(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("settlement not found: %v", err))
	}
	return &pb.GetSettlementResponse{
		Settlement: convertSettlementToProto(settlement),
	}, nil
}

// ListSettlements 处理列出结算单的gRPC请求。
func (s *Server) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	settlements, total, err := s.app.ListSettlements(ctx, req.SellerId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list settlements: %v", err))
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

// ProcessPayment 处理结算单支付的gRPC请求。
func (s *Server) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, err := s.app.ProcessPayment(ctx, req.SettlementId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to process payment: %v", err))
	}

	return &pb.ProcessPaymentResponse{
		Payment: convertPaymentToProto(payment),
	}, nil
}

func convertSettlementToProto(s *domain.Settlement) *pb.Settlement {
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

func convertPaymentToProto(p *domain.SettlementPayment) *pb.SettlementPayment {
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
