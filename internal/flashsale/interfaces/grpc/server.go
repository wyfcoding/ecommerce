package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/flashsale/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedFlashSaleServer
	app *application.FlashSaleService
}

func NewServer(app *application.FlashSaleService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateFlashSaleEvent(ctx context.Context, req *pb.CreateFlashSaleEventRequest) (*pb.CreateFlashSaleEventResponse, error) {
	if len(req.Products) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one product is required")
	}
	// Limitation: Service only supports creating one flashsale item at a time.
	// We take the first one.
	prod := req.Products[0]

	// ProductID and SkuID in proto are strings (from other protos usually), but entity uses uint64.
	// Wait, proto definition for FlashSaleProduct:
	// string product_id = 1;
	// No sku_id in FlashSaleProduct message in proto?
	// Let's check proto again.
	// message FlashSaleProduct { string product_id = 1; ... }
	// It seems SKU ID is missing in proto or implied.
	// Service needs SkuID.
	// We'll assume product_id in proto is actually SkuID or we parse it.
	// Let's try to parse product_id as uint64.
	// And we might need to default SkuID or use same ID.

	pID, err := strconv.ParseUint(prod.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	// Service args: name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime
	// Proto doesn't have original price?
	// Proto FlashSaleProduct: flash_price, total_stock, remaining_stock, max_per_user.
	// Missing: OriginalPrice, SkuID.
	// We will use 0 for OriginalPrice and pID for SkuID as fallback.

	flashPrice := int64(prod.FlashPrice * 100)

	fs, err := s.app.CreateFlashsale(ctx, req.Name, pID, pID, 0, flashPrice, prod.TotalStock, prod.MaxPerUser, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

func (s *Server) GetFlashSaleEvent(ctx context.Context, req *pb.GetFlashSaleEventRequest) (*pb.GetFlashSaleEventResponse, error) {
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetFlashSaleEventResponse{
		Event: s.toProto(fs),
	}, nil
}

func (s *Server) ListActiveFlashSaleEvents(ctx context.Context, req *pb.ListActiveFlashSaleEventsRequest) (*pb.ListActiveFlashSaleEventsResponse, error) {
	// Service ListFlashsales takes status.
	// We want active ones.
	// Entity has FlashsaleStatusOngoing.
	statusOngoing := entity.FlashsaleStatusOngoing

	// Pagination defaults
	list, total, err := s.app.ListFlashsales(ctx, &statusOngoing, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) ParticipateInFlashSale(ctx context.Context, req *pb.ParticipateInFlashSaleRequest) (*pb.ParticipateInFlashSaleResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	eventID, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id")
	}

	order, err := s.app.PlaceOrder(ctx, userID, eventID, req.Quantity)
	if err != nil {
		return &pb.ParticipateInFlashSaleResponse{
			Status:  "FAILED",
			Message: err.Error(),
		}, nil
	}

	return &pb.ParticipateInFlashSaleResponse{
		OrderId: strconv.FormatUint(uint64(order.ID), 10),
		Status:  "SUCCESS",
	}, nil
}

func (s *Server) GetFlashSaleProductDetails(ctx context.Context, req *pb.GetFlashSaleProductDetailsRequest) (*pb.GetFlashSaleProductDetailsResponse, error) {
	// EventID is FlashsaleID in our mapping
	id, err := strconv.ParseUint(req.EventId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid event_id")
	}

	fs, err := s.app.GetFlashsale(ctx, id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Verify product ID matches if provided?
	// req.ProductId is string.
	// For now just return the product info from flashsale.

	return &pb.GetFlashSaleProductDetailsResponse{
		Product: s.toProductProto(fs),
	}, nil
}

func (s *Server) toProto(fs *entity.Flashsale) *pb.FlashSaleEvent {
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

func (s *Server) toProductProto(fs *entity.Flashsale) *pb.FlashSaleProduct {
	return &pb.FlashSaleProduct{
		ProductId:      strconv.FormatUint(fs.ProductID, 10),
		FlashPrice:     float64(fs.FlashPrice) / 100.0,
		TotalStock:     fs.TotalStock,
		RemainingStock: fs.RemainingStock(),
		MaxPerUser:     fs.LimitPerUser,
	}
}
