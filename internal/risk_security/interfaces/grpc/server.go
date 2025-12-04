package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"time"    // 导入时间库。

	pb "github.com/wyfcoding/ecommerce/api/risk_security/v1"              // 导入风控安全模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/risk_security/application"   // 导入风控安全模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/risk_security/domain/entity" // 导入风控安全模块的领域实体。

	"google.golang.org/grpc/codes"                   // gRPC状态码。
	"google.golang.org/grpc/status"                  // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb" // 导入空消息类型。
	// "google.golang.org/protobuf/types/known/timestamppb"  // 导入时间戳消息类型，此文件中未直接使用。
)

// Server 结构体实现了 RiskSecurityService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedRiskSecurityServiceServer                          // 嵌入生成的UnimplementedRiskSecurityServiceServer，确保前向兼容性。
	app                                       *application.RiskService // 依赖Risk应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 RiskSecurity gRPC 服务端实例。
func NewServer(app *application.RiskService) *Server {
	return &Server{app: app}
}

// EvaluateRisk 处理评估风险的gRPC请求。
// req: 包含用户ID、IP地址、设备ID和金额的请求体。
// 返回风险评估结果响应和可能发生的gRPC错误。
func (s *Server) EvaluateRisk(ctx context.Context, req *pb.EvaluateRiskRequest) (*pb.EvaluateRiskResponse, error) {
	result, err := s.app.EvaluateRisk(ctx, req.UserId, req.Ip, req.DeviceId, req.Amount)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to evaluate risk: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.EvaluateRiskResponse{
		Result: convertResultToProto(result),
	}, nil
}

// AddToBlacklist 处理将实体添加到黑名单的gRPC请求。
// req: 包含黑名单类型、值、原因和有效时长的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddToBlacklist(ctx context.Context, req *pb.AddToBlacklistRequest) (*emptypb.Empty, error) {
	// 将请求中的 DurationSeconds (int64) 转换为 time.Duration 类型。
	duration := time.Duration(req.DurationSeconds) * time.Second
	if err := s.app.AddToBlacklist(ctx, req.Type, req.Value, req.Reason, duration); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add to blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RemoveFromBlacklist 处理从黑名单中移除实体的gRPC请求。
// req: 包含黑名单条目ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RemoveFromBlacklist(ctx context.Context, req *pb.RemoveFromBlacklistRequest) (*emptypb.Empty, error) {
	if err := s.app.RemoveFromBlacklist(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove from blacklist: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// RecordUserBehavior 处理记录用户行为的gRPC请求。
// req: 包含用户ID、IP地址和设备ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RecordUserBehavior(ctx context.Context, req *pb.RecordUserBehaviorRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordUserBehavior(ctx, req.UserId, req.Ip, req.DeviceId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record user behavior: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertResultToProto 是一个辅助函数，将领域层的 RiskAnalysisResult 实体转换为 protobuf 的 RiskAnalysisResult 消息。
func convertResultToProto(r *entity.RiskAnalysisResult) *pb.RiskAnalysisResult {
	if r == nil {
		return nil
	}
	return &pb.RiskAnalysisResult{
		UserId:        r.UserID,           // 用户ID。
		RiskScore:     r.RiskScore,        // 风险分数。
		RiskLevel:     int32(r.RiskLevel), // 风险等级。
		RiskItemsJson: r.RiskItems,        // 风险项详情（JSON字符串）。客户端需要自行解析。
		// CreatedAt 和 UpdatedAt 字段需要从 gorm.Model 提取并转换为 timestamppb.Timestamp。
		// 这里Proto定义中没有，故未进行转换。
	}
}
