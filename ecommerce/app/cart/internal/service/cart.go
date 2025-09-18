package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/cart/v1"
	"ecommerce/ecommerce/app/cart/internal/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CartService struct {
	v1.UnimplementedCartServer
	uc *biz.CartUsecase
}

func NewCartService(uc *biz.CartUsecase) *CartService {
	return &CartService{uc: uc}
}

func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.AddItemResponse, error) {
	// user_id 来自网关解析 JWT 后的传递，这里我们信任这个值
	if req.UserId == 0 || req.SkuId == 0 || req.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, sku_id and quantity are required")
	}

	err := s.uc.AddItem(ctx, req.UserId, req.SkuId, req.Quantity)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.AddItemResponse{}, nil
}

// GetCart, UpdateItem, RemoveItem 等其他接口暂时留空，返回未实现错误
func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCart not implemented")
}

func (s *CartService) UpdateItem(ctx context.Context, req *v1.UpdateItemRequest) (*v1.UpdateItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateItem not implemented")
}

func (s *CartService) RemoveItem(ctx context.Context, req *v1.RemoveItemRequest) (*v1.RemoveItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveItem not implemented")
}

func (s *CartService) GetCartItemCount(ctx context.Context, req *v1.GetCartItemCountRequest) (*v1.GetCartItemCountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCartItemCount not implemented")
}
