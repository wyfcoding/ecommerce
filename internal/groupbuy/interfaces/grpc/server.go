package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/groupbuy/v1"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 GroupbuyService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedGroupbuyServiceServer
	app *application.GroupbuyService
}

// NewServer 创建并返回一个新的 Groupbuy gRPC 服务端实例。
func NewServer(app *application.GroupbuyService) *Server {
	return &Server{app: app}
}

// CreateGroupbuy 处理创建拼团活动的gRPC请求。
func (s *Server) CreateGroupbuy(ctx context.Context, req *pb.CreateGroupbuyRequest) (*pb.CreateGroupbuyResponse, error) {
	groupbuy, err := s.app.CreateGroupbuy(
		ctx,
		req.Name,
		req.ProductId,
		req.SkuId,
		req.OriginalPrice,
		req.GroupPrice,
		req.MinPeople,
		req.MaxPeople,
		req.TotalStock,
		req.StartTime.AsTime(),
		req.EndTime.AsTime(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create groupbuy: %v", err))
	}

	return &pb.CreateGroupbuyResponse{
		Groupbuy: convertGroupbuyToProto(groupbuy),
	}, nil
}

// ListGroupbuys 处理列出拼团活动的gRPC请求。
func (s *Server) ListGroupbuys(ctx context.Context, req *pb.ListGroupbuysRequest) (*pb.ListGroupbuysResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	groupbuys, total, err := s.app.ListGroupbuys(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list groupbuys: %v", err))
	}

	pbGroupbuys := make([]*pb.Groupbuy, len(groupbuys))
	for i, g := range groupbuys {
		pbGroupbuys[i] = convertGroupbuyToProto(g)
	}

	return &pb.ListGroupbuysResponse{
		Groupbuys:  pbGroupbuys,
		TotalCount: uint64(total),
	}, nil
}

// InitiateTeam 处理发起拼团团队的gRPC请求。
func (s *Server) InitiateTeam(ctx context.Context, req *pb.InitiateTeamRequest) (*pb.InitiateTeamResponse, error) {
	team, order, err := s.app.InitiateTeam(ctx, req.GroupbuyId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initiate team: %v", err))
	}

	return &pb.InitiateTeamResponse{
		Team:  convertTeamToProto(team),
		Order: convertOrderToProto(order),
	}, nil
}

// JoinTeam 处理加入拼团团队的gRPC请求。
func (s *Server) JoinTeam(ctx context.Context, req *pb.JoinTeamRequest) (*pb.JoinTeamResponse, error) {
	order, err := s.app.JoinTeam(ctx, req.TeamNo, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to join team: %v", err))
	}

	return &pb.JoinTeamResponse{
		Order: convertOrderToProto(order),
	}, nil
}

// GetTeamDetails 处理获取拼团团队详情的gRPC请求。
func (s *Server) GetTeamDetails(ctx context.Context, req *pb.GetTeamDetailsRequest) (*pb.GetTeamDetailsResponse, error) {
	team, orders, err := s.app.GetTeamDetails(ctx, req.TeamId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get team details: %v", err))
	}

	pbOrders := make([]*pb.GroupbuyOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.GetTeamDetailsResponse{
		Team:   convertTeamToProto(team),
		Orders: pbOrders,
	}, nil
}

// convertGroupbuyToProto 将领域层的 Groupbuy 实体转换为 protobuf 的 Groupbuy 消息。
func convertGroupbuyToProto(g *domain.Groupbuy) *pb.Groupbuy {
	if g == nil {
		return nil
	}
	return &pb.Groupbuy{
		Id:            uint64(g.ID),
		Name:          g.Name,
		ProductId:     g.ProductID,
		SkuId:         g.SkuID,
		OriginalPrice: g.OriginalPrice,
		GroupPrice:    g.GroupPrice,
		MinPeople:     g.MinPeople,
		MaxPeople:     g.MaxPeople,
		TotalStock:    g.TotalStock,
		SoldCount:     g.SoldCount,
		StartTime:     timestamppb.New(g.StartTime),
		EndTime:       timestamppb.New(g.EndTime),
		Status:        int32(g.Status),
		Description:   g.Description,
	}
}

// convertTeamToProto 将领域层的 GroupbuyTeam 实体转换为 protobuf 的 GroupbuyTeam 消息。
func convertTeamToProto(t *domain.GroupbuyTeam) *pb.GroupbuyTeam {
	if t == nil {
		return nil
	}
	resp := &pb.GroupbuyTeam{
		Id:            uint64(t.ID),
		GroupbuyId:    t.GroupbuyID,
		TeamNo:        t.TeamNo,
		LeaderId:      t.LeaderID,
		CurrentPeople: t.CurrentPeople,
		MaxPeople:     t.MaxPeople,
		Status:        int32(t.Status),
		ExpireAt:      timestamppb.New(t.ExpireAt),
	}
	if t.SuccessAt != nil {
		resp.SuccessAt = timestamppb.New(*t.SuccessAt)
	}
	return resp
}

// convertOrderToProto 将领域层的 GroupbuyOrder 实体转换为 protobuf 的 GroupbuyOrder 消息。
func convertOrderToProto(o *domain.GroupbuyOrder) *pb.GroupbuyOrder {
	if o == nil {
		return nil
	}
	resp := &pb.GroupbuyOrder{
		Id:          uint64(o.ID),
		GroupbuyId:  o.GroupbuyID,
		TeamId:      o.TeamID,
		TeamNo:      o.TeamNo,
		UserId:      o.UserID,
		ProductId:   o.ProductID,
		SkuId:       o.SkuID,
		Price:       o.Price,
		Quantity:    o.Quantity,
		TotalAmount: o.TotalAmount,
		IsLeader:    o.IsLeader,
		Status:      int32(o.Status),
	}
	if o.PaidAt != nil {
		resp.PaidAt = timestamppb.New(*o.PaidAt)
	}
	if o.RefundedAt != nil {
		resp.RefundedAt = timestamppb.New(*o.RefundedAt)
	}
	return resp
}
