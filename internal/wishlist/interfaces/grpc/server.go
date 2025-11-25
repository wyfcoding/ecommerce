package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/wishlist/v1"
	"github.com/wyfcoding/ecommerce/internal/wishlist/application"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedWishlistServer
	app *application.WishlistService
}

func NewServer(app *application.WishlistService) *Server {
	return &Server{app: app}
}

func (s *Server) AddItemToWishlist(ctx context.Context, req *pb.AddItemToWishlistRequest) (*pb.AddItemToWishlistResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	// Missing fields in proto: SkuID, ProductName, SkuName, Price, ImageURL
	// Using placeholders. This proto needs update to be useful.
	skuID := productID // Assuming 1:1 for now or SkuID passed as ProductID

	item, err := s.app.Add(ctx, userID, productID, skuID, "Unknown Product", "Unknown SKU", 0, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.AddItemToWishlistResponse{
		Item: s.toProto(item),
	}, nil
}

func (s *Server) RemoveItemFromWishlist(ctx context.Context, req *pb.RemoveItemFromWishlistRequest) (*pb.RemoveItemFromWishlistResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	// Service Remove takes (userID, id). ID here is likely SkuID based on Add usage.
	err = s.app.Remove(ctx, userID, productID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RemoveItemFromWishlistResponse{
		Success: true,
		Message: "Item removed",
	}, nil
}

func (s *Server) ListWishlistItems(ctx context.Context, req *pb.ListWishlistItemsRequest) (*pb.ListWishlistItemsResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	items, total, err := s.app.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbItems := make([]*pb.WishlistItem, len(items))
	for i, item := range items {
		pbItems[i] = s.toProto(item)
	}

	return &pb.ListWishlistItemsResponse{
		Items:      pbItems,
		TotalCount: int32(total),
	}, nil
}

func (s *Server) toProto(item *entity.Wishlist) *pb.WishlistItem {
	return &pb.WishlistItem{
		Id:        strconv.FormatUint(uint64(item.ID), 10),
		UserId:    strconv.FormatUint(item.UserID, 10),
		ProductId: strconv.FormatUint(item.ProductID, 10),
		AddedAt:   timestamppb.New(item.CreatedAt),
	}
}
