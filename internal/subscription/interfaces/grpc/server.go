package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。

	pb "github.com/wyfcoding/ecommerce/go-api/subscription/v1"           // 导入订阅模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/subscription/application"   // 导入订阅模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity" // 导入订阅模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 SubscriptionService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedSubscriptionServiceServer                                  // 嵌入生成的UnimplementedSubscriptionServiceServer，确保前向兼容性。
	app                                       *application.SubscriptionService // 依赖Subscription应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Subscription gRPC 服务端实例。
func NewServer(app *application.SubscriptionService) *Server {
	return &Server{app: app}
}

// CreatePlan 处理创建订阅计划的gRPC请求。
// req: 包含计划名称、描述、价格、时长和功能列表的请求体。
// 返回创建成功的计划响应和可能发生的gRPC错误。
func (s *Server) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	plan, err := s.app.CreatePlan(ctx, req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create plan: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreatePlanResponse{
		Plan: convertPlanToProto(plan),
	}, nil
}

// ListPlans 处理列出订阅计划的gRPC请求。
// req: 空消息类型。
// 返回订阅计划列表响应和可能发生的gRPC错误。
func (s *Server) ListPlans(ctx context.Context, _ *emptypb.Empty) (*pb.ListPlansResponse, error) {
	// 调用应用服务层获取订阅计划列表（只列出启用的计划）。
	plans, err := s.app.ListPlans(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list plans: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbPlans := make([]*pb.SubscriptionPlan, len(plans))
	for i, p := range plans {
		pbPlans[i] = convertPlanToProto(p)
	}

	return &pb.ListPlansResponse{
		Plans: pbPlans,
	}, nil
}

// Subscribe 处理用户订阅计划的gRPC请求。
// req: 包含用户ID和计划ID的请求体。
// 返回订阅信息响应和可能发生的gRPC错误。
func (s *Server) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	sub, err := s.app.Subscribe(ctx, req.UserId, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to subscribe: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.SubscribeResponse{
		Subscription: convertSubscriptionToProto(sub),
	}, nil
}

// Cancel 处理取消订阅的gRPC请求。
// req: 包含订阅ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) Cancel(ctx context.Context, req *pb.CancelRequest) (*emptypb.Empty, error) {
	if err := s.app.Cancel(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel subscription: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// Renew 处理续订的gRPC请求。
// req: 包含订阅ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) Renew(ctx context.Context, req *pb.RenewRequest) (*emptypb.Empty, error) {
	if err := s.app.Renew(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to renew subscription: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListSubscriptions 处理列出订阅的gRPC请求。
// req: 包含用户ID和分页参数的请求体。
// 返回订阅列表响应和可能发生的gRPC错误。
func (s *Server) ListSubscriptions(ctx context.Context, req *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取订阅列表。
	// 注意：Proto请求中没有 status 字段来过滤订阅状态，应用服务层 ListSubscriptions 方法接受 status 指针。
	// 当前这里传递nil，表示不按状态过滤。如果需要状态过滤，需要修改Proto定义。
	subs, total, err := s.app.ListSubscriptions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list subscriptions: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbSubs := make([]*pb.Subscription, len(subs))
	for i, sub := range subs {
		pbSubs[i] = convertSubscriptionToProto(sub)
	}

	return &pb.ListSubscriptionsResponse{
		Subscriptions: pbSubs,
		TotalCount:    total, // 总记录数。
	}, nil
}

// convertPlanToProto 是一个辅助函数，将领域层的 SubscriptionPlan 实体转换为 protobuf 的 SubscriptionPlan 消息。
func convertPlanToProto(p *entity.SubscriptionPlan) *pb.SubscriptionPlan {
	if p == nil {
		return nil
	}
	return &pb.SubscriptionPlan{
		Id:          uint64(p.ID),                 // ID。
		Name:        p.Name,                       // 名称。
		Description: p.Description,                // 描述。
		Price:       p.Price,                      // 价格。
		Duration:    p.Duration,                   // 时长。
		Features:    p.Features,                   // 特性列表。
		Enabled:     p.Enabled,                    // 是否启用。
		CreatedAt:   timestamppb.New(p.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(p.UpdatedAt), // 更新时间。
	}
}

// convertSubscriptionToProto 是一个辅助函数，将领域层的 Subscription 实体转换为 protobuf 的 Subscription 消息。
func convertSubscriptionToProto(s *entity.Subscription) *pb.Subscription {
	if s == nil {
		return nil
	}
	// 转换可选的取消时间字段。
	var canceledAt *timestamppb.Timestamp
	if s.CanceledAt != nil {
		canceledAt = timestamppb.New(*s.CanceledAt)
	}

	return &pb.Subscription{
		Id:         uint64(s.ID),                 // ID。
		UserId:     s.UserID,                     // 用户ID。
		PlanId:     s.PlanID,                     // 计划ID。
		Status:     int32(s.Status),              // 状态。
		StartDate:  timestamppb.New(s.StartDate), // 开始时间。
		EndDate:    timestamppb.New(s.EndDate),   // 结束时间。
		AutoRenew:  s.AutoRenew,                  // 自动续订。
		CanceledAt: canceledAt,                   // 取消时间。
		CreatedAt:  timestamppb.New(s.CreatedAt), // 创建时间。
		UpdatedAt:  timestamppb.New(s.UpdatedAt), // 更新时间。
	}
}
