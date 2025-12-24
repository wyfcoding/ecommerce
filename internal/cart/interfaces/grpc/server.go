package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/cart/v1"         // 导入购物车模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/cart/application" // 导入购物车模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/cart/domain"      // 导入购物车模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Server 结构体实现了 CartService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedCartServer                          // 嵌入生成的UnimplementedCartServer，确保前向兼容性。
	app                        *application.CartService // 依赖Cart应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Cart gRPC 服务端实例。
func NewServer(app *application.CartService) *Server {
	return &Server{app: app}
}

// AddItemToCart 处理添加商品到购物车的gRPC请求。
// req: 包含用户ID、商品ID、SKU ID和数量的请求体。
// 返回更新后的购物车信息和可能发生的gRPC错误。
func (s *Server) AddItemToCart(ctx context.Context, req *pb.AddItemToCartRequest) (*pb.CartInfo, error) {
	// TODO: 从商品服务（Product Service）获取商品的详细信息（名称、价格、图片URL）。
	// 目前，Proto请求中缺少这些字段，此处使用占位符。
	// 这是一个设计上的权衡：购物车服务是否应该调用商品服务来获取这些信息，
	// 还是客户端（调用方）应该在请求中提供这些信息。
	// 假设：为了简化，当前使用占位符。
	err := s.app.AddItem(ctx, req.UserId, req.ProductId, req.SkuId, "Unknown Product", "Unknown SKU", 0.0, req.Quantity, "")
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add item to cart: %v", err))
	}

	// 添加成功后，返回最新的购物车信息。
	// 这里直接调用GetCart方法来获取最新的购物车数据。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

// UpdateCartItem 处理更新购物车中商品数量的gRPC请求。
// req: 包含用户ID、购物车商品ID和更新数量的请求体。
// 返回更新后的购物车信息和可能发生的gRPC错误。
func (s *Server) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartInfo, error) {
	// 注意：应用服务层的UpdateItemQuantity方法使用 SkuID 作为商品标识，
	// 而Proto请求中使用 CartItemId。
	// 当前假设 req.CartItemId 即为 SkuID。如果 CartItemId 是一个独立的购物车项ID，
	// 则需要先通过 CartItemID 查找对应的 SkuID。
	err := s.app.UpdateItemQuantity(ctx, req.UserId, req.CartItemId, req.Quantity)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update cart item quantity: %v", err))
	}

	// 数量更新成功后，返回最新的购物车信息。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

// RemoveItemFromCart 处理从购物车中移除商品的gRPC请求。
// req: 包含用户ID和要移除的购物车商品ID列表的请求体。
// 返回更新后的购物车信息和可能发生的gRPC错误。
func (s *Server) RemoveItemFromCart(ctx context.Context, req *pb.RemoveItemFromCartRequest) (*pb.CartInfo, error) {
	// 遍历要移除的购物车商品ID列表。
	for _, id := range req.CartItemIds {
		// 假设 id 是 SkuID。
		if err := s.app.RemoveItem(ctx, req.UserId, id); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove item from cart: %v", err))
		}
	}
	// 移除成功后，返回最新的购物车信息。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

