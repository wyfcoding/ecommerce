package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/risk/v1"
	"github.com/wyfcoding/ecommerce/internal/risk/application"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server 结构体实现了 RiskService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedRiskServiceServer
	cmdService   *application.RiskSecurityCommandService
	queryService *application.RiskSecurityQueryService
}

// NewServer 创建并返回一个新的 RiskSecurity gRPC 服务端实例。
func NewServer(cmd *application.RiskSecurityCommandService, query *application.RiskSecurityQueryService) *Server {
	return &Server{cmdService: cmd, queryService: query}
}

// EvaluateRisk 处理评估风险的gRPC请求。
func (s *Server) EvaluateRisk(ctx context.Context, req *pb.EvaluateRiskRequest) (*pb.EvaluateRiskResponse, error) {
	userID, _ := strconv.ParseUint(req.UserId, 10, 64)
	amount := int64(0)
	if val, ok := req.Context["amount"]; ok {
		fmt.Sscanf(val, "%d", &amount)
	}

	result, err := s.cmdService.EvaluateRisk(ctx, userID, req.IpAddress, req.DeviceId, amount)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to evaluate risk: %v", err))
	}

	strategy := "PASS"
	if result.RiskLevel == domain.RiskLevelCritical {
		strategy = "REJECT"
	} else if result.RiskLevel == domain.RiskLevelHigh {
		strategy = "CHALLENGE"
	}

	return &pb.EvaluateRiskResponse{
		RiskLevel: result.RiskLevel.String(),
		Strategy:  strategy,
		Reason:    "Analysis completed",
	}, nil
}

// Add To Blacklist 处理将实体添加到黑名单的 gRPC 请求。
func (s *Server) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*emptypb.Empty, error) {
	duration := time.Duration(3600*24) * time.Second // Default 1 day
	if err := s.cmdService.AddToBlacklist(ctx, req.TargetType, req.TargetValue, req.Reason, duration); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add to blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RemoveFromBlacklist 处理从黑名单中移除实体的gRPC请求。
func (s *Server) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*emptypb.Empty, error) {
	if err := s.cmdService.RemoveFromBlacklistByTypeAndValue(ctx, req.TargetType, req.TargetValue); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove from blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RecordUserBehavior 处理记录用户行为的gRPC请求。
func (s *Server) RecordUserBehavior(ctx context.Context, req *pb.RecordUserBehaviorRequest) (*emptypb.Empty, error) {
	userID, _ := strconv.ParseUint(req.UserId, 10, 64)
	ip := ""
	deviceID := ""
	if req.Metadata != "" {
		var meta map[string]string
		if err := json.Unmarshal([]byte(req.Metadata), &meta); err == nil {
			ip = meta["ip"]
			deviceID = meta["device_id"]
		}
	}
	if err := s.cmdService.RecordUserBehavior(ctx, userID, ip, deviceID); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record user behavior: %v", err))
	}
	return &emptypb.Empty{}, nil
}
