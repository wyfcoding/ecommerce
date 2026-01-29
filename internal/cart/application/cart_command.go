package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartCommandService 处理购物车的写操作（增删改）。
type CartCommandService struct {
	repo      domain.CartRepository
	publisher domain.EventPublisher
	logger    *slog.Logger
	query     *CartQuery // 用于获取购物车实体进行内部操作
}

// NewCartCommandService 负责处理 NewCart 相关的写操作和业务逻辑。
func NewCartCommandService(repo domain.CartRepository, publisher domain.EventPublisher, logger *slog.Logger, query *CartQuery) *CartCommandService {
	return &CartCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		query:     query,
	}
}

// AddItem 添加商品到购物车。
func (s *CartCommandService) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AddItem(productID, skuID, productName, skuName, price, quantity, imageURL)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to add item to cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CartItemAddedEvent{
		UserID:    userID,
		ProductID: productID,
		SkuID:     skuID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.item.added", event)

	s.logger.InfoContext(ctx, "item added to cart successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// UpdateItemQuantity 更新商品数量。
func (s *CartCommandService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.UpdateItemQuantity(skuID, quantity)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to update item quantity", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CartItemUpdatedEvent{
		UserID:    userID,
		SkuID:     skuID,
		Quantity:  quantity,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.item.updated", event)

	s.logger.InfoContext(ctx, "item quantity updated successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// RemoveItem 移除商品。
func (s *CartCommandService) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.RemoveItem(skuID)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove item from cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CartItemRemovedEvent{
		UserID:    userID,
		SkuIDs:    []uint64{skuID},
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.item.removed", event)

	s.logger.InfoContext(ctx, "item removed from cart successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// RemoveItems 批量移除商品 (用于下单后的购物车自动清理)
func (s *CartCommandService) RemoveItems(ctx context.Context, userID uint64, skuIDs []uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	for _, skuID := range skuIDs {
		cart.RemoveItem(skuID)
	}

	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to batch remove items from cart", "user_id", userID, "count", len(skuIDs), "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CartItemRemovedEvent{
		UserID:    userID,
		SkuIDs:    skuIDs,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.item.removed", event)

	s.logger.InfoContext(ctx, "cart items removed after checkout", "user_id", userID, "count", len(skuIDs))
	return nil
}

// ClearCart 清空购物车。
func (s *CartCommandService) ClearCart(ctx context.Context, userID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.Clear()
	if err := s.repo.Clear(ctx, uint64(cart.ID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear cart", "user_id", userID, "cart_id", cart.ID, "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CartClearedEvent{
		UserID:    userID,
		Timestamp: time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.cleared", event)

	s.logger.InfoContext(ctx, "cart cleared successfully", "user_id", userID, "cart_id", cart.ID)
	return nil
}

// MergeCarts 将来源用户的购物车项合并到目标用户的购物车中（常用于登录后的匿名购物车迁移）。
func (s *CartCommandService) MergeCarts(ctx context.Context, sourceUserID, targetUserID uint64) error {
	sourceCart, err := s.repo.GetByUserID(ctx, sourceUserID)
	if err != nil {
		return err
	}
	if sourceCart == nil || len(sourceCart.Items) == 0 {
		return nil
	}

	targetCart, err := s.query.GetCart(ctx, targetUserID)
	if err != nil {
		return err
	}

	for _, item := range sourceCart.Items {
		targetCart.AddItem(item.ProductID, item.SkuID, item.ProductName, item.SkuName, item.Price, item.Quantity, item.ProductImageURL)
	}

	if err := s.repo.Save(ctx, targetCart); err != nil {
		s.logger.ErrorContext(ctx, "failed to save target cart after merge", "target_user_id", targetUserID, "error", err)
		return err
	}

	if err := s.repo.Clear(ctx, uint64(sourceCart.ID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear source cart after merge", "source_user_id", sourceUserID, "error", err)
	}

	// 发布领域事件
	event := &domain.CartsMergedEvent{
		SourceUserID: sourceUserID,
		TargetUserID: targetUserID,
		Timestamp:    time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.merged", event)

	s.logger.InfoContext(ctx, "carts merged successfully", "source_user_id", sourceUserID, "target_user_id", targetUserID)
	return nil
}

// ApplyCoupon 在购物车中记录要使用的优惠券码。
func (s *CartCommandService) ApplyCoupon(ctx context.Context, userID uint64, couponCode string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AppliedCouponCode = couponCode
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to apply coupon to cart", "user_id", userID, "coupon_code", couponCode, "error", err)
		return err
	}

	// 发布领域事件
	event := &domain.CouponAppliedEvent{
		UserID:     userID,
		CouponCode: couponCode,
		Timestamp:  time.Now(),
	}
	_ = s.publisher.Publish(ctx, "cart.coupon.applied", event)

	s.logger.InfoContext(ctx, "coupon applied to cart", "user_id", userID, "coupon_code", couponCode)
	return nil
}

// RemoveCoupon 清除购物车中已记录的优惠券码。
func (s *CartCommandService) RemoveCoupon(ctx context.Context, userID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AppliedCouponCode = ""
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove coupon from cart", "user_id", userID, "error", err)
		return err
	}

	s.logger.InfoContext(ctx, "coupon removed from cart", "user_id", userID)
	return nil
}
