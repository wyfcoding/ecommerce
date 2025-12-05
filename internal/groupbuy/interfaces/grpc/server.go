package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/groupbuy/v1"              // 导入拼团模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"   // 导入拼团模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain/entity" // 导入拼团模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 GroupbuyService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedGroupbuyServiceServer                              // 嵌入生成的UnimplementedGroupbuyServiceServer，确保前向兼容性。
	app                                   *application.GroupbuyService // 依赖Groupbuy应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Groupbuy gRPC 服务端实例。
func NewServer(app *application.GroupbuyService) *Server {
	return &Server{app: app}
}

// CreateGroupbuy 处理创建拼团活动的gRPC请求。
// req: 包含活动名称、商品信息、价格、人数和时间范围等请求体。
// 返回创建成功的拼团活动响应和可能发生的gRPC错误。
func (s *Server) CreateGroupbuy(ctx context.Context, req *pb.CreateGroupbuyRequest) (*pb.CreateGroupbuyResponse, error) {
	// 调用应用服务层创建拼团活动。
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

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateGroupbuyResponse{
		Groupbuy: convertGroupbuyToProto(groupbuy),
	}, nil
}

// ListGroupbuys 处理列出拼团活动的gRPC请求。
// req: 包含分页参数的请求体。
// 返回拼团活动列表响应和可能发生的gRPC错误。
func (s *Server) ListGroupbuys(ctx context.Context, req *pb.ListGroupbuysRequest) (*pb.ListGroupbuysResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取拼团活动列表。
	groupbuys, total, err := s.app.ListGroupbuys(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list groupbuys: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbGroupbuys := make([]*pb.Groupbuy, len(groupbuys))
	for i, g := range groupbuys {
		pbGroupbuys[i] = convertGroupbuyToProto(g)
	}

	return &pb.ListGroupbuysResponse{
		Groupbuys:  pbGroupbuys,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// InitiateTeam 处理发起拼团团队的gRPC请求。
// req: 包含拼团活动ID和用户ID的请求体。
// 返回发起的团队信息和团队发起人的订单信息。
func (s *Server) InitiateTeam(ctx context.Context, req *pb.InitiateTeamRequest) (*pb.InitiateTeamResponse, error) {
	team, order, err := s.app.InitiateTeam(ctx, req.GroupbuyId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initiate team: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.InitiateTeamResponse{
		Team:  convertTeamToProto(team),
		Order: convertOrderToProto(order),
	}, nil
}

// JoinTeam 处理加入拼团团队的gRPC请求。
// req: 包含团队编号和用户ID的请求体。
// 返回加入团队后的订单信息。
func (s *Server) JoinTeam(ctx context.Context, req *pb.JoinTeamRequest) (*pb.JoinTeamResponse, error) {
	order, err := s.app.JoinTeam(ctx, req.TeamNo, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to join team: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.JoinTeamResponse{
		Order: convertOrderToProto(order),
	}, nil
}

// GetTeamDetails 处理获取拼团团队详情的gRPC请求。
// req: 包含团队ID的请求体。
// 返回团队详情（包括团队信息和成员订单列表）。
func (s *Server) GetTeamDetails(ctx context.Context, req *pb.GetTeamDetailsRequest) (*pb.GetTeamDetailsResponse, error) {
	team, orders, err := s.app.GetTeamDetails(ctx, req.TeamId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get team details: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbOrders := make([]*pb.GroupbuyOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.GetTeamDetailsResponse{
		Team:   convertTeamToProto(team),
		Orders: pbOrders,
	}, nil
}

// convertGroupbuyToProto 是一个辅助函数，将领域层的 Groupbuy 实体转换为 protobuf 的 Groupbuy 消息。
func convertGroupbuyToProto(g *entity.Groupbuy) *pb.Groupbuy {
	if g == nil {
		return nil
	}
	return &pb.Groupbuy{
		Id:            uint64(g.ID),                 // 拼团活动ID。
		Name:          g.Name,                       // 活动名称。
		ProductId:     g.ProductID,                  // 商品ID。
		SkuId:         g.SkuID,                      // SKU ID。
		OriginalPrice: g.OriginalPrice,              // 原价。
		GroupPrice:    g.GroupPrice,                 // 拼团价。
		MinPeople:     g.MinPeople,                  // 最小成团人数。
		MaxPeople:     g.MaxPeople,                  // 最大成团人数。
		TotalStock:    g.TotalStock,                 // 总库存。
		SoldCount:     g.SoldCount,                  // 已售数量。
		StartTime:     timestamppb.New(g.StartTime), // 开始时间。
		EndTime:       timestamppb.New(g.EndTime),   // 结束时间。
		Status:        int32(g.Status),              // 状态。
		Description:   g.Description,                // 描述。
	}
}

// convertTeamToProto 是一个辅助函数，将领域层的 GroupbuyTeam 实体转换为 protobuf 的 GroupbuyTeam 消息。
func convertTeamToProto(t *entity.GroupbuyTeam) *pb.GroupbuyTeam {
	if t == nil {
		return nil
	}
	resp := &pb.GroupbuyTeam{
		Id:            uint64(t.ID),                // 团队ID。
		GroupbuyId:    t.GroupbuyID,                // 拼团活动ID。
		TeamNo:        t.TeamNo,                    // 团队编号。
		LeaderId:      t.LeaderID,                  // 团长ID。
		CurrentPeople: t.CurrentPeople,             // 当前人数。
		MaxPeople:     t.MaxPeople,                 // 最大人数。
		Status:        int32(t.Status),             // 状态。
		ExpireAt:      timestamppb.New(t.ExpireAt), // 过期时间。
	}
	if t.SuccessAt != nil {
		resp.SuccessAt = timestamppb.New(*t.SuccessAt) // 成团时间。
	}
	return resp
}

// convertOrderToProto 是一个辅助函数，将领域层的 GroupbuyOrder 实体转换为 protobuf 的 GroupbuyOrder 消息。
func convertOrderToProto(o *entity.GroupbuyOrder) *pb.GroupbuyOrder {
	if o == nil {
		return nil
	}
	resp := &pb.GroupbuyOrder{
		Id:          uint64(o.ID),    // 订单ID。
		GroupbuyId:  o.GroupbuyID,    // 拼团活动ID。
		TeamId:      o.TeamID,        // 团队ID。
		TeamNo:      o.TeamNo,        // 团队编号。
		UserId:      o.UserID,        // 用户ID。
		ProductId:   o.ProductID,     // 商品ID。
		SkuId:       o.SkuID,         // SKU ID。
		Price:       o.Price,         // 单价。
		Quantity:    o.Quantity,      // 数量。
		TotalAmount: o.TotalAmount,   // 总金额。
		IsLeader:    o.IsLeader,      // 是否团长。
		Status:      int32(o.Status), // 状态。
	}
	if o.PaidAt != nil {
		resp.PaidAt = timestamppb.New(*o.PaidAt) // 支付时间。
	}
	if o.RefundedAt != nil {
		resp.RefundedAt = timestamppb.New(*o.RefundedAt) // 退款时间。
	}
	return resp
}
