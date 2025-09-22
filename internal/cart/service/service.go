package service

import (
	"context"
	"strconv"

	v1 "ecommerce/api/cart/v1"
	v1Product "ecommerce/api/product/v1" // Added product v1 import
	"ecommerce/internal/cart/biz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CartService 是 gRPC 服务的实现。
type CartService struct {
	v1.UnimplementedCartServer
	uc *biz.CartUsecase
}

// NewCartService 是 CartService 的构造函数。
func NewCartService(uc *biz.CartUsecase) *CartService {
	return &CartService{uc: uc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

func bizSkuInfoToProto(skuInfo *biz.SkuInfo) *v1Product.SkuInfo {
	if skuInfo == nil {
		return nil
	}
	return &v1Product.SkuInfo{
		SkuId:  skuInfo.SkuID,
		SpuId:  skuInfo.SpuID,
		Title:  skuInfo.Title,
		Price:  skuInfo.Price,
		Image:  skuInfo.Image,
		Specs:  skuInfo.Specs,
		Status: skuInfo.Status,
	}
}

// bizCartToProto 将 biz.UsecaseCartItem 领域模型转换为 v1.CartItem API 模型。
func bizCartToProto(cartItem *biz.UsecaseCartItem) *v1.CartItem {
	if cartItem == nil || cartItem.SkuInfo == nil {
		return nil
	}
	return &v1.CartItem{
		SkuId:    cartItem.SkuID,
		Quantity: cartItem.Quantity,
		Checked:  cartItem.Checked,
		SkuInfo:  bizSkuInfoToProto(cartItem.SkuInfo),
	}
}

// AddItem 添加商品到购物车。
func (s *CartService) AddItem(ctx context.Context, req *v1.AddItemRequest) (*v1.AddItemResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.SkuId == 0 || req.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "sku_id and quantity are required")
	}

	checked := true
	if req.HasChecked() {
		checked = req.GetChecked()
	}

	err = s.uc.AddItem(ctx, userID, req.SkuId, req.Quantity, checked)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add item: %v", err)
	}
	return &v1.AddItemResponse{}, nil
}

// GetCart 获取用户的购物车。
func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.GetCartResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.uc.GetCartDetails(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}

	protoItems := make([]*v1.CartItem, 0, len(items))
	for _, item := range items {
		protoItems = append(protoItems, bizCartToProto(item))
	}

	return &v1.GetCartResponse{Items: protoItems}, nil
}

// UpdateItem 更新购物车中的商品（数量、选中状态）。
func (s *CartService) UpdateItem(ctx context.Context, req *v1.UpdateItemRequest) (*v1.UpdateItemResponse, error) {
	if req.UserId == 0 || req.SkuId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and sku_id are required")
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if userID != req.UserId {
		return nil, status.Errorf(codes.Unauthenticated, "无权操作其他用户的购物车")
	}

	// 更新数量
	if req.HasQuantity() {
		if err := s.uc.UpdateItemInCart(ctx, req.UserId, req.SkuId, req.GetQuantity()); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update quantity: %v", err)
		}
	}

	// 更新勾选状态
	if req.HasChecked() {
		if err := s.uc.UpdateItemCheckStatus(ctx, req.UserId, []uint64{req.SkuId}, req.GetChecked()); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update check status: %v", err)
		}
	}

	return &v1.UpdateItemResponse{}, nil
}

// RemoveItem 从购物车移除商品。
func (s *CartService) RemoveItem(ctx context.Context, req *v1.RemoveItemRequest) (*v1.RemoveItemResponse, error) {
	if req.UserId == 0 || len(req.SkuIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and sku_ids are required")
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if userID != req.UserId {
		return nil, status.Errorf(codes.Unauthenticated, "无权操作其他用户的购物车")
	}

	if err := s.uc.RemoveItemsFromCart(ctx, req.UserId, req.SkuIds...); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove items: %v", err)
	}

	return &v1.RemoveItemResponse{}, nil
}

// GetCartItemCount 获取购物车商品总数。
func (s *CartService) GetCartItemCount(ctx context.Context, req *v1.GetCartItemCountRequest) (*v1.GetCartItemCountResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := s.uc.GetCartItemCount(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get item count: %v", err)
	}

	return &v1.GetCartItemCountResponse{Count: count}, nil
}
