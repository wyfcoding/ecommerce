package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartManager 处理购物车的写操作（增删改）。
type CartManager struct {
	repo   domain.CartRepository
	logger *slog.Logger
	query  *CartQuery // 用于获取购物车实体进行内部操作
}

// NewCartManager 负责处理 NewCart 相关的写操作和业务逻辑。
func NewCartManager(repo domain.CartRepository, logger *slog.Logger, query *CartQuery) *CartManager {
	return &CartManager{
		repo:   repo,
		logger: logger,
		query:  query,
	}
}

// AddItem 添加商品到购物车。
func (s *CartManager) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AddItem(productID, skuID, productName, skuName, price, quantity, imageURL)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to add item to cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item added to cart successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// UpdateItemQuantity 更新商品数量。
func (s *CartManager) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.UpdateItemQuantity(skuID, quantity)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to update item quantity", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item quantity updated successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// RemoveItem 移除商品。
func (s *CartManager) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.RemoveItem(skuID)
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove item from cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item removed from cart successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// ClearCart 清空购物车。
func (s *CartManager) ClearCart(ctx context.Context, userID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.Clear()
	if err := s.repo.Clear(ctx, uint64(cart.ID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear cart", "user_id", userID, "cart_id", cart.ID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "cart cleared successfully", "user_id", userID, "cart_id", cart.ID)
	return nil
}

// MergeCarts 合并购物车。
func (s *CartManager) MergeCarts(ctx context.Context, sourceUserID, targetUserID uint64) error {
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

	return nil
}

// ApplyCoupon 应用优惠券。
func (s *CartManager) ApplyCoupon(ctx context.Context, userID uint64, couponCode string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AppliedCouponCode = couponCode
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to apply coupon to cart", "user_id", userID, "coupon_code", couponCode, "error", err)
		return err
	}

	return nil
}

// RemoveCoupon 移除优惠券。
func (s *CartManager) RemoveCoupon(ctx context.Context, userID uint64) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AppliedCouponCode = ""
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove coupon from cart", "user_id", userID, "error", err)
		return err
	}

	return nil
}
