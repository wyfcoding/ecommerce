package grpc

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/application"
	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 结构体实现了 RiskSecurityService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedRiskSecurityServiceServer
	app *application.RiskService
}

// NewServer 创建并返回一个新的 RiskSecurity gRPC 服务端实例。
func NewServer(app *application.RiskService) *Server {
	return &Server{app: app}
}

// EvaluateRisk 处理评估风险的gRPC请求。
func (s *Server) EvaluateRisk(ctx context.Context, req *pb.EvaluateRiskRequest) (*pb.EvaluateRiskResponse, error) {
	result, err := s.app.EvaluateRisk(ctx, req.UserId, req.Ip, req.DeviceId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to evaluate risk: %v", err))
	}

	return &pb.EvaluateRiskResponse{
		Result: convertResultToProto(result),
	}, nil
}

// AddToBlacklist 处理将实体添加到黑名单的gRPC请求。
func (s *Server) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*emptypb.Empty, error) {
	duration := time.Duration(req.DurationSeconds) * time.Second
	if err := s.app.AddToBlacklist(ctx, req.Type, req.Value, req.Reason, duration); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add to blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RemoveFromBlacklist 处理从黑名单中移除实体的gRPC请求。
func (s *Server) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*emptypb.Empty, error) {
	if err := s.app.RemoveFromBlacklist(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove from blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RecordUserBehavior 处理记录用户行为的gRPC请求。
func (s *Server) RecordUserBehavior(ctx context.Context, req *pb.RecordUserBehaviorRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordUserBehavior(ctx, req.UserId, req.Ip, req.DeviceId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record user behavior: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertResultToProto 是一个辅助函数，将领域层的 RiskAnalysisResult 实体转换为 protobuf 的 RiskAnalysisResult 消息。
func convertResultToProto(r *domain.RiskAnalysisResult) *pb.RiskAnalysisResult {
	if r == nil {
		return nil
	}
	return &pb.RiskAnalysisResult{
		UserId:        r.UserID,
		RiskScore:     r.RiskScore,
		RiskLevel:     int32(r.RiskLevel),
		RiskItemsJson: r.RiskItems,
	}
}
