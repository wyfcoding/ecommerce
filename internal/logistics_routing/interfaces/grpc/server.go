package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/logistics_routing/v1"           // 导入物流路由模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/application"   // 导入物流路由模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity" // 导入物流路由模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 LogisticsRoutingService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedLogisticsRoutingServiceServer                                      // 嵌入生成的UnimplementedLogisticsRoutingServiceServer，确保前向兼容性。
	app                                           *application.LogisticsRoutingService // 依赖LogisticsRouting应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 LogisticsRouting gRPC 服务端实例。
func NewServer(app *application.LogisticsRoutingService) *Server {
	return &Server{app: app}
}

// RegisterCarrier 处理注册配送商的gRPC请求。
// req: 包含配送商名称、类型、费用模型、支持区域等信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RegisterCarrier(ctx context.Context, req *pb.RegisterCarrierRequest) (*emptypb.Empty, error) {
	// 将protobuf请求转换为领域实体所需的 Carrier 实体。
	carrier := &entity.Carrier{
		Name:              req.Name,
		Type:              req.Type,
		BaseCost:          req.BaseCost,
		WeightRate:        req.WeightRate,
		DistanceRate:      req.DistanceRate,
		BaseDeliveryTime:  req.BaseDeliveryTime,
		SupportedRegions:  req.SupportedRegions, // 直接映射StringArray。
		AvailableCapacity: req.AvailableCapacity,
		Rating:            5.0,  // 默认评分。
		IsActive:          true, // 默认激活。
	}

	// 调用应用服务层注册配送商。
	if err := s.app.RegisterCarrier(ctx, carrier); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register carrier: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// OptimizeRoute 处理优化配送路线的gRPC请求。
// req: 包含订单ID列表的请求体。
// 返回优化后的路线响应和可能发生的gRPC错误。
func (s *Server) OptimizeRoute(ctx context.Context, req *pb.OptimizeRouteRequest) (*pb.OptimizeRouteResponse, error) {
	route, err := s.app.OptimizeRoute(ctx, req.OrderIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to optimize route: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.OptimizeRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

// GetRoute 处理获取优化路由详情的gRPC请求。
// req: 包含路由ID的请求体。
// 返回优化路由响应和可能发生的gRPC错误。
func (s *Server) GetRoute(ctx context.Context, req *pb.GetRouteRequest) (*pb.GetRouteResponse, error) {
	route, err := s.app.GetRoute(ctx, req.Id)
	if err != nil {
		// 如果路由未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("route not found: %v", err))
	}

	return &pb.GetRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

// ListCarriers 处理列出配送商的gRPC请求。
// _ 是一个空消息类型，表示请求体为空。
// 返回配送商列表响应和可能发生的gRPC错误。
func (s *Server) ListCarriers(ctx context.Context, _ *emptypb.Empty) (*pb.ListCarriersResponse, error) {
	// 调用应用服务层获取配送商列表。
	// 当前应用服务层的 ListCarriers 默认返回所有配送商。
	carriers, err := s.app.ListCarriers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list carriers: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbCarriers := make([]*pb.Carrier, len(carriers))
	for i, c := range carriers {
		pbCarriers[i] = convertCarrierToProto(c)
	}

	return &pb.ListCarriersResponse{
		Carriers: pbCarriers,
	}, nil
}

// convertCarrierToProto 是一个辅助函数，将领域层的 Carrier 实体转换为 protobuf 的 Carrier 消息。
func convertCarrierToProto(c *entity.Carrier) *pb.Carrier {
	if c == nil {
		return nil
	}
	return &pb.Carrier{
		Id:                uint64(c.ID),                 // 配送商ID。
		Name:              c.Name,                       // 名称。
		Type:              c.Type,                       // 类型。
		BaseCost:          c.BaseCost,                   // 基础费用。
		WeightRate:        c.WeightRate,                 // 每kg费用。
		DistanceRate:      c.DistanceRate,               // 每km费用。
		BaseDeliveryTime:  c.BaseDeliveryTime,           // 基础配送时间。
		SupportedRegions:  c.SupportedRegions,           // 支持地区。
		AvailableCapacity: c.AvailableCapacity,          // 可用容量。
		Rating:            c.Rating,                     // 评分。
		IsActive:          c.IsActive,                   // 是否激活。
		CreatedAt:         timestamppb.New(c.CreatedAt), // 创建时间。
		UpdatedAt:         timestamppb.New(c.UpdatedAt), // 更新时间。
	}
}

// convertRouteToProto 是一个辅助函数，将领域层的 OptimizedRoute 实体转换为 protobuf 的 OptimizedRoute 消息。
func convertRouteToProto(r *entity.OptimizedRoute) *pb.OptimizedRoute {
	if r == nil {
		return nil
	}
	// 转换关联的 RouteOrder。
	orders := make([]*pb.RouteOrder, len(r.Orders))
	for i, o := range r.Orders {
		orders[i] = &pb.RouteOrder{
			OrderId:       o.OrderID,       // 订单ID。
			CarrierId:     o.CarrierID,     // 配送商ID。
			CarrierName:   o.CarrierName,   // 配送商名称。
			EstimatedCost: o.EstimatedCost, // 预估成本。
			EstimatedTime: o.EstimatedTime, // 预估时间。
		}
	}
	return &pb.OptimizedRoute{
		Id:          uint64(r.ID),                 // 优化路由ID。
		Orders:      orders,                       // 订单列表。
		OrderCount:  r.OrderCount,                 // 订单数量。
		TotalCost:   r.TotalCost,                  // 总费用。
		AverageCost: r.AverageCost,                // 平均费用。
		CreatedAt:   timestamppb.New(r.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(r.UpdatedAt), // 更新时间。
	}
}
