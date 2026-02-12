package application

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/livestream/v1"
	"github.com/wyfcoding/ecommerce/internal/livestream/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LivestreamService struct {
	repo domain.LivestreamRepository
}

func NewLivestreamService(repo domain.LivestreamRepository) *LivestreamService {
	return &LivestreamService{repo: repo}
}

func (s *LivestreamService) CreateRoom(ctx context.Context, ownerID, title, coverURL string) (*pb.CreateRoomResponse, error) {
	roomID := fmt.Sprintf("room_%d", time.Now().UnixNano())
	room := &domain.Room{
		RoomID:    roomID,
		OwnerID:   ownerID,
		Title:     title,
		CoverURL:  coverURL,
		Status:    domain.StatusCreated,
		StreamURL: "",
	}

	if err := s.repo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}

	return &pb.CreateRoomResponse{
		RoomId: roomID,
		Status: string(domain.StatusCreated),
	}, nil
}

func (s *LivestreamService) StartStream(ctx context.Context, roomID string) (*pb.StartStreamResponse, error) {
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	room.Status = domain.StatusLiving
	room.StreamURL = fmt.Sprintf("rtmp://live.wyf.com/livestream/%s", roomID)

	if err := s.repo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}

	return &pb.StartStreamResponse{
		StreamUrl: room.StreamURL,
	}, nil
}

func (s *LivestreamService) EndStream(ctx context.Context, roomID string) (*pb.EndStreamResponse, error) {
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	room.Status = domain.StatusEnded
	if err := s.repo.SaveRoom(ctx, room); err != nil {
		return nil, err
	}

	return &pb.EndStreamResponse{Success: true}, nil
}

func (s *LivestreamService) GetRoom(ctx context.Context, roomID string) (*pb.GetRoomResponse, error) {
	room, err := s.repo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, fmt.Errorf("room %s not found", roomID)
	}

	return &pb.GetRoomResponse{
		Room: mapRoomToPb(room),
	}, nil
}

func (s *LivestreamService) ListRooms(ctx context.Context, status string, page, size int32) (*pb.ListRoomsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := int((page - 1) * size)
	rooms, err := s.repo.ListRooms(ctx, status, int(size), offset)
	if err != nil {
		return nil, err
	}

	var pbRooms []*pb.RoomDetail
	for _, r := range rooms {
		pbRooms = append(pbRooms, mapRoomToPb(r))
	}

	return &pb.ListRoomsResponse{Rooms: pbRooms}, nil
}

func (s *LivestreamService) AddProductToStream(ctx context.Context, roomID, productID string) (*pb.AddProductToStreamResponse, error) {
	// 简单逻辑：直接添加。实际中需要调用 product service 校验合法性
	product := &domain.Product{
		RoomID:    roomID,
		ProductID: productID,
		Price:     "0.00", // 占位
		Stock:     100,    // 占位
	}

	if err := s.repo.AddProduct(ctx, product); err != nil {
		return nil, err
	}

	return &pb.AddProductToStreamResponse{Success: true}, nil
}

func mapRoomToPb(r *domain.Room) *pb.RoomDetail {
	var products []*pb.StreamProduct
	for _, p := range r.Products {
		products = append(products, &pb.StreamProduct{
			ProductId: p.ProductID,
			Price:     p.Price,
			Stock:     p.Stock,
		})
	}

	return &pb.RoomDetail{
		RoomId:      r.RoomID,
		OwnerId:     r.OwnerID,
		Title:       r.Title,
		Status:      string(r.Status),
		ViewerCount: r.ViewerCount,
		CreatedAt:   timestamppb.New(r.CreatedAt),
		Products:    products,
	}
}
