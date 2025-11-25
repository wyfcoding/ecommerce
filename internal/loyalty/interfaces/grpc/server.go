package grpc

import (
	"context"
	pb "ecommerce/api/loyalty/v1"
	"ecommerce/internal/loyalty/application"
	"ecommerce/internal/loyalty/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedLoyaltyServiceServer
	app *application.LoyaltyService
}

func NewServer(app *application.LoyaltyService) *Server {
	return &Server{app: app}
}

func (s *Server) GetMemberAccount(ctx context.Context, req *pb.GetMemberAccountRequest) (*pb.GetMemberAccountResponse, error) {
	account, err := s.app.GetOrCreateAccount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetMemberAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) AddSpent(ctx context.Context, req *pb.AddSpentRequest) (*emptypb.Empty, error) {
	if err := s.app.AddSpent(ctx, req.UserId, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListPointsTransactions(ctx context.Context, req *pb.ListPointsTransactionsRequest) (*pb.ListPointsTransactionsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	transactions, total, err := s.app.GetPointsTransactions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTransactions := make([]*pb.PointsTransaction, len(transactions))
	for i, tx := range transactions {
		pbTransactions[i] = convertTransactionToProto(tx)
	}

	return &pb.ListPointsTransactionsResponse{
		Transactions: pbTransactions,
		TotalCount:   uint64(total),
	}, nil
}

func (s *Server) AddBenefit(ctx context.Context, req *pb.AddBenefitRequest) (*pb.AddBenefitResponse, error) {
	benefit, err := s.app.AddBenefit(ctx, entity.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddBenefitResponse{
		Benefit: convertBenefitToProto(benefit),
	}, nil
}

func (s *Server) ListBenefits(ctx context.Context, req *pb.ListBenefitsRequest) (*pb.ListBenefitsResponse, error) {
	benefits, err := s.app.ListBenefits(ctx, entity.MemberLevel(req.Level))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbBenefits := make([]*pb.MemberBenefit, len(benefits))
	for i, b := range benefits {
		pbBenefits[i] = convertBenefitToProto(b)
	}

	return &pb.ListBenefitsResponse{
		Benefits: pbBenefits,
	}, nil
}

func (s *Server) DeleteBenefit(ctx context.Context, req *pb.DeleteBenefitRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBenefit(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func convertAccountToProto(a *entity.MemberAccount) *pb.MemberAccount {
	if a == nil {
		return nil
	}
	return &pb.MemberAccount{
		Id:              uint64(a.ID),
		UserId:          a.UserID,
		Level:           string(a.Level),
		TotalPoints:     a.TotalPoints,
		AvailablePoints: a.AvailablePoints,
		FrozenPoints:    a.FrozenPoints,
		TotalSpent:      a.TotalSpent,
		CreatedAt:       timestamppb.New(a.CreatedAt),
		UpdatedAt:       timestamppb.New(a.UpdatedAt),
	}
}

func convertTransactionToProto(t *entity.PointsTransaction) *pb.PointsTransaction {
	if t == nil {
		return nil
	}
	resp := &pb.PointsTransaction{
		Id:              uint64(t.ID),
		UserId:          t.UserID,
		TransactionType: t.TransactionType,
		Points:          t.Points,
		Balance:         t.Balance,
		OrderId:         t.OrderID,
		Description:     t.Description,
		CreatedAt:       timestamppb.New(t.CreatedAt),
	}
	if t.ExpireAt != nil {
		resp.ExpireAt = timestamppb.New(*t.ExpireAt)
	}
	return resp
}

func convertBenefitToProto(b *entity.MemberBenefit) *pb.MemberBenefit {
	if b == nil {
		return nil
	}
	return &pb.MemberBenefit{
		Id:           uint64(b.ID),
		Level:        string(b.Level),
		Name:         b.Name,
		Description:  b.Description,
		DiscountRate: b.DiscountRate,
		PointsRate:   b.PointsRate,
		Enabled:      b.Enabled,
		CreatedAt:    timestamppb.New(b.CreatedAt),
		UpdatedAt:    timestamppb.New(b.UpdatedAt),
	}
}
