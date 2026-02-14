package interfaces

import (
	"context"
	"strconv"

	pb "github.com/wyfcoding/ecommerce/go-api/livestream/v1"
	"github.com/wyfcoding/ecommerce/internal/livestream/application"
	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LivestreamHandler struct {
	pb.UnimplementedLivestreamServiceServer
	app *application.LivestreamApplicationService
}

func NewLivestreamHandler(app *application.LivestreamApplicationService) *LivestreamHandler {
	return &LivestreamHandler{app: app}
}

func (h *LivestreamHandler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	room, err := h.app.CreateRoom(ctx, req.OwnerId, req.Title, "", req.CoverUrl)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRoomResponse{
		RoomId: room.RoomID,
		Status: string(room.Status),
	}, nil
}

func (h *LivestreamHandler) StartStream(ctx context.Context, req *pb.StartStreamRequest) (*pb.StartStreamResponse, error) {
	streamURL := "rtmp://placeholder/" + req.RoomId
	playURL := "https://placeholder/" + req.RoomId
	if err := h.app.StartRoom(ctx, req.RoomId, streamURL, playURL); err != nil {
		return nil, err
	}
	return &pb.StartStreamResponse{StreamUrl: streamURL}, nil
}

func (h *LivestreamHandler) EndStream(ctx context.Context, req *pb.EndStreamRequest) (*pb.EndStreamResponse, error) {
	if err := h.app.EndRoom(ctx, req.RoomId); err != nil {
		return nil, err
	}
	return &pb.EndStreamResponse{Success: true}, nil
}

func (h *LivestreamHandler) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	room, err := h.app.GetRoom(ctx, req.RoomId)
	if err != nil {
		return nil, err
	}
	return &pb.GetRoomResponse{
		Room: toProtoRoom(room),
	}, nil
}

func (h *LivestreamHandler) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	rooms, _, err := h.app.ListRooms(ctx, req.Status, int(req.Page), int(req.Size))
	if err != nil {
		return nil, err
	}
	resp := &pb.ListRoomsResponse{
		Rooms: make([]*pb.RoomDetail, 0, len(rooms)),
	}
	for _, room := range rooms {
		resp.Rooms = append(resp.Rooms, toProtoRoom(room))
	}
	return resp, nil
}

func (h *LivestreamHandler) AddProductToStream(ctx context.Context, req *pb.AddProductToStreamRequest) (*pb.AddProductToStreamResponse, error) {
	if err := h.app.AddProduct(ctx, req.RoomId, req.ProductId, "", "", 0, 0, 0); err != nil {
		return nil, err
	}
	return &pb.AddProductToStreamResponse{Success: true}, nil
}

func toProtoRoom(room *domain.Room) *pb.RoomDetail {
	products := make([]*pb.StreamProduct, 0, len(room.Products))
	for _, p := range room.Products {
		products = append(products, &pb.StreamProduct{
			ProductId: p.ProductID,
			Price:     strconv.FormatUint(p.LivePrice, 10),
			Stock:     p.Stock,
		})
	}
	return &pb.RoomDetail{
		RoomId:      room.RoomID,
		OwnerId:     room.OwnerID,
		Title:       room.Title,
		Status:      string(room.Status),
		ViewerCount: room.ViewerCount,
		CreatedAt:   timestamppb.New(room.CreatedAt),
		Products:    products,
	}
}
