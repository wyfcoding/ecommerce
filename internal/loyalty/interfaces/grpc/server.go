package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/loyalty/v1"              // 导入忠诚度模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/loyalty/application"   // 导入忠诚度模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity" // 导入忠诚度模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 LoyaltyService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedLoyaltyServiceServer                             // 嵌入生成的UnimplementedLoyaltyServiceServer，确保前向兼容性。
	app                                  *application.LoyaltyService // 依赖Loyalty应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Loyalty gRPC 服务端实例。
func NewServer(app *application.LoyaltyService) *Server {
	return &Server{app: app}
}

// GetMemberAccount 处理获取会员账户信息的gRPC请求。
// req: 包含用户ID的请求体。
// 返回会员账户响应和可能发生的gRPC错误。
func (s *Server) GetMemberAccount(ctx context.Context, req *pb.GetMemberAccountRequest) (*pb.GetMemberAccountResponse, error) {
	account, err := s.app.GetOrCreateAccount(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get or create member account: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.GetMemberAccountResponse{
		Account: convertAccountToProto(account),
	}, nil
}

// AddPoints 处理增加用户积分的gRPC请求。
// req: 包含用户ID、积分数量、交易类型和描述的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductPoints 处理扣减用户积分的gRPC请求。
// req: 包含用户ID、积分数量、交易类型和描述的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.TransactionType, req.Description, req.OrderId); err != nil {
		// 如果是积分不足错误，可以返回InvalidArgument状态码。
		if errors.Is(err, entity.ErrInsufficientPoints) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// AddSpent 处理增加用户消费金额的gRPC请求。
// req: 包含用户ID和消费金额的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddSpent(ctx context.Context, req *pb.AddSpentRequest) (*emptypb.Empty, error) {
	if err := s.app.AddSpent(ctx, req.UserId, req.Amount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add spent amount: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPointsTransactions 处理列出积分交易记录的gRPC请求。
// req: 包含用户ID和分页参数的请求体。
// 返回积分交易记录列表响应和可能发生的gRPC错误。
func (s *Server) ListPointsTransactions(ctx context.Context, req *pb.ListPointsTransactionsRequest) (*pb.ListPointsTransactionsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取积分交易记录列表。
	transactions, total, err := s.app.GetPointsTransactions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list points transactions: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbTransactions := make([]*pb.PointsTransaction, len(transactions))
	for i, tx := range transactions {
		pbTransactions[i] = convertTransactionToProto(tx)
	}

	return &pb.ListPointsTransactionsResponse{
		Transactions: pbTransactions,
		TotalCount:   uint64(total), // 总记录数。
	}, nil
}

// AddBenefit 处理添加会员权益的gRPC请求。
// req: 包含会员等级、权益名称、描述和费率的请求体。
// 返回创建成功的会员权益响应和可能发生的gRPC错误。
func (s *Server) AddBenefit(ctx context.Context, req *pb.AddBenefitRequest) (*pb.AddBenefitResponse, error) {
	// 调用应用服务层添加会员权益。
	benefit, err := s.app.AddBenefit(ctx, entity.MemberLevel(req.Level), req.Name, req.Description, req.DiscountRate, req.PointsRate)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add benefit: %v", err))
	}

	return &pb.AddBenefitResponse{
		Benefit: convertBenefitToProto(benefit),
	}, nil
}

// ListBenefits 处理列出会员权益的gRPC请求。
// req: 包含会员等级过滤的请求体。
// 返回会员权益列表响应和可能发生的gRPC错误。
func (s *Server) ListBenefits(ctx context.Context, req *pb.ListBenefitsRequest) (*pb.ListBenefitsResponse, error) {
	// 调用应用服务层获取会员权益列表。
	benefits, err := s.app.ListBenefits(ctx, entity.MemberLevel(req.Level))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list benefits: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbBenefits := make([]*pb.MemberBenefit, len(benefits))
	for i, b := range benefits {
		pbBenefits[i] = convertBenefitToProto(b)
	}

	return &pb.ListBenefitsResponse{
		Benefits: pbBenefits,
	}, nil
}

// DeleteBenefit 处理删除会员权益的gRPC请求。
// req: 包含权益ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteBenefit(ctx context.Context, req *pb.DeleteBenefitRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBenefit(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete benefit: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertAccountToProto 是一个辅助函数，将领域层的 MemberAccount 实体转换为 protobuf 的 MemberAccount 消息。
func convertAccountToProto(a *entity.MemberAccount) *pb.MemberAccount {
	if a == nil {
		return nil
	}
	return &pb.MemberAccount{
		Id:              uint64(a.ID),                 // 账户ID。
		UserId:          a.UserID,                     // 用户ID。
		Level:           string(a.Level),              // 会员等级。
		TotalPoints:     a.TotalPoints,                // 总积分。
		AvailablePoints: a.AvailablePoints,            // 可用积分。
		FrozenPoints:    a.FrozenPoints,               // 冻结积分。
		TotalSpent:      a.TotalSpent,                 // 总消费金额。
		CreatedAt:       timestamppb.New(a.CreatedAt), // 创建时间。
		UpdatedAt:       timestamppb.New(a.UpdatedAt), // 更新时间。
	}
}

// convertTransactionToProto 是一个辅助函数，将领域层的 PointsTransaction 实体转换为 protobuf 的 PointsTransaction 消息。
func convertTransactionToProto(t *entity.PointsTransaction) *pb.PointsTransaction {
	if t == nil {
		return nil
	}
	resp := &pb.PointsTransaction{
		Id:              uint64(t.ID),                 // 交易ID。
		UserId:          t.UserID,                     // 用户ID。
		TransactionType: t.TransactionType,            // 交易类型。
		Points:          t.Points,                     // 积分变动。
		Balance:         t.Balance,                    // 变动后余额。
		OrderId:         t.OrderID,                    // 关联订单ID。
		Description:     t.Description,                // 描述。
		CreatedAt:       timestamppb.New(t.CreatedAt), // 创建时间。
	}
	if t.ExpireAt != nil {
		resp.ExpireAt = timestamppb.New(*t.ExpireAt) // 过期时间。
	}
	return resp
}

// convertBenefitToProto 是一个辅助函数，将领域层的 MemberBenefit 实体转换为 protobuf 的 MemberBenefit 消息。
func convertBenefitToProto(b *entity.MemberBenefit) *pb.MemberBenefit {
	if b == nil {
		return nil
	}
	return &pb.MemberBenefit{
		Id:           uint64(b.ID),                 // 权益ID。
		Level:        string(b.Level),              // 会员等级。
		Name:         b.Name,                       // 名称。
		Description:  b.Description,                // 描述。
		DiscountRate: b.DiscountRate,               // 折扣率。
		PointsRate:   b.PointsRate,                 // 积分倍率。
		Enabled:      b.Enabled,                    // 是否启用。
		CreatedAt:    timestamppb.New(b.CreatedAt), // 创建时间。
		UpdatedAt:    timestamppb.New(b.UpdatedAt), // 更新时间。
	}
}
