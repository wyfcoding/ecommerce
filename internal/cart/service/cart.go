package service

import (
	"context"
	"fmt"
	"time"

	v1 "ecommerce/api/cart/v1"
	// 假设有 ProductService 的客户端
	// v1_product "ecommerce/api/product/v1"
	"ecommerce/internal/cart/model"
	"ecommerce/internal/cart/repository"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CartService 封装了购物车相关的业务逻辑，实现了 cart.proto 中定义的 CartServer 接口。
type CartService struct {
	// 嵌入 v1.UnimplementedCartServer 以确保向前兼容性
	v1.UnimplementedCartServer
	cartRepo    repository.CartRepo
	cartItemRepo repository.CartItemRepo
	validator   *validator.Validate
	maxCartItemQuantity int32
	cartExpiration      time.Duration

	// 假设这里有对其他服务的客户端，例如 ProductService 客户端
	// productClient v1_product.ProductClient
}

// NewCartService 是 CartService 的构造函数。
func NewCartService(
	cartRepo repository.CartRepo,
	cartItemRepo repository.CartItemRepo,
	maxCartItemQuantity int32,
	cartExpirationHours int,
) *CartService {
	return &CartService{
		cartRepo:            cartRepo,
		cartItemRepo:        cartItemRepo,
		validator:           validator.New(),
		maxCartItemQuantity: maxCartItemQuantity,
		cartExpiration:      time.Duration(cartExpirationHours) * time.Hour,
	}
}

// AddItemToCart 实现了向用户购物车添加商品的 RPC 方法。
// 如果商品已存在，则更新数量；如果不存在，则添加新项。
func (s *CartService) AddItemToCart(ctx context.Context, req *v1.AddItemToCartRequest) (*v1.CartInfo, error) {
	zap.S().Infof("AddItemToCart request received for user %d, product %d, sku %d, quantity %d",
		req.UserId, req.ProductId, req.SkuId, req.Quantity)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("AddItemToCart request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}
	if req.Quantity <= 0 {
		zap.S().Warn("AddItemToCart request with non-positive quantity")
		return nil, status.Errorf(codes.InvalidArgument, "quantity must be positive")
	}

	// 2. 检查商品和SKU是否存在，并获取价格、库存等信息 (此处模拟，实际应调用 ProductService)
	// skuInfo, err := s.productClient.GetSKUByID(ctx, &v1_product.GetSKUByIDRequest{Id: req.SkuId})
	// if err != nil || skuInfo == nil {
	// 	return nil, status.Errorf(codes.NotFound, "SKU %d not found", req.SkuId)
	// }
	// if skuInfo.StockQuantity < req.Quantity {
	// 	return nil, status.Errorf(codes.ResourceExhausted, "SKU %d stock not enough", req.SkuId)
	// }
	productName := fmt.Sprintf("Product-%d", req.ProductId)
	skuName := fmt.Sprintf("SKU-%d-Spec", req.SkuId)
	price := int64(10000) // 模拟价格，100.00元

	// 3. 获取用户购物车
	cart, err := s.cartRepo.GetCartByUserID(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to get cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve cart")
	}

	var cartItem *model.CartItem
	if cart == nil { // 购物车不存在，创建新购物车
		cart = &model.Cart{UserID: req.UserId}
		_, err = s.cartRepo.CreateCart(ctx, cart)
		if err != nil {
			zap.S().Errorf("failed to create new cart for user %d: %v", req.UserId, err)
			return nil, status.Errorf(codes.Internal, "failed to create cart")
		}
		zap.S().Infof("New cart created for user %d", req.UserId)
	}

	// 查找购物车中是否已存在该SKU
	for _, item := range cart.Items {
		if item.SKUID == req.SkuId {
			cartItem = item
			break
		}
	}

	// 4. 更新或添加购物车项
	if cartItem != nil { // 购物车项已存在，更新数量
		newQuantity := cartItem.Quantity + req.Quantity
		if newQuantity > s.maxCartItemQuantity {
			zap.S().Warnf("user %d attempted to add SKU %d beyond max quantity %d", req.UserId, req.SkuId, s.maxCartItemQuantity)
			return nil, status.Errorf(codes.ResourceExhausted, "item quantity exceeds maximum allowed (%d)", s.maxCartItemQuantity)
		}
		cartItem.Quantity = newQuantity
		cartItem.TotalPrice = price * int64(cartItem.Quantity)
		_, err = s.cartItemRepo.UpdateCartItem(ctx, cartItem)
		if err != nil {
			zap.S().Errorf("failed to update cart item %d for user %d: %v", cartItem.ID, req.UserId, err)
			return nil, status.Errorf(codes.Internal, "failed to update cart item")
		}
		zap.S().Infof("Cart item %d updated for user %d, new quantity: %d", cartItem.ID, req.UserId, cartItem.Quantity)
	} else { // 购物车项不存在，创建新项
		if req.Quantity > s.maxCartItemQuantity {
			zap.S().Warnf("user %d attempted to add SKU %d beyond max quantity %d", req.UserId, req.SkuId, s.maxCartItemQuantity)
			return nil, status.Errorf(codes.ResourceExhausted, "item quantity exceeds maximum allowed (%d)", s.maxCartItemQuantity)
		}
		cartItem = &model.CartItem{
			UserID:          req.UserId,
			ProductID:       req.ProductId,
			SKUID:           req.SkuId,
			ProductName:     productName,
			SKUName:         skuName,
			ProductImageURL: "http://example.com/image.jpg", // 模拟图片URL
			Price:           price,
			Quantity:        req.Quantity,
			TotalPrice:      price * int64(req.Quantity),
		}
		_, err = s.cartItemRepo.CreateCartItem(ctx, cartItem)
		if err != nil {
			zap.S().Errorf("failed to create cart item for user %d, sku %d: %v", req.UserId, req.SkuId, err)
			return nil, status.Errorf(codes.Internal, "failed to add item to cart")
		}
		zap.S().Infof("New cart item %d added for user %d, sku %d", cartItem.ID, req.UserId, req.SkuId)
	}

	// 5. 重新计算购物车总价和总数量
	updatedCart, err := s.recalculateCart(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to recalculate cart for user %d after adding item: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart totals")
	}

	zap.S().Infof("Cart for user %d updated successfully", req.UserId)
	return s.bizCartToProto(updatedCart), nil
}

// UpdateCartItem 实现了更新购物车中某个商品数量的 RPC 方法。
func (s *CartService) UpdateCartItem(ctx context.Context, req *v1.UpdateCartItemRequest) (*v1.CartInfo, error) {
	zap.S().Infof("UpdateCartItem request received for user %d, cart item %d, new quantity %d",
		req.UserId, req.CartItemId, req.Quantity)

	// 1. 参数校验
	if err := s.validator.Struct(req); err != nil {
		zap.S().Warnf("UpdateCartItem request validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid argument: %v", err)
	}
	if req.Quantity <= 0 {
		zap.S().Warn("UpdateCartItem request with non-positive quantity, attempting to remove item")
		// 如果数量为0或负数，则视为移除该商品
		_, err := s.RemoveItemFromCart(ctx, &v1.RemoveItemFromCartRequest{UserId: req.UserId, CartItemIds: []uint64{req.CartItemId}})
		return nil, err // RemoveItemFromCart 会返回 CartInfo 或错误
	}
	if req.Quantity > s.maxCartItemQuantity {
		zap.S().Warnf("user %d attempted to set cart item %d quantity beyond max %d", req.UserId, req.CartItemId, s.maxCartItemQuantity)
		return nil, status.Errorf(codes.ResourceExhausted, "item quantity exceeds maximum allowed (%d)", s.maxCartItemQuantity)
	}

	// 2. 获取购物车项
	cartItem, err := s.cartItemRepo.GetCartItemByID(ctx, req.CartItemId)
	if err != nil {
		zap.S().Errorf("failed to get cart item %d for user %d: %v", req.CartItemId, req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve cart item")
	}
	if cartItem == nil || cartItem.UserID != req.UserId {
		zap.S().Warnf("cart item %d not found or does not belong to user %d", req.CartItemId, req.UserId)
		return nil, status.Errorf(codes.NotFound, "cart item not found or access denied")
	}

	// 3. 更新数量和总价
	cartItem.Quantity = req.Quantity
	cartItem.TotalPrice = cartItem.Price * int64(req.Quantity)
	_, err = s.cartItemRepo.UpdateCartItem(ctx, cartItem)
	if err != nil {
		zap.S().Errorf("failed to update cart item %d for user %d: %v", req.CartItemId, req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart item")
	}

	// 4. 重新计算购物车总价和总数量
	updatedCart, err := s.recalculateCart(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to recalculate cart for user %d after updating item: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart totals")
	}

	zap.S().Infof("Cart item %d quantity updated to %d for user %d", req.CartItemId, req.Quantity, req.UserId)
	return s.bizCartToProto(updatedCart), nil
}

// RemoveItemFromCart 实现了从购物车中移除一个或多个商品SKU的 RPC 方法。
func (s *CartService) RemoveItemFromCart(ctx context.Context, req *v1.RemoveItemFromCartRequest) (*v1.CartInfo, error) {
	zap.S().Infof("RemoveItemFromCart request received for user %d, item IDs: %v", req.UserId, req.CartItemIds)

	// 1. 参数校验
	if req.UserId == 0 || len(req.CartItemIds) == 0 {
		zap.S().Warn("RemoveItemFromCart request with invalid user ID or empty item IDs list")
		return nil, status.Errorf(codes.InvalidArgument, "user ID cannot be zero or item IDs list cannot be empty")
	}

	// 2. 批量删除购物车项
	if err := s.cartItemRepo.DeleteCartItemsByIDs(ctx, req.UserId, req.CartItemIds); err != nil {
		zap.S().Errorf("failed to delete cart items %v for user %d: %v", req.CartItemIds, req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to remove items from cart")
	}

	// 3. 重新计算购物车总价和总数量
	updatedCart, err := s.recalculateCart(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to recalculate cart for user %d after removing items: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart totals")
	}

	zap.S().Infof("Cart items %v removed for user %d", req.CartItemIds, req.UserId)
	return s.bizCartToProto(updatedCart), nil
}

// GetCart 实现了获取指定用户购物车详细信息的 RPC 方法。
func (s *CartService) GetCart(ctx context.Context, req *v1.GetCartRequest) (*v1.CartInfo, error) {
	zap.S().Infof("GetCart request received for user %d", req.UserId)

	// 1. 参数校验
	if req.UserId == 0 {
		zap.S().Warn("GetCart request with zero user ID")
		return nil, status.Errorf(codes.InvalidArgument, "user ID cannot be zero")
	}

	// 2. 获取购物车
	cart, err := s.cartRepo.GetCartByUserID(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to get cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve cart")
	}

	// 如果购物车不存在，返回一个空的购物车信息
	if cart == nil || len(cart.Items) == 0 {
		zap.S().Infof("Cart not found or empty for user %d", req.UserId)
		return &v1.CartInfo{UserId: req.UserId}, nil
	}

	// 3. 重新计算购物车总价和总数量 (确保数据最新)
	updatedCart, err := s.recalculateCart(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to recalculate cart for user %d during get: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart totals")
	}

	zap.S().Infof("Cart retrieved for user %d, total items: %d", req.UserId, len(updatedCart.Items))
	return s.bizCartToProto(updatedCart), nil
}

// ClearCart 实现了清空指定用户购物车的 RPC 方法。
func (s *CartService) ClearCart(ctx context.Context, req *v1.ClearCartRequest) (*emptypb.Empty, error) {
	zap.S().Infof("ClearCart request received for user %d", req.UserId)

	// 1. 参数校验
	if req.UserId == 0 {
		zap.S().Warn("ClearCart request with zero user ID")
		return nil, status.Errorf(codes.InvalidArgument, "user ID cannot be zero")
	}

	// 2. 逻辑删除购物车主记录和所有购物车项
	if err := s.cartRepo.DeleteCart(ctx, req.UserId); err != nil {
		zap.S().Errorf("failed to clear cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to clear cart")
	}

	zap.S().Infof("Cart cleared for user %d successfully", req.UserId)
	return &emptypb.Empty{}, nil
}

// MergeCarts 实现了合并两个购物车的 RPC 方法。
// 通常用于用户登录后，将匿名购物车与用户已有的购物车合并。
func (s *CartService) MergeCarts(ctx context.Context, req *v1.MergeCartsRequest) (*v1.CartInfo, error) {
	zap.S().Infof("MergeCarts request received: source user %d, target user %d", req.SourceUserId, req.TargetUserId)

	// 1. 参数校验
	if req.SourceUserId == 0 || req.TargetUserId == 0 || req.SourceUserId == req.TargetUserId {
		zap.S().Warn("MergeCarts request with invalid user IDs")
		return nil, status.Errorf(codes.InvalidArgument, "invalid source or target user ID")
	}

	// 2. 获取源购物车和目标购物车
	sourceCart, err := s.cartRepo.GetCartByUserID(ctx, req.SourceUserId)
	if err != nil {
		zap.S().Errorf("failed to get source cart for user %d: %v", req.SourceUserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve source cart")
	}
	if sourceCart == nil || len(sourceCart.Items) == 0 {
		zap.S().Infof("source cart for user %d is empty or not found, no merge needed", req.SourceUserId)
		// 如果源购物车为空，直接返回目标购物车
		return s.GetCart(ctx, &v1.GetCartRequest{UserId: req.TargetUserId})
	}

	targetCart, err := s.cartRepo.GetCartByUserID(ctx, req.TargetUserId)
	if err != nil {
		zap.S().Errorf("failed to get target cart for user %d: %v", req.TargetUserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve target cart")
	}

	// 如果目标购物车不存在，将源购物车直接转移给目标用户
	if targetCart == nil {
		sourceCart.UserID = req.TargetUserId
		_, err = s.cartRepo.UpdateCart(ctx, sourceCart)
		if err != nil {
			zap.S().Errorf("failed to transfer source cart to target user %d: %v", req.TargetUserId, err)
			return nil, status.Errorf(codes.Internal, "failed to merge carts")
		}
		// 更新购物车项的用户ID
		for _, item := range sourceCart.Items {
			item.UserID = req.TargetUserId
			_, err = s.cartItemRepo.UpdateCartItem(ctx, item)
			if err != nil {
				zap.S().Errorf("failed to update cart item %d user ID to %d: %v", item.ID, req.TargetUserId, err)
				// 错误处理，可能需要回滚或重试
			}
		}
		zap.S().Infof("Source cart %d transferred to target user %d", req.SourceUserId, req.TargetUserId)
		// 清空源购物车
		_ = s.cartRepo.DeleteCart(ctx, req.SourceUserId)
		return s.bizCartToProto(sourceCart), nil
	}

	// 3. 合并购物车项
	for _, sourceItem := range sourceCart.Items {
		found := false
		for _, targetItem := range targetCart.Items {
			if sourceItem.SKUID == targetItem.SKUID {
				// SKU相同，合并数量
				newQuantity := targetItem.Quantity + sourceItem.Quantity
				if newQuantity > s.maxCartItemQuantity {
					newQuantity = s.maxCartItemQuantity // 达到最大数量限制
				}
				targetItem.Quantity = newQuantity
				targetItem.TotalPrice = targetItem.Price * int64(newQuantity)
				_, err = s.cartItemRepo.UpdateCartItem(ctx, targetItem)
				if err != nil {
					zap.S().Errorf("failed to merge update cart item %d for user %d: %v", targetItem.ID, req.TargetUserId, err)
					return nil, status.Errorf(codes.Internal, "failed to merge carts")
				}
				found = true
				break
			}
		}
		if !found { // SKU不同，将源购物车项添加到目标购物车
			sourceItem.UserID = req.TargetUserId
			_, err = s.cartItemRepo.CreateCartItem(ctx, sourceItem)
			if err != nil {
				zap.S().Errorf("failed to merge add cart item %d for user %d: %v", sourceItem.ID, req.TargetUserId, err)
				return nil, status.Errorf(codes.Internal, "failed to merge carts")
			}
		}
	}

	// 4. 清空源购物车
	_ = s.cartRepo.DeleteCart(ctx, req.SourceUserId)

	// 5. 重新计算目标购物车总价和总数量
	updatedCart, err := s.recalculateCart(ctx, req.TargetUserId)
	if err != nil {
		zap.S().Errorf("failed to recalculate target cart for user %d after merge: %v", req.TargetUserId, err)
		return nil, status.Errorf(codes.Internal, "failed to update cart totals after merge")
	}

	zap.S().Infof("Carts merged successfully from user %d to user %d", req.SourceUserId, req.TargetUserId)
	return s.bizCartToProto(updatedCart), nil
}

// ApplyCouponToCart 实现了为购物车应用优惠券的 RPC 方法。
// 此处仅为模拟，实际应调用 CouponService 进行优惠券校验和计算。
func (s *CartService) ApplyCouponToCart(ctx context.Context, req *v1.ApplyCouponToCartRequest) (*v1.CartInfo, error) {
	zap.S().Infof("ApplyCouponToCart request received for user %d, coupon: %s", req.UserId, req.CouponCode)

	// 1. 参数校验
	if req.UserId == 0 || req.CouponCode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid coupon application parameters")
	}

	// 2. 获取购物车
	cart, err := s.cartRepo.GetCartByUserID(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to get cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve cart")
	}
	if cart == nil || len(cart.Items) == 0 {
		zap.S().Warnf("cart not found or empty for user %d, cannot apply coupon", req.UserId)
		return nil, status.Errorf(codes.NotFound, "cart not found or empty")
	}

	// 3. 模拟优惠券校验和计算 (实际应调用 CouponService)
	// couponInfo, err := s.couponClient.ValidateCoupon(ctx, req.CouponCode, cart.TotalAmount)
	// if err != nil || !couponInfo.IsValid {
	// 	return nil, status.Errorf(codes.InvalidArgument, "invalid or expired coupon")
	// }
	// discountAmount := couponInfo.DiscountAmount

	discountAmount := int64(1000) // 模拟优惠10元
	if cart.TotalAmount < discountAmount {
		discountAmount = cart.TotalAmount // 优惠金额不能超过商品总价
	}

	// 4. 更新购物车优惠信息
	cart.AppliedCouponCode = req.CouponCode
	cart.DiscountAmount = discountAmount
	cart.ActualAmount = cart.TotalAmount - cart.DiscountAmount

	// 5. 调用仓库层更新购物车
	updatedCart, err := s.cartRepo.UpdateCart(ctx, cart)
	if err != nil {
		zap.S().Errorf("failed to apply coupon %s to cart for user %d: %v", req.CouponCode, req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to apply coupon")
	}

	zap.S().Infof("Coupon %s applied to cart for user %d, discount: %d", req.CouponCode, req.UserId, discountAmount)
	return s.bizCartToProto(updatedCart), nil
}

// RemoveCouponFromCart 实现了从购物车移除已应用的优惠券的 RPC 方法。
func (s *CartService) RemoveCouponFromCart(ctx context.Context, req *v1.RemoveCouponFromCartRequest) (*v1.CartInfo, error) {
	zap.S().Infof("RemoveCouponFromCart request received for user %d", req.UserId)

	// 1. 参数校验
	if req.UserId == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	// 2. 获取购物车
	cart, err := s.cartRepo.GetCartByUserID(ctx, req.UserId)
	if err != nil {
		zap.S().Errorf("failed to get cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to retrieve cart")
	}
	if cart == nil || cart.AppliedCouponCode == "" {
		zap.S().Infof("cart not found or no coupon applied for user %d", req.UserId)
		return s.bizCartToProto(cart), nil // 没有优惠券，直接返回当前购物车状态
	}

	// 3. 移除优惠券信息，并重新计算金额
	cart.AppliedCouponCode = ""
	cart.DiscountAmount = 0
	cart.ActualAmount = cart.TotalAmount // 恢复原价

	// 4. 调用仓库层更新购物车
	updatedCart, err := s.cartRepo.UpdateCart(ctx, cart)
	if err != nil {
		zap.S().Errorf("failed to remove coupon from cart for user %d: %v", req.UserId, err)
		return nil, status.Errorf(codes.Internal, "failed to remove coupon")
	}

	zap.S().Infof("Coupon removed from cart for user %d", req.UserId)
	return s.bizCartToProto(updatedCart), nil
}

// --- 辅助函数：模型转换 ---

// bizCartToProto 将 model.Cart 领域模型转换为 v1.CartInfo API 模型。
func (s *CartService) bizCartToProto(c *model.Cart) *v1.CartInfo {
	if c == nil {
		return nil
	}

	protoItems := make([]*v1.CartItem, len(c.Items))
	for i, item := range c.Items {
		protoItems[i] = s.bizCartItemToProto(&item)
	}

	return &v1.CartInfo{
		UserId:            c.UserID,
		Items:             protoItems,
		TotalQuantity:     c.TotalQuantity,
		TotalAmount:       c.TotalAmount,
		DiscountAmount:    c.DiscountAmount,
		ActualAmount:      c.ActualAmount,
		AppliedCouponCode: &v1.Google_Protobuf_StringValue{Value: c.AppliedCouponCode},
		CreatedAt:         timestamppb.New(c.CreatedAt),
		UpdatedAt:         timestamppb.New(c.UpdatedAt),
	}
}

// bizCartItemToProto 将 model.CartItem 领域模型转换为 v1.CartItem API 模型。
func (s *CartService) bizCartItemToProto(item *model.CartItem) *v1.CartItem {
	if item == nil {
		return nil
	}
	return &v1.CartItem{
		Id:              item.ID,
		UserId:          item.UserID,
		ProductId:       item.ProductID,
		SkuId:           item.SKUID,
		ProductName:     item.ProductName,
		SkuName:         item.SKUName,
		ProductImageURL: item.ProductImageURL,
		Price:           item.Price,
		Quantity:        item.Quantity,
		TotalPrice:      item.TotalPrice,
		CreatedAt:       timestamppb.New(item.CreatedAt),
		UpdatedAt:       timestamppb.New(item.UpdatedAt),
	}
}

// --- 内部辅助函数 ---

// recalculateCart 重新计算购物车总价和总数量。
func (s *CartService) recalculateCart(ctx context.Context, userID uint64) (*model.Cart, error) {
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, status.Errorf(codes.NotFound, "cart not found for user %d", userID)
	}

	var totalQuantity int32 = 0
	var totalAmount int64 = 0

	// 重新加载购物车项，确保是最新的数据
	items, err := s.cartItemRepo.ListCartItemsByUserID(ctx, userID)
	if err != nil {
		zap.S().Errorf("failed to list cart items for user %d during recalculation: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "failed to recalculate cart")
	}
	cart.Items = items // 更新购物车中的项列表

	for _, item := range cart.Items {
		totalQuantity += item.Quantity
		totalAmount += item.TotalPrice
	}

	cart.TotalQuantity = totalQuantity
	cart.TotalAmount = totalAmount

	// 重新应用优惠券 (如果存在)
	if cart.AppliedCouponCode != "" {
		// 模拟优惠券重新计算
		discountAmount := int64(1000) // 假设优惠10元
		if totalAmount < discountAmount {
			discountAmount = totalAmount
		}
		cart.DiscountAmount = discountAmount
		cart.ActualAmount = totalAmount - discountAmount
	} else {
		cart.DiscountAmount = 0
		cart.ActualAmount = totalAmount
	}

	updatedCart, err := s.cartRepo.UpdateCart(ctx, cart)
	if err != nil {
		zap.S().Errorf("failed to update cart totals for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to update cart totals: %w", err)
	}

	return updatedCart, nil
}