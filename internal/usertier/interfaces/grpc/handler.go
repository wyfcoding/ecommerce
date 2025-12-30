package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/usertier/v1"
	"github.com/wyfcoding/ecommerce/internal/usertier/application"
	"github.com/wyfcoding/ecommerce/internal/usertier/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 UserTierService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedUserTierServiceServer
	app *application.UserTierService
}

// NewServer 创建并返回一个新的 UserTier gRPC 服务端实例。
func NewServer(app *application.UserTierService) *Server {
	return &Server{app: app}
}

// GetUserTier 处理获取用户等级的gRPC请求。
func (s *Server) GetUserTier(ctx context.Context, req *pb.GetUserTierRequest) (*pb.GetUserTierResponse, error) {
	tier, err := s.app.GetUserTier(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user tier: %v", err))
	}

	return &pb.GetUserTierResponse{
		Tier: convertTierToProto(tier),
	}, nil
}

// AddScore 处理增加用户成长值的gRPC请求。
func (s *Server) AddScore(ctx context.Context, req *pb.AddScoreRequest) (*emptypb.Empty, error) {
	if err := s.app.AddScore(ctx, req.UserId, req.Score); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add score: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// GetPoints 处理获取用户积分的gRPC请求。
func (s *Server) GetPoints(ctx context.Context, req *pb.GetPointsRequest) (*pb.GetPointsResponse, error) {
	points, err := s.app.GetPoints(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get points: %v", err))
	}

	return &pb.GetPointsResponse{
		Points: points,
	}, nil
}

// AddPoints 处理增加用户积分的gRPC请求。
func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// DeductPoints 处理扣除用户积分的gRPC请求。
func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to deduct points: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListPointsLogs 处理列出用户积分日志的gRPC请求。
func (s *Server) ListPointsLogs(ctx context.Context, req *pb.ListPointsLogsRequest) (*pb.ListPointsLogsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.ListPointsLogs(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list points logs: %v", err))
	}

	pbLogs := make([]*pb.PointsLog, len(logs))
	for i, l := range logs {
		pbLogs[i] = convertPointsLogToProto(l)
	}

	return &pb.ListPointsLogsResponse{
		Logs:       pbLogs,
		TotalCount: total,
	}, nil
}

// Exchange 处理兑换商品的gRPC请求。
func (s *Server) Exchange(ctx context.Context, req *pb.ExchangeRequest) (*emptypb.Empty, error) {
	if err := s.app.Exchange(ctx, req.UserId, req.ExchangeId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to exchange item: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListExchanges 处理列出可兑换商品列表的gRPC请求。
func (s *Server) ListExchanges(ctx context.Context, req *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	items, total, err := s.app.ListExchanges(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list exchanges: %v", err))
	}

	pbItems := make([]*pb.ExchangeItem, len(items))
	for i, item := range items {
		pbItems[i] = convertExchangeItemToProto(item)
	}

	return &pb.ListExchangesResponse{
		Items:      pbItems,
		TotalCount: total,
	}, nil
}

func convertTierToProto(t *domain.UserTier) *pb.UserTier {
	if t == nil {
		return nil
	}
	return &pb.UserTier{
		UserId:              t.UserID,
		Level:               int32(t.Level),
		LevelName:           t.LevelName,
		Score:               t.Score,
		NextLevelScore:      t.NextLevelScore,
		ProgressToNextLevel: t.ProgressToNextLevel,
		DiscountRate:        t.DiscountRate,
		Points:              t.Points,
		CreatedAt:           timestamppb.New(t.CreatedAt),
		UpdatedAt:           timestamppb.New(t.UpdatedAt),
	}
}

func convertPointsLogToProto(l *domain.PointsLog) *pb.PointsLog {
	if l == nil {
		return nil
	}
	return &pb.PointsLog{
		Id:        uint64(l.ID),
		UserId:    l.UserID,
		Points:    l.Points,
		Reason:    l.Reason,
		Type:      l.Type,
		CreatedAt: timestamppb.New(l.CreatedAt),
	}
}

func convertExchangeItemToProto(e *domain.Exchange) *pb.ExchangeItem {
	if e == nil {
		return nil
	}
	return &pb.ExchangeItem{
		Id:             uint64(e.ID),
		Name:           e.Name,
		Description:    e.Description,
		RequiredPoints: e.RequiredPoints,
		Stock:          e.Stock,
		CreatedAt:      timestamppb.New(e.CreatedAt),
		UpdatedAt:      timestamppb.New(e.UpdatedAt),
	}
}
