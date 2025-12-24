package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/goapi/usertier/v1"           // 导入用户等级模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/usertier/application"   // 导入用户等级模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/usertier/domain/entity" // 导入用户等级模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 UserTierService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedUserTierServiceServer                              // 嵌入生成的UnimplementedUserTierServiceServer，确保前向兼容性。
	app                                   *application.UserTierService // 依赖UserTier应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 UserTier gRPC 服务端实例。
func NewServer(app *application.UserTierService) *Server {
	return &Server{app: app}
}

// GetUserTier 处理获取用户等级的gRPC请求。
// req: 包含用户ID的请求体。
// 返回用户等级响应和可能发生的gRPC错误。
func (s *Server) GetUserTier(ctx context.Context, req *pb.GetUserTierRequest) (*pb.GetUserTierResponse, error) {
	tier, err := s.app.GetUserTier(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user tier: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.GetUserTierResponse{
		Tier: convertTierToProto(tier),
	}, nil
}

// AddScore 处理增加用户成长值的gRPC请求。
// req: 包含用户ID和成长值分数的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddScore(ctx context.Context, req *pb.AddScoreRequest) (*emptypb.Empty, error) {
	if err := s.app.AddScore(ctx, req.UserId, req.Score); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add score: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetPoints 处理获取用户积分的gRPC请求。
// req: 包含用户ID的请求体。
// 返回用户积分响应和可能发生的gRPC错误。
func (s *Server) GetPoints(ctx context.Context, req *pb.GetPointsRequest) (*pb.GetPointsResponse, error) {
	points, err := s.app.GetPoints(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get points: %v", err))
	}

	return &pb.GetPointsResponse{
		Points: points, // 返回积分数量。
	}, nil
}

// AddPoints 处理增加用户积分的gRPC请求。
// req: 包含用户ID、积分数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductPoints 处理扣除用户积分的gRPC请求。
// req: 包含用户ID、积分数量和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPointsLogs 处理列出用户积分日志的gRPC请求。
// req: 包含用户ID和分页参数的请求体。
// 返回积分日志列表响应和可能发生的gRPC错误。
func (s *Server) ListPointsLogs(ctx context.Context, req *pb.ListPointsLogsRequest) (*pb.ListPointsLogsResponse, error) {
	// 获取分页参数。
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取积分日志列表。
	logs, total, err := s.app.ListPointsLogs(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list points logs: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbLogs := make([]*pb.PointsLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertPointsLogToProto(l)
	}

	return &pb.ListPointsLogsResponse{
		Logs:       pbLogs,
		TotalCount: total, // 总记录数。
	}, nil
}

// Exchange 处理兑换商品的gRPC请求。
// req: 包含用户ID和兑换商品ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) Exchange(ctx context.Context, req *pb.ExchangeRequest) (*emptypb.Empty, error) {
	if err := s.app.Exchange(ctx, req.UserId, req.ExchangeId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to exchange item: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListExchanges 处理列出可兑换商品列表的gRPC请求。
// req: 包含分页参数的请求体。
// 返回可兑换商品列表响应和可能发生的gRPC错误。
func (s *Server) ListExchanges(ctx context.Context, req *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
	// 获取分页参数。
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取可兑换商品列表。
	items, total, err := s.app.ListExchanges(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list exchanges: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbItems := make([]*pb.ExchangeItem, len(items))
	for i, item := range items {
		pbItems[i] = convertExchangeItemToProto(item)
	}

	return &pb.ListExchangesResponse{
		Items:      pbItems,
		TotalCount: total, // 总记录数。
	}, nil
}

// convertTierToProto 是一个辅助函数，将领域层的 UserTier 实体转换为 protobuf 的 UserTier 消息。
func convertTierToProto(t *entity.UserTier) *pb.UserTier {
	if t == nil {
		return nil
	}
	return &pb.UserTier{
		UserId:              t.UserID,                     // 用户ID。
		Level:               int32(t.Level),               // 等级。
		LevelName:           t.LevelName,                  // 等级名称。
		Score:               t.Score,                      // 成长值。
		NextLevelScore:      t.NextLevelScore,             // 下一级所需成长值。
		ProgressToNextLevel: t.ProgressToNextLevel,        // 升级进度。
		DiscountRate:        t.DiscountRate,               // 折扣率。
		Points:              t.Points,                     // 当前积分。
		CreatedAt:           timestamppb.New(t.CreatedAt), // 创建时间。
		UpdatedAt:           timestamppb.New(t.UpdatedAt), // 更新时间。
	}
}

// convertPointsLogToProto 是一个辅助函数，将领域层的 PointsLog 实体转换为 protobuf 的 PointsLog 消息。
func convertPointsLogToProto(l *entity.PointsLog) *pb.PointsLog {
	if l == nil {
		return nil
	}
	return &pb.PointsLog{
		Id:        uint64(l.ID),                 // ID。
		UserId:    l.UserID,                     // 用户ID。
		Points:    l.Points,                     // 变动积分。
		Reason:    l.Reason,                     // 原因。
		Type:      l.Type,                       // 类型。
		CreatedAt: timestamppb.New(l.CreatedAt), // 创建时间。
	}
}

// convertExchangeItemToProto 是一个辅助函数，将领域层的 Exchange 实体转换为 protobuf 的 ExchangeItem 消息。
func convertExchangeItemToProto(e *entity.Exchange) *pb.ExchangeItem {
	if e == nil {
		return nil
	}
	return &pb.ExchangeItem{
		Id:             uint64(e.ID),                 // ID。
		Name:           e.Name,                       // 名称。
		Description:    e.Description,                // 描述。
		RequiredPoints: e.RequiredPoints,             // 所需积分。
		Stock:          e.Stock,                      // 库存。
		CreatedAt:      timestamppb.New(e.CreatedAt), // 创建时间。
		UpdatedAt:      timestamppb.New(e.UpdatedAt), // 更新时间。
	}
}
