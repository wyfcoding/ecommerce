package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/api/multi_channel/v1"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/application"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedMultiChannelServiceServer
	app *application.MultiChannelService
}

func NewServer(app *application.MultiChannelService) *Server {
	return &Server{app: app}
}

func (s *Server) RegisterChannel(ctx context.Context, req *pb.RegisterChannelRequest) (*pb.RegisterChannelResponse, error) {
	channel := &entity.Channel{
		Name:      req.Name,
		Type:      req.Type,
		APIKey:    req.ApiKey,
		APISecret: req.ApiSecret,
		IsEnabled: req.IsEnabled,
	}

	if err := s.app.RegisterChannel(ctx, channel); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterChannelResponse{
		Channel: convertChannelToProto(channel),
	}, nil
}

func (s *Server) ListChannels(ctx context.Context, req *pb.ListChannelsRequest) (*pb.ListChannelsResponse, error) {
	channels, err := s.app.ListChannels(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbChannels := make([]*pb.Channel, len(channels))
	for i, c := range channels {
		pbChannels[i] = convertChannelToProto(c)
	}

	return &pb.ListChannelsResponse{
		Channels: pbChannels,
	}, nil
}

func (s *Server) SyncOrders(ctx context.Context, req *pb.SyncOrdersRequest) (*emptypb.Empty, error) {
	if err := s.app.SyncOrders(ctx, req.ChannelId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	orders, total, err := s.app.ListOrders(ctx, req.ChannelId, req.Status, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbOrders := make([]*pb.LocalOrder, len(orders))
	for i, o := range orders {
		pbOrders[i] = convertOrderToProto(o)
	}

	return &pb.ListOrdersResponse{
		Orders:     pbOrders,
		TotalCount: total,
	}, nil
}

func convertChannelToProto(c *entity.Channel) *pb.Channel {
	if c == nil {
		return nil
	}
	return &pb.Channel{
		Id:        uint64(c.ID),
		Name:      c.Name,
		Type:      c.Type,
		ApiKey:    c.APIKey,
		ApiSecret: c.APISecret,
		IsEnabled: c.IsEnabled,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func convertOrderToProto(o *entity.LocalOrder) *pb.LocalOrder {
	if o == nil {
		return nil
	}
	items := make([]*pb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &pb.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Sku:         item.SKU,
		}
	}

	return &pb.LocalOrder{
		Id:             uint64(o.ID),
		ChannelId:      o.ChannelID,
		ChannelName:    o.ChannelName,
		ChannelOrderId: o.ChannelOrderID,
		Items:          items,
		TotalAmount:    o.TotalAmount,
		BuyerInfo: &pb.BuyerInfo{
			Name:    o.BuyerInfo.Name,
			Email:   o.BuyerInfo.Email,
			Phone:   o.BuyerInfo.Phone,
			Country: o.BuyerInfo.Country,
		},
		ShippingInfo: &pb.ShippingInfo{
			Address: o.ShippingInfo.Address,
			City:    o.ShippingInfo.City,
			State:   o.ShippingInfo.State,
			ZipCode: o.ShippingInfo.ZipCode,
			Country: o.ShippingInfo.Country,
		},
		Status:    o.Status,
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}
