package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/order_optimization/v1"              // 导入订单优化模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/order_optimization/application"   // 导入订单优化模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain/entity" // 导入订单优化模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 OrderOptimizationService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedOrderOptimizationServiceServer                                       // 嵌入生成的UnimplementedOrderOptimizationServiceServer，确保前向兼容性。
	app                                            *application.OrderOptimizationService // 依赖OrderOptimization应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 OrderOptimization gRPC 服务端实例。
func NewServer(app *application.OrderOptimizationService) *Server {
	return &Server{app: app}
}

// MergeOrders 处理合并订单的gRPC请求。
// req: 包含用户ID和待合并订单ID列表的请求体。
// 返回合并后的订单响应和可能发生的gRPC错误。
func (s *Server) MergeOrders(ctx context.Context, req *pb.MergeOrdersRequest) (*pb.MergeOrdersResponse, error) {
	mergedOrder, err := s.app.MergeOrders(ctx, req.UserId, req.OrderIds)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to merge orders: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.MergeOrdersResponse{
		MergedOrder: convertMergedOrderToProto(mergedOrder),
	}, nil
}

// SplitOrder 处理拆分订单的gRPC请求。
// req: 包含原始订单ID的请求体。
// 返回拆分后的子订单列表响应和可能发生的gRPC错误。
func (s *Server) SplitOrder(ctx context.Context, req *pb.SplitOrderRequest) (*pb.SplitOrderResponse, error) {
	splitOrders, err := s.app.SplitOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to split order: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbSplitOrders := make([]*pb.SplitOrder, len(splitOrders))
	for i, o := range splitOrders {
		pbSplitOrders[i] = convertSplitOrderToProto(o)
	}

	return &pb.SplitOrderResponse{
		SplitOrders: pbSplitOrders,
	}, nil
}

// AllocateWarehouse 处理仓库分配的gRPC请求。
// req: 包含订单ID的请求体。
// 返回仓库分配计划响应和可能发生的gRPC错误。
func (s *Server) AllocateWarehouse(ctx context.Context, req *pb.AllocateWarehouseRequest) (*pb.AllocateWarehouseResponse, error) {
	plan, err := s.app.AllocateWarehouse(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to allocate warehouse: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.AllocateWarehouseResponse{
		Plan: convertAllocationPlanToProto(plan),
	}, nil
}

// convertMergedOrderToProto 是一个辅助函数，将领域层的 MergedOrder 实体转换为 protobuf 的 MergedOrder 消息。
func convertMergedOrderToProto(o *entity.MergedOrder) *pb.MergedOrder {
	if o == nil {
		return nil
	}
	// 转换订单项列表。
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.MergedOrder{
		Id:               uint64(o.ID),       // ID。
		UserId:           o.UserID,           // 用户ID。
		OriginalOrderIds: o.OriginalOrderIDs, // 原始订单ID列表。
		Items:            items,              // 订单项列表。
		TotalAmount:      o.TotalAmount,      // 总金额。
		DiscountAmount:   o.DiscountAmount,   // 优惠金额。
		FinalAmount:      o.FinalAmount,      // 最终金额。
		ShippingAddress: &pb.ShippingAddress{ // 配送地址。
			Name:     o.ShippingAddress.Name,
			Phone:    o.ShippingAddress.Phone,
			Province: o.ShippingAddress.Province,
			City:     o.ShippingAddress.City,
			District: o.ShippingAddress.District,
			Address:  o.ShippingAddress.Address,
		},
		Status:    o.Status,                     // 状态。
		CreatedAt: timestamppb.New(o.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(o.UpdatedAt), // 更新时间。
	}
}

// convertSplitOrderToProto 是一个辅助函数，将领域层的 SplitOrder 实体转换为 protobuf 的 SplitOrder 消息。
func convertSplitOrderToProto(o *entity.SplitOrder) *pb.SplitOrder {
	if o == nil {
		return nil
	}
	// 转换订单项列表。
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.SplitOrder{
		Id:              uint64(o.ID),      // ID。
		OriginalOrderId: o.OriginalOrderID, // 原始订单ID。
		SplitIndex:      o.SplitIndex,      // 拆分序号。
		Items:           items,             // 订单项列表。
		Amount:          o.Amount,          // 金额。
		WarehouseId:     o.WarehouseID,     // 仓库ID。
		ShippingAddress: &pb.ShippingAddress{ // 配送地址。
			Name:     o.ShippingAddress.Name,
			Phone:    o.ShippingAddress.Phone,
			Province: o.ShippingAddress.Province,
			City:     o.ShippingAddress.City,
			District: o.ShippingAddress.District,
			Address:  o.ShippingAddress.Address,
		},
		Status:    o.Status,                     // 状态。
		CreatedAt: timestamppb.New(o.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(o.UpdatedAt), // 更新时间。
	}
}

// convertAllocationPlanToProto 是一个辅助函数，将领域层的 WarehouseAllocationPlan 实体转换为 protobuf 的 WarehouseAllocationPlan 消息。
func convertAllocationPlanToProto(p *entity.WarehouseAllocationPlan) *pb.WarehouseAllocationPlan {
	if p == nil {
		return nil
	}
	// 转换分配详情列表。
	allocations := make([]*pb.WarehouseAllocation, len(p.Allocations))
	for i, a := range p.Allocations {
		allocations[i] = &pb.WarehouseAllocation{
			ProductId:   a.ProductID,   // 商品ID。
			Quantity:    a.Quantity,    // 数量。
			WarehouseId: a.WarehouseID, // 仓库ID。
			Distance:    a.Distance,    // 距离。
		}
	}

	return &pb.WarehouseAllocationPlan{
		Id:          uint64(p.ID),                 // ID。
		OrderId:     p.OrderID,                    // 订单ID。
		Allocations: allocations,                  // 分配详情列表。
		CreatedAt:   timestamppb.New(p.CreatedAt), // 创建时间。
		UpdatedAt:   timestamppb.New(p.UpdatedAt), // 更新时间。
	}
}
