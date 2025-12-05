package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/multi_channel/v1"           // 导入多渠道模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/multi_channel/application"   // 导入多渠道模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity" // 导入多渠道模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 MultiChannelService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedMultiChannelServiceServer                                  // 嵌入生成的UnimplementedMultiChannelServiceServer，确保前向兼容性。
	app                                       *application.MultiChannelService // 依赖MultiChannel应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 MultiChannel gRPC 服务端实例。
func NewServer(app *application.MultiChannelService) *Server {
	return &Server{app: app}
}

// RegisterChannel 处理注册渠道的gRPC请求。
// req: 包含渠道名称、类型、API凭证和启用状态的请求体。
// 返回创建成功的渠道响应和可能发生的gRPC错误。
func (s *Server) RegisterChannel(ctx context.Context, req *pb.RegisterChannelRequest) (*pb.RegisterChannelResponse, error) {
	// 将protobuf请求转换为领域实体所需的 Channel 实体。
	channel := &entity.Channel{
		Name:      req.Name,
		Type:      req.Type,
		APIKey:    req.ApiKey,
		APISecret: req.ApiSecret,
		IsEnabled: req.IsEnabled,
	}

	// 调用应用服务层注册渠道。
	if err := s.app.RegisterChannel(ctx, channel); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to register channel: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.RegisterChannelResponse{
		Channel: convertChannelToProto(channel),
	}, nil
}

// ListChannels 处理列出渠道的gRPC请求。
// req: 包含分页参数的请求体（当前为空）。
// 返回渠道列表响应和可能发生的gRPC错误。
func (s *Server) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	// 调用应用服务层获取渠道列表。
	// 当前应用服务层的 ListChannels 默认返回所有渠道。
	channels, err := s.app.ListChannels(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list channels: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbChannels := make([]*pb.Channel, len(channels))
	for i, c := range channels {
		pbChannels[i] = convertChannelToProto(c)
	}

	return &pb.ListChannelsResponse{
		Channels: pbChannels,
	}, nil
}

// SyncOrders 处理同步订单的gRPC请求。
// req: 包含渠道ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) SyncOrders(ctx context.Context, req *pb.SyncOrdersRequest) (*emptypb.Empty, error) {
	if err := s.app.SyncOrders(ctx, req.ChannelId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to sync orders: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListOrders 处理列出订单的gRPC请求。
// req: 包含渠道ID、状态和分页参数的请求体。
// 返回订单列表响应和可能发生的gRPC错误。
func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取订单列表。
	orders, total, err := s.app.ListOrders(ctx, req.ChannelId, req.Status, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list orders: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbOrders := make([]*pb.LocalOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:     pbOrders,
		TotalCount: total, // 总记录数。
	}, nil
}

// convertChannelToProto 是一个辅助函数，将领域层的 Channel 实体转换为 protobuf 的 Channel 消息。
func convertChannelToProto(c *entity.Channel) *pb.Channel {
	if c == nil {
		return nil
	}
	return &pb.Channel{
		Id:        uint64(c.ID),                 // 渠道ID。
		Name:      c.Name,                       // 名称。
		Type:      c.Type,                       // 类型。
		ApiKey:    c.APIKey,                     // API Key。
		ApiSecret: c.APISecret,                  // API Secret。
		IsEnabled: c.IsEnabled,                  // 是否启用。
		CreatedAt: timestamppb.New(c.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(c.UpdatedAt), // 更新时间。
	}
}

// convertOrderToProto 是一个辅助函数，将领域层的 LocalOrder 实体转换为 protobuf 的 LocalOrder 消息。
func convertOrderToProto(o *entity.LocalOrder) *pb.LocalOrder {
	if o == nil {
		return nil
	}
	// 转换关联的 OrderItem。
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId:   item.ProductID,   // 商品ID。
			ProductName: item.ProductName, // 商品名称。
			Quantity:    item.Quantity,    // 数量。
			Price:       item.Price,       // 价格。
			Sku:         item.SKU,         // SKU。
		}
	}

	// 转换 BuyerInfo 和 ShippingInfo。
	return &pb.LocalOrder{
		Id:             uint64(o.ID),     // 订单ID。
		ChannelId:      o.ChannelID,      // 渠道ID。
		ChannelName:    o.ChannelName,    // 渠道名称。
		ChannelOrderId: o.ChannelOrderID, // 渠道订单ID。
		Items:          items,            // 商品项列表。
		TotalAmount:    o.TotalAmount,    // 总金额。
		BuyerInfo: &pb.BuyerInfo{ // 买家信息。
			Name:    o.BuyerInfo.Name,
			Email:   o.BuyerInfo.Email,
			Phone:   o.BuyerInfo.Phone,
			Country: o.BuyerInfo.Country,
		},
		ShippingInfo: &pb.ShippingInfo{ // 配送信息。
			Address: o.ShippingInfo.Address,
			City:    o.ShippingInfo.City,
			State:   o.ShippingInfo.State,
			ZipCode: o.ShippingInfo.ZipCode,
			Country: o.ShippingInfo.Country,
		},
		Status:    o.Status,                     // 状态。
		CreatedAt: timestamppb.New(o.CreatedAt), // 创建时间。
		UpdatedAt: timestamppb.New(o.UpdatedAt), // 更新时间。
	}
}
