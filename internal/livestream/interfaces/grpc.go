package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/livestream/v1"
	"github.com/wyfcoding/ecommerce/internal/livestream/application"
)

type LivestreamHandler struct {
	pb.UnimplementedLivestreamServiceServer
	app *application.LivestreamService
}

func NewLivestreamHandler(app *application.LivestreamService) *LivestreamHandler {
	return &LivestreamHandler{app: app}
}

func (h *LivestreamHandler) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.CreateRoomResponse, error) {
	return h.app.CreateRoom(ctx, req.OwnerId, req.Title, req.CoverUrl)
}

func (h *LivestreamHandler) StartStream(ctx context.Context, req *pb.StartStreamRequest) (*pb.StartStreamResponse, error) {
	return h.app.StartStream(ctx, req.RoomId)
}

func (h *LivestreamHandler) EndStream(ctx context.Context, req *pb.EndStreamRequest) (*pb.EndStreamResponse, error) {
	return h.app.EndStream(ctx, req.RoomId)
}

func (h *LivestreamHandler) GetRoom(ctx context.Context, req *pb.GetRoomRequest) (*pb.GetRoomResponse, error) {
	return h.app.GetRoom(ctx, req.RoomId)
}

func (h *LivestreamHandler) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	return h.app.ListRooms(ctx, req.Status, req.Page, req.Size)
}

func (h *LivestreamHandler) AddProductToStream(ctx context.Context, req *pb.AddProductToStreamRequest) (*pb.AddProductToStreamResponse, error) {
	return h.app.AddProductToStream(ctx, req.RoomId, req.ProductId)
}
