package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/logistics/v1"          // 导入物流模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/logistics/application" // 导入物流模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/logistics/domain"      // 导入物流模块的领域层。
	"github.com/wyfcoding/pkg/algorithm"                            // 导入算法包，用于路线优化。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 LogisticsService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedLogisticsServiceServer                               // 嵌入生成的UnimplementedLogisticsServiceServer，确保前向兼容性。
	app                                    *application.LogisticsService // 依赖Logistics应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Logistics gRPC 服务端实例。
func NewServer(app *application.LogisticsService) *Server {
	return &Server{app: app}
}

// CreateLogistics 处理创建物流单的gRPC请求。
// req: 包含订单信息、跟踪号、承运商和发收件人信息的请求体。
// 返回created successfully的物流单响应和可能发生的gRPC错误。
func (s *Server) CreateLogistics(ctx context.Context, req *pb.CreateLogisticsRequest) (*pb.CreateLogisticsResponse, error) {
	// 调用应用服务层创建物流单.
	logistics, err := s.app.CreateLogistics(
		ctx,
		req.OrderId,
		req.OrderNo,
		req.TrackingNo,
		req.Carrier,
		req.CarrierCode,
		req.SenderName,
		req.SenderPhone,
		req.SenderAddress,
		req.SenderLat,
		req.SenderLon,
		req.ReceiverName,
		req.ReceiverPhone,
		req.ReceiverAddress,
		req.ReceiverLat,
		req.ReceiverLon,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create logistics: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

// GetLogistics 处理获取物流单详情的gRPC请求。
// req: 包含物流单ID的请求体。
// 返回物流单响应和可能发生的gRPC错误。
func (s *Server) GetLogistics(ctx context.Context, req *pb.GetLogisticsRequest) (*pb.GetLogisticsResponse, error) {
	logistics, err := s.app.GetLogistics(ctx, req.Id)
	if err != nil {
		// 如果物流单未找到,返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("logistics not found: %v", err))
	}
	return &pb.GetLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

// GetLogisticsByTrackingNo 处理根据运单号获取物流单详情的gRPC请求。
// req: 包含运单号的请求体。
// 返回物流单响应和可能发生的gRPC错误。
func (s *Server) GetLogisticsByTrackingNo(ctx context.Context, req *pb.GetLogisticsByTrackingNoRequest) (*pb.GetLogisticsResponse, error) {
	logistics, err := s.app.GetLogisticsByTrackingNo(ctx, req.TrackingNo)
	if err != nil {
		// 如果物流单未找到,返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("logistics not found by tracking no: %v", err))
	}
	return &pb.GetLogisticsResponse{
		Logistics: convertLogisticsToProto(logistics),
	}, nil
}

// UpdateStatus 处理更新物流状态的gRPC请求.
// req: 包含物流单ID、新的状态、位置和描述的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*emptypb.Empty, error) {
	// 将Proto的Status（int32）转换为实体LogisticsStatus。
	if err := s.app.UpdateStatus(ctx, req.Id, domain.LogisticsStatus(req.Status), req.Location, req.Description); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update logistics status: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// AddTrace 处理添加物流轨迹的gRPC请求。
// req: 包含物流单ID、位置、描述和状态的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) AddTrace(ctx context.Context, req *pb.AddTraceRequest) (*emptypb.Empty, error) {
	if err := s.app.AddTrace(ctx, req.Id, req.Location, req.Description, req.Status); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add trace: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// SetEstimatedTime 处理设置预计送达时间的gRPC请求.
// req: 包含物流单ID和预计时间的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) SetEstimatedTime(ctx context.Context, req *pb.SetEstimatedTimeRequest) (*emptypb.Empty, error) {
	if err := s.app.SetEstimatedTime(ctx, req.Id, req.EstimatedTime.AsTime()); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to set estimated time: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListLogistics 处理列出物流单的gRPC请求。
// req: 包含分页参数的请求体。
// 返回物流单列表响应和可能发生的gRPC错误。
func (s *Server) ListLogistics(ctx context.Context, req *pb.ListLogisticsRequest) (*pb.ListLogisticsResponse, error) {
	// 获取分页参数。
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取物流单列表。
	logisticsList, total, err := s.app.ListLogistics(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list logistics: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbLogistics := make([]*pb.Logistics, len(logisticsList))
	for i, l := range logisticsList {
		pbLogistics[i] = convertLogisticsToProto(l)
	}

	return &pb.ListLogisticsResponse{
		Logistics:  pbLogistics,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// OptimizeDeliveryRoute 处理优化配送路线的gRPC请求.
// req: 包含物流单ID和目的地列表的请求体。
// 返回优化后的配送路线响应和可能发生的gRPC错误。
func (s *Server) OptimizeDeliveryRoute(ctx context.Context, req *pb.OptimizeDeliveryRouteRequest) (*pb.OptimizeDeliveryRouteResponse, error) {
	// 将protobuf的目的地列表转换为算法层所需的 Location 结构体列表.
	destinations := make([]algorithm.Location, len(req.Destinations))
	for i, d := range req.Destinations {
		destinations[i] = algorithm.Location{
			ID:  d.Id,
			Lat: d.Lat,
			Lon: d.Lon,
		}
	}

	route, err := s.app.OptimizeDeliveryRoute(ctx, req.LogisticsId, destinations)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to optimize delivery route: %v", err))
	}

	return &pb.OptimizeDeliveryRouteResponse{
		Route: convertRouteToProto(route),
	}, nil
}

// convertLogisticsToProto 是一个辅助函数,将领域层的 Logistics 实体转换为 protobuf 的 Logistics 消息。
func convertLogisticsToProto(l *domain.Logistics) *pb.Logistics {
	if l == nil {
		return nil
	}
	// 转换关联的Traces。
	traces := make([]*pb.LogisticsTrace, len(l.Traces))
	for i, t := range l.Traces {
		traces[i] = convertTraceToProto(t)
	}
	// 创建并填充Logistics protobuf消息。
	resp := &pb.Logistics{
		Id:              uint64(l.ID),                 // 物流ID。
		OrderId:         l.OrderID,                    // 订单ID。
		OrderNo:         l.OrderNo,                    // 订单号。
		TrackingNo:      l.TrackingNo,                 // 运单号。
		Carrier:         l.Carrier,                    // 承运商。
		CarrierCode:     l.CarrierCode,                // 承运商编码。
		SenderName:      l.SenderName,                 // 发件人姓名。
		SenderPhone:     l.SenderPhone,                // 发件人电话。
		SenderAddress:   l.SenderAddress,              // 发件人地址。
		SenderLat:       l.SenderLat,                  // 发件人纬度。
		SenderLon:       l.SenderLon,                  // 发件人经度。
		ReceiverName:    l.ReceiverName,               // 收件人姓名。
		ReceiverPhone:   l.ReceiverPhone,              // 收件人电话。
		ReceiverAddress: l.ReceiverAddress,            // 收件人地址。
		ReceiverLat:     l.ReceiverLat,                // 收件人纬度。
		ReceiverLon:     l.ReceiverLon,                // 收件人经度。
		Status:          int32(l.Status),              // 状态。
		CurrentLocation: l.CurrentLocation,            // 当前位置。
		Traces:          traces,                       // 轨迹列表。
		Route:           convertRouteToProto(l.Route), // 配送路线。
		CreatedAt:       timestamppb.New(l.CreatedAt), // 创建时间。
		UpdatedAt:       timestamppb.New(l.UpdatedAt), // 更新时间。
	}
	// 映射可选的时间字段.
	if l.EstimatedTime != nil {
		resp.EstimatedTime = timestamppb.New(*l.EstimatedTime)
	}
	if l.DeliveredAt != nil {
		resp.DeliveredAt = timestamppb.New(*l.DeliveredAt)
	}
	return resp
}

// convertTraceToProto 是一个辅助函数,将领域层的 LogisticsTrace 实体转换为 protobuf 的 LogisticsTrace 消息。
func convertTraceToProto(t *domain.LogisticsTrace) *pb.LogisticsTrace {
	if t == nil {
		return nil
	}
	return &pb.LogisticsTrace{
		Id:          uint64(t.ID),                 // 轨迹ID。
		LogisticsId: t.LogisticsID,                // 物流ID。
		TrackingNo:  t.TrackingNo,                 // 运单号。
		Location:    t.Location,                   // 位置。
		Description: t.Description,                // 描述。
		Status:      t.Status,                     // 状态描述。
		CreatedAt:   timestamppb.New(t.CreatedAt), // 创建时间。
	}
}

// convertRouteToProto 是一个辅助函数,将领域层的 DeliveryRoute 实体转换为 protobuf 的 DeliveryRoute 消息。
func convertRouteToProto(r *domain.DeliveryRoute) *pb.DeliveryRoute {
	if r == nil {
		return nil
	}
	return &pb.DeliveryRoute{
		Id:          uint64(r.ID),  // 路线ID。
		LogisticsId: r.LogisticsID, // 物流ID。
		RouteData:   r.RouteData,   // 路线数据（JSON）。
		Distance:    r.Distance,    // 总距离。
	}
}
