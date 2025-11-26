package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/subscription/v1"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedSubscriptionServiceServer
	app *application.SubscriptionService
}

func NewServer(app *application.SubscriptionService) *Server {
	return &Server{app: app}
}

func (s *Server) CreatePlan(ctx context.Context, req *pb.CreatePlanRequest) (*pb.CreatePlanResponse, error) {
	plan, err := s.app.CreatePlan(ctx, req.Name, req.Description, req.Price, req.Duration, req.Features)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreatePlanResponse{
		Plan: convertPlanToProto(plan),
	}, nil
}

func (s *Server) ListPlans(ctx context.Context, _ *emptypb.Empty) (*pb.ListPlansResponse, error) {
	plans, err := s.app.ListPlans(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbPlans := make([]*pb.SubscriptionPlan, len(plans))
	for i, p := range plans {
		pbPlans[i] = convertPlanToProto(p)
	}

	return &pb.ListPlansResponse{
		Plans: pbPlans,
	}, nil
}

func (s *Server) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	sub, err := s.app.Subscribe(ctx, req.UserId, req.PlanId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SubscribeResponse{
		Subscription: convertSubscriptionToProto(sub),
	}, nil
}

func (s *Server) Cancel(ctx context.Context, req *pb.CancelRequest) (*emptypb.Empty, error) {
	if err := s.app.Cancel(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Renew(ctx context.Context, req *pb.RenewRequest) (*emptypb.Empty, error) {
	if err := s.app.Renew(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListSubscriptions(ctx context.Context, req *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	subs, total, err := s.app.ListSubscriptions(ctx, req.UserId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func convertPlanToProto(p *entity.SubscriptionPlan) *pb.SubscriptionPlan {
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

func convertSubscriptionToProto(s *entity.Subscription) *pb.Subscription {
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