// GetCart 处理获取用户购物车信息的gRPC请求。
// req: 包含用户ID的请求体。
// 返回用户的购物车信息和可能发生的gRPC错误。
func (s *Server) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartInfo, error) {
	cart, err := s.app.GetCart(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get cart: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return s.toProto(cart), nil
}

// ClearCart 处理清空用户购物车的gRPC请求。
// req: 包含用户ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*emptypb.Empty, error) {
	err := s.app.ClearCart(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to clear cart: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// MergeCarts 处理合并购物车的gRPC请求。
// req: 包含源用户ID和目标用户ID的请求体。
// 返回合并后的购物车信息和可能发生的gRPC错误。
func (s *Server) MergeCarts(ctx context.Context, req *pb.MergeCartsRequest) (*pb.CartInfo, error) {
	if err := s.app.MergeCarts(ctx, req.SourceUserId, req.TargetUserId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to merge carts: %v", err))
	}

	// 合并成功后,返回目标用户的购物车信息。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.TargetUserId})
}

// ApplyCouponToCart 处理为购物车应用优惠券的gRPC请求。
// req: 包含用户ID和优惠券码的请求体。
// 返回应用优惠券后的购物车信息和可能发生的gRPC错误。
func (s *Server) ApplyCouponToCart(ctx context.Context, req *pb.ApplyCouponToCartRequest) (*pb.CartInfo, error) {
	if err := s.app.ApplyCoupon(ctx, req.UserId, req.CouponCode); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to apply coupon to cart: %v", err))
	}

	// 应用优惠券成功后,返回最新的购物车信息。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

// RemoveCouponFromCart 处理从购物车中移除优惠券的gRPC请求。
// req: 包含用户ID的请求体。
// 返回移除优惠券后的购物车信息和可能发生的gRPC错误。
func (s *Server) RemoveCouponFromCart(ctx context.Context, req *pb.RemoveCouponFromCartRequest) (*pb.CartInfo, error) {
	if err := s.app.RemoveCoupon(ctx, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove coupon from cart: %v", err))
	}

	// 移除优惠券成功后,返回最新的购物车信息。
	return s.GetCart(ctx, &pb.GetCartRequest{UserId: req.UserId})
}

// toProto 是一个辅助函数，将领域层的 Cart 实体转换为 protobuf 的 CartInfo 消息。
func (s *Server) toProto(cart *domain.Cart) *pb.CartInfo {
	// 转换购物车中的商品项列表。
	items := make([]*pb.CartItem, len(cart.Items))
	var totalQuantity int64 // 总商品数量。
	var totalAmount int64   // 总金额（以分为单位）。

	for i, item := range cart.Items {
		// 将商品价格从浮点数（元）转换为整数（分）。
		priceInt := int64(item.Price * 100)               // 转换为分。
		totalItemPrice := priceInt * int64(item.Quantity) // 计算商品项总价。

		items[i] = &pb.CartItem{
			Id:              uint64(item.ID),                 // 购物车项ID。
			UserId:          cart.UserID,                     // 用户ID。
			ProductId:       item.ProductID,                  // 商品ID。
			SkuId:           item.SkuID,                      // SKU ID。
			ProductName:     item.ProductName,                // 商品名称。
			SkuName:         item.SkuName,                    // SKU名称。
			ProductImageUrl: item.ProductImageURL,            // 商品图片URL。
			Price:           priceInt,                        // 单价（分）。
			Quantity:        item.Quantity,                   // 数量。
			TotalPrice:      totalItemPrice,                  // 总价（分）。
			CreatedAt:       timestamppb.New(item.CreatedAt), // 创建时间。
			UpdatedAt:       timestamppb.New(item.UpdatedAt), // 更新时间。
		}
		totalQuantity += int64(item.Quantity) // 累加总数量。
		totalAmount += totalItemPrice         // 累加总金额。
	}

	return &pb.CartInfo{
		UserId:            cart.UserID,                                  // 用户ID。
		Items:             items,                                        // 购物车商品项列表。
		TotalQuantity:     totalQuantity,                                // 购物车中商品总数量。
		TotalAmount:       totalAmount,                                  // 购物车中商品总金额(分)。
		DiscountAmount:    0,                                            // 优惠金额(当前未实现)。
		ActualAmount:      totalAmount,                                  // 实际支付金额(当前等于总金额)。
		AppliedCouponCode: getCouponCodeWrapper(cart.AppliedCouponCode), // 已应用的优惠券码。
		CreatedAt:         timestamppb.New(cart.CreatedAt),              // 创建时间。
		UpdatedAt:         timestamppb.New(cart.UpdatedAt),              // 更新时间。
	}
}

// getCouponCodeWrapper 将优惠券码转换为StringValue包装器。
// 如果优惠券码为空,则返回nil。
func getCouponCodeWrapper(code string) *wrapperspb.StringValue {
	if code == "" {
		return nil
	}
	return wrapperspb.String(code)
}
