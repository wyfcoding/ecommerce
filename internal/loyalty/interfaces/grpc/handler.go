package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/loyalty/v1"
	"github.com/wyfcoding/ecommerce/internal/loyalty/application"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 Loyalty 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedLoyaltyServiceServer
	app *application.Loyalty
}

// NewServer 创建并返回一个新的 Loyalty gRPC 服务端实例。
func NewServer(app *application.Loyalty) *Server {
	return &Server{app: app}
}

// GetMemberAccount 处理获取会员账户信息的gRPC请求。
func (s *Server) GetMemberAccount(ctx context.Context, req *pb.GetMemberAccountRequest) (*pb.GetMemberAccountResponse, error) {
	account, err := s.app.GetOrCreateAccount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get or create member account: %v", err))
	}

	return &pb.GetMemberAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

// AddPoints 处理增加用户积分的gRPC请求。
func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductPoints 处理扣减用户积分的gRPC请求。
func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		if errors.Is(err, domain.ErrInsufficientPoints) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// AddSpent 处理增加用户消费金额的gRPC请求。
func (s *Server) AddSpent(ctx context.Context, req *pb.AddSpentRequest) (*emptypb.Empty, error) {
	if err := s.app.AddSpent(ctx, req.UserId, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add spent amount: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPointsTransactions 处理列出积分交易记录的gRPC请求。
func (s *Server) ListPointsTransactions(ctx context.Context, req *pb.ListPointsTransactionsRequest) (*pb.ListPointsTransactionsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	transactions, total, err := s.app.GetPointsTransactions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list points transactions: %v", err))
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

// AddBenefit 处理添加会员权益的gRPC请求。
func (s *Server) AddBenefit(ctx context.Context, req *pb.AddBenefitRequest) (*pb.AddBenefitResponse, error) {
	benefit, err := s.app.AddBenefit(ctx, domain.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add benefit: %v", err))
	}

	return &pb.AddBenefitResponse{
		Benefit: convertBenefitToProto(benefit),
	}, nil
}

// ListBenefits 处理列出会员权益的gRPC请求。
func (s *Server) ListBenefits(ctx context.Context, req *pb.ListBenefitsRequest) (*pb.ListBenefitsResponse, error) {
	benefits, err := s.app.ListBenefits(ctx, domain.MemberLevel(req.Level))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list benefits: %v", err))
	}

	pbBenefits := make([]*pb.MemberBenefit, len(benefits))
	for i, b := range benefits {
		pbBenefits[i] = convertBenefitToProto(b)
	}

	return &pb.ListBenefitsResponse{
		Benefits: pbBenefits,
	}, nil
}

// DeleteBenefit 处理删除会员权益的gRPC请求。
func (s *Server) DeleteBenefit(ctx context.Context, req *pb.DeleteBenefitRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBenefit(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete benefit: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertAccountToProto 是一个辅助函数，将领域层的 MemberAccount 实体转换为 protobuf 的 MemberAccount 消息。
func convertAccountToProto(a *domain.MemberAccount) *pb.MemberAccount {
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

// convertTransactionToProto 是一个辅助函数，将领域层的 PointsTransaction 实体转换为 protobuf 的 PointsTransaction 消息。
func convertTransactionToProto(t *domain.PointsTransaction) *pb.PointsTransaction {
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

// convertBenefitToProto 是一个辅助函数，将领域层的 MemberBenefit 实体转换为 protobuf 的 MemberBenefit 消息。
func convertBenefitToProto(b *domain.MemberBenefit) *pb.MemberBenefit {
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
