package grpc

import (
	"context"
	"time"

	pb "github.com/wyfcoding/ecommerce/api/risk_security/v1"
	"github.com/wyfcoding/ecommerce/internal/risk_security/application"
	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedRiskSecurityServiceServer
	app *application.RiskService
}

func NewServer(app *application.RiskService) *Server {
	return &Server{app: app}
}

func (s *Server) EvaluateRisk(ctx context.Context, req *pb.EvaluateRiskRequest) (*pb.EvaluateRiskResponse, error) {
	result, err := s.app.EvaluateRisk(ctx, req.UserId, req.Ip, req.DeviceId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.EvaluateRiskResponse{
		Result: convertResultToProto(result),
	}, nil
}

func (s *Server) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*emptypb.Empty, error) {
	duration := time.Duration(req.DurationSeconds) * time.Second
	if err := s.app.AddToBlacklist(ctx, req.Type, req.Value, req.Reason, duration); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*emptypb.Empty, error) {
	if err := s.app.RemoveFromBlacklist(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RecordUserBehavior(ctx context.Context, req *pb.RecordUserBehaviorRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordUserBehavior(ctx, req.UserId, req.Ip, req.DeviceId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func convertResultToProto(r *entity.RiskAnalysisResult) *pb.RiskAnalysisResult {
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
