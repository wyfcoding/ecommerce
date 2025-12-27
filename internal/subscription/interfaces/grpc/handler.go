package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/subscription/v1"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 SubscriptionService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedSubscriptionServer
	app *application.SubscriptionService
}

// NewServer 创建并返回一个新的 Subscription gRPC 服务端实例。
func NewServer(app *application.SubscriptionService) *Server {
	return &Server{app: app}
}

// CreatePlan 处理创建订阅计划的gRPC请求。
func (s *Server) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	plan, err := s.app.CreatePlan(ctx, req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create plan: %v", err))
	}

	return &pb.CreatePlanResponse{
		Plan: convertPlanToProto(plan),
	}, nil
}

// ListPlans 处理列出订阅计划的gRPC请求。
func (s *Server) ListPlans(ctx context.Context, _ *emptypb.Empty) (*pb.ListPlansResponse, error) {
	plans, err := s.app.ListPlans(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list plans: %v", err))
	}

	pbPlans := make([]*pb.SubscriptionPlan, len(plans))
	for i, p := range plans {
		pbPlans[i] = convertPlanToProto(p)
	}

	return &pb.ListPlansResponse{
		Plans: pbPlans,
	}, nil
}

// Subscribe 处理用户订阅计划的gRPC请求。
func (s *Server) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	sub, err := s.app.Subscribe(ctx, req.UserId, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to subscribe: %v", err))
	}

	return &pb.SubscribeResponse{
		Subscription: convertSubscriptionToProto(sub),
	}, nil
}

// Cancel 处理取消订阅的gRPC请求。
func (s *Server) Cancel(ctx context.Context, req *pb.CancelRequest) (*emptypb.Empty, error) {
	if err := s.app.Cancel(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to cancel subscription: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// Renew 处理续订的gRPC请求。
func (s *Server) Renew(ctx context.Context, req *pb.RenewRequest) (*emptypb.Empty, error) {
	if err := s.app.Renew(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to renew subscription: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListSubscriptions 处理列出订阅的gRPC请求。
func (s *Server) ListSubscriptions(ctx context.Context, req *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	subs, total, err := s.app.ListSubscriptions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list subscriptions: %v", err))
	}

	pbSubs := make([]*pb.Subscription, len(subs))
	for i, sub := range subs {
		pbSubs[i] = convertSubscriptionToProto(sub)
	}

	return &pb.ListSubscriptionsResponse{
		Subscriptions: pbSubs,
		TotalCount:    total,
	}, nil
}

// convertPlanToProto 是一个辅助函数，将领域层的 SubscriptionPlan 实体转换为 protobuf 的 SubscriptionPlan 消息。
func convertPlanToProto(p *domain.SubscriptionPlan) *pb.SubscriptionPlan {
	if p == nil {
		return nil
	}
	return &pb.SubscriptionPlan{
		Id:          uint64(p.ID),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Duration:    p.Duration,
		Features:    p.Features,
		Enabled:     p.Enabled,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// convertSubscriptionToProto 是一个辅助函数，将领域层的 Subscription 实体转换为 protobuf 的 Subscription 消息。
func convertSubscriptionToProto(s *domain.Subscription) *pb.Subscription {
	if s == nil {
		return nil
	}
	var canceledAt *timestamppb.Timestamp
	if s.CanceledAt != nil {
		canceledAt = timestamppb.New(*s.CanceledAt)
	}

	return &pb.Subscription{
		Id:         uint64(s.ID),
		UserId:     s.UserID,
		PlanId:     s.PlanID,
		Status:     int32(s.Status),
		StartDate:  timestamppb.New(s.StartDate),
		EndDate:    timestamppb.New(s.EndDate),
		AutoRenew:  s.AutoRenew,
		CanceledAt: canceledAt,
		CreatedAt:  timestamppb.New(s.CreatedAt),
		UpdatedAt:  timestamppb.New(s.UpdatedAt),
	}
}
