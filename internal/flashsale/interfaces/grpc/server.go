package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	pb "github.com/wyfcoding/ecommerce/goapi/flashsale/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 FlashSaleService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedFlashSaleServer
	app *application.FlashsaleService
}

// NewServer 创建并返回一个新的 FlashSale gRPC 服务端实例。
func NewServer(app *application.FlashsaleService) *Server {
	return &Server{app: app}
}

// CreateFlashSaleEvent 处理创建秒杀活动的gRPC请求。
func (s *Server) CreateFlashSaleEvent(ctx context.Context, req *pb.CreateFlashSaleEventRequest) (*pb.CreateFlashSaleEventResponse, error) {
	if len(req.Products) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one product is required for flash sale event")
	}
	prod := req.Products[0]

	pID, err := strconv.ParseUint(prod.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id format: %v", err))
	}
	flashPrice := int64(prod.FlashPrice * 100)

	fs, err := s.app.CreateFlashsale(ctx, req.Name, pID, pID, 0, flashPrice, prod.TotalStock, prod.MaxPerUser, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create flash sale event: %v", err))
	}

	return &pb.CreateFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

// GetFlashSaleEvent 处理获取秒杀活动详情的gRPC请求。
func (s *Server) GetFlashSaleEvent(ctx context.Context, req *pb.GetFlashSaleEventRequest) (*pb.GetFlashSaleEventResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("flash sale event not found: %v", err))
	}

	return &pb.GetFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

// ListActiveFlashSaleEvents 处理列出正在进行的秒杀活动的gRPC请求。
func (s *Server) ListActiveFlashSaleEvents(ctx context.Context, req *pb.ListActiveFlashSaleEventsRequest) (*pb.ListActiveFlashSaleEventsResponse, error) {
	statusOngoing := domain.FlashsaleStatusOngoing
	list, total, err := s.app.ListFlashsales(ctx, &statusOngoing, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list active flash sale events: %v", err))
	}

	events := make([]*pb.FlashSaleEvent, len(list))
	for i, fs := range list {
		events[i] = s.toProto(fs)
	}

	return &pb.ListActiveFlashSaleEventsResponse{
		Events:     events,
		TotalCount: int32(total),
	}, nil
}

// ParticipateInFlashSale 处理用户参与秒杀（下单）的gRPC请求。
func (s *Server) ParticipateInFlashSale(ctx context.Context, req *pb.ParticipateInFlashSaleRequest) (*pb.ParticipateInFlashSaleResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}
	eventID, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	order, err := s.app.PlaceOrder(ctx, userID, eventID, req.Quantity)
	if err != nil {
		if errors.Is(err, domain.ErrFlashsaleSoldOut) || errors.Is(err, domain.ErrFlashsaleLimit) || errors.Is(err, domain.ErrFlashsaleNotStarted) || errors.Is(err, domain.ErrFlashsaleEnded) {
			return &pb.ParticipateInFlashSaleResponse{
				Status:  "FAILED",
				Message: err.Error(),
			}, nil
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to participate in flash sale: %v", err))
	}

	return &pb.ParticipateInFlashSaleResponse{
		OrderId: strconv.FormatUint(uint64(order.ID), 10),
		Status:  "SUCCESS",
		Message: "Order placed successfully",
	}, nil
}

// GetFlashSaleProductDetails 处理获取秒杀商品详情的gRPC请求。
func (s *Server) GetFlashSaleProductDetails(ctx context.Context, req *pb.GetFlashSaleProductDetailsRequest) (*pb.GetFlashSaleProductDetailsResponse, error) {
	id, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id format")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("flash sale event not found: %v", err))
	}

	return &pb.GetFlashSaleProductDetailsResponse{
		Product: s.toProductProto(fs),
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Flashsale 实体转换为 protobuf 的 FlashSaleEvent 消息。
func (s *Server) toProto(fs *domain.Flashsale) *pb.FlashSaleEvent {
	statusStr := "UNKNOWN"
	switch fs.Status {
	case domain.FlashsaleStatusPending:
		statusStr = "UPCOMING"
	case domain.FlashsaleStatusOngoing:
		statusStr = "ACTIVE"
	case domain.FlashsaleStatusEnded:
		statusStr = "ENDED"
	case domain.FlashsaleStatusCanceled:
		statusStr = "CANCELED"
	}

	return &pb.FlashSaleEvent{
		Id:          strconv.FormatUint(uint64(fs.ID), 10),
		Name:        fs.Name,
		Description: fs.Description,
		StartTime:   timestamppb.New(fs.StartTime),
		EndTime:     timestamppb.New(fs.EndTime),
		Status:      statusStr,
		Products:    []*pb.FlashSaleProduct{s.toProductProto(fs)},
		CreatedAt:   timestamppb.New(fs.CreatedAt),
		UpdatedAt:   timestamppb.New(fs.UpdatedAt),
	}
}

// toProductProto 是一个辅助函数，将领域层的 Flashsale 实体转换为 protobuf 的 FlashSaleProduct 消息。
func (s *Server) toProductProto(fs *domain.Flashsale) *pb.FlashSaleProduct {
	return &pb.FlashSaleProduct{
		ProductId:      strconv.FormatUint(fs.ProductID, 10),
		FlashPrice:     float64(fs.FlashPrice) / 100.0,
		TotalStock:     fs.TotalStock,
		RemainingStock: fs.RemainingStock(),
		MaxPerUser:     fs.LimitPerUser,
	}
}
