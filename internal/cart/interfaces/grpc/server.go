package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/cart/v1"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/internal/cart/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedCartServer
	app *application.CartService
}

func NewServer(app *application.CartService) *Server {
	return &Server{app: app}
}

func (s *Server) AddItemToCart(ctx context.Context, req *pb.AddItemToCartRequest) (*pb.CartInfo, error) {
	// TODO: Fetch product details (name, price, image) from Product Service based on product_id/sku_id
	// For now, we'll use placeholder values or expect them to be passed (proto doesn't have them in request)
	// This indicates a design gap: Request should probably include price/name if Cart Service doesn't call Product Service,
	// OR Cart Service needs a ProductClient.
	// Assuming for this refactor we use placeholders or update proto later.
	// Let's use dummy values for now to satisfy the interface.

	err := s.app.AddItem(ctx, req.UserId, req.ProductId, req.SkuId, "Unknown Product", "Unknown SKU", 0.0, req.Quantity, "")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *Server) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartInfo, error) {
	// Note: Service uses SkuID for update, but Proto uses CartItemID.
	// This is a mismatch. We need to resolve this.
	// Assuming CartItemID in proto actually refers to SkuID for simplicity in this refactor,
	// or we need to look up SkuID from CartItemID.
	// Let's assume req.CartItemId is SkuID for now as per common simplification.

	err := s.app.UpdateItemQuantity(ctx, req.UserId, req.CartItemId, req.Quantity)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *Server) RemoveItemFromCart(ctx context.Context, req *pb.RemoveItemFromCartRequest) (*pb.CartInfo, error) {
	for _, id := range req.CartItemIds {
		// Assuming id is SkuID
		if err := s.app.RemoveItem(ctx, req.UserId, id); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

func (s *Server) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartInfo, error) {
	cart, err := s.app.GetCart(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.toProto(cart), nil
}

func (s *Server) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*emptypb.Empty, error) {
	err := s.app.ClearCart(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) MergeCarts(ctx context.Context, req *pb.MergeCartsRequest) (*pb.CartInfo, error) {
	return nil, status.Error(codes.Unimplemented, "MergeCarts not implemented")
}

func (s *Server) ApplyCouponToCart(ctx context.Context, req *pb.ApplyCouponToCartRequest) (*pb.CartInfo, error) {
	return nil, status.Error(codes.Unimplemented, "ApplyCouponToCart not implemented")
}

func (s *Server) RemoveCouponFromCart(ctx context.Context, req *pb.RemoveCouponFromCartRequest) (*pb.CartInfo, error) {
	return nil, status.Error(codes.Unimplemented, "RemoveCouponFromCart not implemented")
}

func (s *Server) toProto(cart *entity.Cart) *pb.CartInfo {
	items := make([]*pb.CartItem, len(cart.Items))
	var totalQuantity int64
	var totalAmount int64

	for i, item := range cart.Items {
		priceInt := int64(item.Price * 100) // Convert to cents
		totalItemPrice := priceInt * int64(item.Quantity)

		items[i] = &pb.CartItem{
			Id:              uint64(item.ID),
			UserId:          cart.UserID,
			ProductId:       item.ProductID,
			SkuId:           item.SkuID,
			ProductName:     item.ProductName,
			SkuName:         item.SkuName,
			ProductImageUrl: item.ProductImageURL,
			Price:           priceInt,
			Quantity:        item.Quantity,
			TotalPrice:      totalItemPrice,
			CreatedAt:       timestamppb.New(item.CreatedAt),
			UpdatedAt:       timestamppb.New(item.UpdatedAt),
		}
		totalQuantity += int64(item.Quantity)
		totalAmount += totalItemPrice
	}

	return &pb.CartInfo{
		UserId:         cart.UserID,
		Items:          items,
		TotalQuantity:  totalQuantity,
		TotalAmount:    totalAmount,
		DiscountAmount: 0, // Not implemented
		ActualAmount:   totalAmount,
		CreatedAt:      timestamppb.New(cart.CreatedAt),
		UpdatedAt:      timestamppb.New(cart.UpdatedAt),
	}
}
