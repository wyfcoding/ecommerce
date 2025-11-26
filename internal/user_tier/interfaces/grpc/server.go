package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/user_tier/v1"
	"github.com/wyfcoding/ecommerce/internal/user_tier/application"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedUserTierServiceServer
	app *application.UserTierService
}

func NewServer(app *application.UserTierService) *Server {
	return &Server{app: app}
}

func (s *Server) GetUserTier(ctx context.Context, req *pb.GetUserTierRequest) (*pb.GetUserTierResponse, error) {
	tier, err := s.app.GetUserTier(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetUserTierResponse{
		Tier: convertTierToProto(tier),
	}, nil
}

func (s *Server) AddScore(ctx context.Context, req *pb.AddScoreRequest) (*emptypb.Empty, error) {
	if err := s.app.AddScore(ctx, req.UserId, req.Score); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetPoints(ctx context.Context, req *pb.GetPointsRequest) (*pb.GetPointsResponse, error) {
	points, err := s.app.GetPoints(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetPointsResponse{
		Points: points,
	}, nil
}

func (s *Server) AddPoints(ctx context.Context, req *pb.AddPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.AddPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeductPoints(ctx context.Context, req *pb.DeductPointsRequest) (*emptypb.Empty, error) {
	if err := s.app.DeductPoints(ctx, req.UserId, req.Points, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListPointsLogs(ctx context.Context, req *pb.ListPointsLogsRequest) (*pb.ListPointsLogsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	logs, total, err := s.app.ListPointsLogs(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) Exchange(ctx context.Context, req *pb.ExchangeRequest) (*emptypb.Empty, error) {
	if err := s.app.Exchange(ctx, req.UserId, req.ExchangeId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListExchanges(ctx context.Context, req *pb.ListExchangesRequest) (*pb.ListExchangesResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	items, total, err := s.app.ListExchanges(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func convertTierToProto(t *entity.UserTier) *pb.UserTier {
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

func convertPointsLogToProto(l *entity.PointsLog) *pb.PointsLog {
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

func convertExchangeItemToProto(e *entity.Exchange) *pb.ExchangeItem {
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
