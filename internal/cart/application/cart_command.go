package application

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// CartCommandService 处理购物车的写操作（增删改）。
type CartCommandService struct {
	repo      domain.CartRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
	query     *CartQueryService // 用于获取购物车实体进行内部操作
}

// NewCartCommandService 负责处理 NewCart 相关的写操作和业务逻辑。
func NewCartCommandService(repo domain.CartRepository, publisher messagequeue.EventPublisher, logger *slog.Logger, query *CartQueryService) *CartCommandService {
	return &CartCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		query:     query,
	}
}

// AddItem 添加商品到购物车。
func (s *CartCommandService) AddItem(ctx context.Context, userID uint64, productID, skuID string, productName, skuName string, price float64, quantity int32, imageURL string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.AddItem(productID, skuID, productName, skuName, price, quantity, imageURL)

	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to add item to cart", "user_id", userID, "sku_id", skuID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartItemAddedEvent{
				UserID:    userID,
				ProductID: productID,
				SkuID:     skuID,
				Quantity:  quantity,
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartItemAddedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart item added event", "user_id", userID, "sku_id", skuID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "item added to cart successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// UpdateItemQuantity 更新商品数量。
func (s *CartCommandService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID string, quantity int32) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.UpdateItemQuantity(skuID, quantity)
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to update item quantity", "user_id", userID, "sku_id", skuID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartItemUpdatedEvent{
				UserID:    userID,
				SkuID:     skuID,
				Quantity:  quantity,
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartItemUpdatedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart item updated event", "user_id", userID, "sku_id", skuID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "item quantity updated successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// RemoveItem 移除商品。
func (s *CartCommandService) RemoveItem(ctx context.Context, userID uint64, skuID string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	cart.RemoveItem(skuID)
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to remove item from cart", "user_id", userID, "sku_id", skuID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartItemRemovedEvent{
				UserID:    userID,
				SkuIDs:    []string{skuID},
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartItemRemovedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart item removed event", "user_id", userID, "sku_id", skuID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "item removed from cart successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// RemoveItems 批量移除商品 (用于下单后的购物车自动清理)
func (s *CartCommandService) RemoveItems(ctx context.Context, userID uint64, skuIDs []string) error {
	cart, err := s.query.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	for _, skuID := range skuIDs {
		cart.RemoveItem(skuID)
	}

	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to batch remove items from cart", "user_id", userID, "count", len(skuIDs), "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartItemRemovedEvent{
				UserID:    userID,
				SkuIDs:    skuIDs,
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartItemRemovedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart item removed event", "user_id", userID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

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
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to clear cart", "user_id", userID, "cart_id", cart.ID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartClearedEvent{
				UserID:    userID,
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartClearedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart cleared event", "user_id", userID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

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

	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, targetCart); err != nil {
			s.logger.ErrorContext(ctx, "failed to save target cart after merge", "target_user_id", targetUserID, "error", err)
			return err
		}
		if err := s.repo.ClearInTx(ctx, tx, sourceCart.ID); err != nil {
			s.logger.ErrorContext(ctx, "failed to clear source cart after merge", "source_user_id", sourceUserID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CartsMergedEvent{
				SourceUserID: sourceUserID,
				TargetUserID: targetUserID,
				Timestamp:    time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartMergedEventType, strconv.FormatUint(targetUserID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart merged event", "target_user_id", targetUserID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

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
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to apply coupon to cart", "user_id", userID, "coupon_code", couponCode, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CouponAppliedEvent{
				UserID:     userID,
				CouponCode: couponCode,
				Timestamp:  time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartCouponAppliedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart coupon applied event", "user_id", userID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

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
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to remove coupon from cart", "user_id", userID, "error", err)
			return err
		}
		if s.publisher != nil {
			event := &domain.CouponRemovedEvent{
				UserID:    userID,
				Timestamp: time.Now(),
			}
			if err := s.publisher.PublishInTx(ctx, tx, domain.CartCouponRemovedEventType, strconv.FormatUint(userID, 10), event); err != nil {
				s.logger.ErrorContext(ctx, "failed to publish cart coupon removed event", "user_id", userID, "error", err)
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "coupon removed from cart", "user_id", userID)
	return nil
}
