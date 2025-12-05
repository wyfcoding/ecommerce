package grpc

import (
	"context"
	"errors"
	"fmt"     // 导入格式化包，用于错误信息。
	"strconv" // 导入字符串和数字转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/flashsale/v1"              // 导入秒杀模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"   // 导入秒杀模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity" // 导入秒杀模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 FlashSaleService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedFlashSaleServer                               // 嵌入生成的UnimplementedFlashSaleServer，确保前向兼容性。
	app                             *application.FlashSaleService // 依赖FlashSale应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 FlashSale gRPC 服务端实例。
func NewServer(app *application.FlashSaleService) *Server {
	return &Server{app: app}
}

// CreateFlashSaleEvent 处理创建秒杀活动的gRPC请求。
// req: 包含秒杀活动名称、商品信息、时间范围等请求体。
// 返回创建成功的秒杀活动响应和可能发生的gRPC错误。
func (s *Server) CreateFlashSaleEvent(ctx context.Context, req *pb.CreateFlashSaleEventRequest) (*pb.CreateFlashSaleEventResponse, error) {
	if len(req.Products) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one product is required for flash sale event")
	}
	// 当前应用服务层的 CreateFlashsale 方法一次只能处理一个秒杀商品。
	// 这里假设Proto请求 Products 列表中只包含一个商品，并取第一个进行处理。
	prod := req.Products[0]

	// Proto中的 ProductId 是字符串，但实体和应用服务期望的是 uint64。
	// 这里尝试将 Proto 的 ProductId 解析为 uint64。
	pID, err := strconv.ParseUint(prod.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id format: %v", err))
	}
	// 注意：Proto中的FlashSaleProduct消息缺少OriginalPrice和SkuID字段。
	// 这里 OriginalPrice 使用0，SkuID 暂时与 ProductID 相同作为占位符。
	// 实际生产环境需要确保这些字段的正确映射或处理。
	flashPrice := int64(prod.FlashPrice * 100) // 将浮点价格转换为整数分。

	// 调用应用服务层创建秒杀活动。
	fs, err := s.app.CreateFlashsale(ctx, req.Name, pID, pID, 0, flashPrice, prod.TotalStock, prod.MaxPerUser, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create flash sale event: %v", err))
	}

	return &pb.CreateFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

// GetFlashSaleEvent 处理获取秒杀活动详情的gRPC请求。
// req: 包含秒杀活动ID的请求体。
// 返回秒杀活动响应和可能发生的gRPC错误。
func (s *Server) GetFlashSaleEvent(ctx context.Context, req *pb.GetFlashSaleEventRequest) (*pb.GetFlashSaleEventResponse, error) {
	// 将Proto的ID（字符串）转换为uint64。
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		// 根据错误类型返回NotFound状态码。
		return nil, status.Error(codes.Internal, fmt.Sprintf("flash sale event not found: %v", err))
	}

	return &pb.GetFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

// ListActiveFlashSaleEvents 处理列出正在进行的秒杀活动的gRPC请求。
// req: 包含分页参数的请求体（当前Proto请求中没有分页字段）。
// 返回活跃的秒杀活动列表响应和可能发生的gRPC错误。
func (s *Server) ListActiveFlashSaleEvents(ctx context.Context, req *pb.ListActiveFlashSaleEventsRequest) (*pb.ListActiveFlashSaleEventsResponse, error) {
	// 应用服务层的 ListFlashsales 方法需要状态参数。
	// 这里过滤状态为进行中的秒杀活动。
	statusOngoing := entity.FlashsaleStatusOngoing
	// 注意：Proto请求中没有分页字段。此处使用默认分页参数1页100条。
	list, total, err := s.app.ListFlashsales(ctx, &statusOngoing, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list active flash sale events: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	events := make([]*pb.FlashSaleEvent, len(list))
	for i, fs := range list {
		events[i] = s.toProto(fs)
	}

	return &pb.ListActiveFlashSaleEventsResponse{
		Events:     events,
		TotalCount: int32(total), // 总记录数。
	}, nil
}

// ParticipateInFlashSale 处理用户参与秒杀（下单）的gRPC请求。
// req: 包含用户ID、秒杀活动ID和购买数量的请求体。
// 返回参与秒杀结果响应和可能发生的gRPC错误。
func (s *Server) ParticipateInFlashSale(ctx context.Context, req *pb.ParticipateInFlashSaleRequest) (*pb.ParticipateInFlashSaleResponse, error) {
	// 将Proto中的用户ID和事件ID（字符串）转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	eventID, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	// 调用应用服务层进行下单操作。
	order, err := s.app.PlaceOrder(ctx, userID, eventID, req.Quantity)
	if err != nil {
		// 根据错误类型返回不同的状态和消息。
		if errors.Is(err, entity.ErrFlashsaleSoldOut) || errors.Is(err, entity.ErrFlashsaleLimit) || errors.Is(err, entity.ErrFlashsaleNotStarted) || errors.Is(err, entity.ErrFlashsaleEnded) {
			return &pb.ParticipateInFlashSaleResponse{
				Status:  "FAILED",
				Message: err.Error(),
			}, nil
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to participate in flash sale: %v", err))
	}

	return &pb.ParticipateInFlashSaleResponse{
		OrderId: strconv.FormatUint(uint64(order.ID), 10), // 返回订单ID的字符串形式。
		Status:  "SUCCESS",
		Message: "Order placed successfully",
	}, nil
}

// GetFlashSaleProductDetails 处理获取秒杀商品详情的gRPC请求。
// req: 包含事件ID（秒杀活动ID）的请求体。
// 返回秒杀商品详情响应和可能发生的gRPC错误。
func (s *Server) GetFlashSaleProductDetails(ctx context.Context, req *pb.GetFlashSaleProductDetailsRequest) (*pb.GetFlashSaleProductDetailsResponse, error) {
	// 将Proto的EventId（字符串）转换为uint64，作为秒杀活动ID。
	id, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("flash sale event not found: %v", err))
	}

	// 注意：Proto请求中也包含 ProductId 字段（字符串），但此处未进行匹配验证。
	// 假定直接返回秒杀活动中的商品信息。

	return &pb.GetFlashSaleProductDetailsResponse{
		Product: s.toProductProto(fs), // 使用辅助函数转换为ProductProto。
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Flashsale 实体转换为 protobuf 的 FlashSaleEvent 消息。
func (s *Server) toProto(fs *entity.Flashsale) *pb.FlashSaleEvent {
	// 映射领域实体状态到protobuf状态字符串。
	statusStr := "UNKNOWN"
	switch fs.Status {
	case entity.FlashsaleStatusPending:
		statusStr = "UPCOMING"
	case entity.FlashsaleStatusOngoing:
		statusStr = "ACTIVE"
	case entity.FlashsaleStatusEnded:
		statusStr = "ENDED"
	case entity.FlashsaleStatusCanceled:
		statusStr = "CANCELED"
	}

	return &pb.FlashSaleEvent{
		Id:          strconv.FormatUint(uint64(fs.ID), 10),        // 将ID转换为字符串。
		Name:        fs.Name,                                      // 活动名称。
		Description: fs.Description,                               // 活动描述。
		StartTime:   timestamppb.New(fs.StartTime),                // 开始时间。
		EndTime:     timestamppb.New(fs.EndTime),                  // 结束时间。
		Status:      statusStr,                                    // 状态。
		Products:    []*pb.FlashSaleProduct{s.toProductProto(fs)}, // 关联的商品列表。
		CreatedAt:   timestamppb.New(fs.CreatedAt),                // 创建时间。
		UpdatedAt:   timestamppb.New(fs.UpdatedAt),                // 更新时间。
	}
}

// toProductProto 是一个辅助函数，将领域层的 Flashsale 实体转换为 protobuf 的 FlashSaleProduct 消息。
// 注意：Flashsale 实体中包含了商品信息，这里将它转换为 FlashSaleProduct 消息。
func (s *Server) toProductProto(fs *entity.Flashsale) *pb.FlashSaleProduct {
	return &pb.FlashSaleProduct{
		ProductId:      strconv.FormatUint(fs.ProductID, 10), // 商品ID。
		FlashPrice:     float64(fs.FlashPrice) / 100.0,       // 秒杀价格（分转元）。
		TotalStock:     fs.TotalStock,                        // 总库存。
		RemainingStock: fs.RemainingStock(),                  // 剩余库存。
		MaxPerUser:     fs.LimitPerUser,                      // 每用户限购数量。
		// Proto中还包含 OriginalPrice, SkuId 等字段，但实体中没有或未映射。
	}
}
