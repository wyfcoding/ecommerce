package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/cart/domain/entity"     // 导入购物车领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/cart/domain/repository" // 导入购物车领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// CartService 结构体定义了购物车相关的应用服务。
// 它协调领域层和基础设施层，处理购物车的创建、商品添加、更新、移除以及清空等业务逻辑。
type CartService struct {
	repo   repository.CartRepository // 依赖CartRepository接口，用于数据持久化操作。
	logger *slog.Logger              // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewCartService 创建并返回一个新的 CartService 实例。
func NewCartService(repo repository.CartRepository, logger *slog.Logger) *CartService {
	return &CartService{
		repo:   repo,
		logger: logger,
	}
}

// GetCart 获取指定用户ID的购物车。
// 如果用户还没有购物车，则会自动创建一个新的购物车。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回用户的购物车实体和可能发生的错误。
func (s *CartService) GetCart(ctx context.Context, userID uint64) (*entity.Cart, error) {
	// 尝试从仓储获取用户的购物车。
	cart, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	// 如果用户没有购物车，则创建一个新的。
	if cart == nil {
		cart = entity.NewCart(userID)
		// 保存新创建的购物车到仓储。
		if err := s.repo.Save(ctx, cart); err != nil {
			s.logger.ErrorContext(ctx, "failed to create cart", "user_id", userID, "error", err)
			return nil, err
		}
		s.logger.InfoContext(ctx, "cart created successfully", "user_id", userID, "cart_id", cart.ID)
	}
	return cart, nil
}

// AddItem 添加商品到购物车。
// 如果购物车中已存在该SKU的商品，则更新其数量；否则添加新商品。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// productID: 商品ID。
// skuID: SKU ID。
// productName: 商品名称。
// skuName: SKU名称。
// price: 商品单价。
// quantity: 添加到购物车的数量。
// imageURL: 商品图片URL。
// 返回可能发生的错误。
func (s *CartService) AddItem(ctx context.Context, userID uint64, productID, skuID uint64, productName, skuName string, price float64, quantity int32, imageURL string) error {
	// 获取用户的购物车。
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 调用购物车实体的方法添加商品。
	cart.AddItem(productID, skuID, productName, skuName, price, quantity, imageURL)
	// 保存更新后的购物车到仓储。
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to add item to cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item added to cart successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// UpdateItemQuantity 更新购物车中指定SKU商品的数量。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// skuID: 待更新商品的SKU ID。
// quantity: 更新后的商品数量。
// 返回可能发生的错误。
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uint64, skuID uint64, quantity int32) error {
	// 获取用户的购物车。
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 调用购物车实体的方法更新商品数量。
	cart.UpdateItemQuantity(skuID, quantity)
	// 保存更新后的购物车到仓储。
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to update item quantity", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item quantity updated successfully", "user_id", userID, "sku_id", skuID, "quantity", quantity)
	return nil
}

// RemoveItem 从购物车中移除指定SKU的商品。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// skuID: 待移除商品的SKU ID。
// 返回可能发生的错误。
func (s *CartService) RemoveItem(ctx context.Context, userID uint64, skuID uint64) error {
	// 获取用户的购物车。
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 调用购物车实体的方法移除商品。
	cart.RemoveItem(skuID)
	// 保存更新后的购物车到仓储。
	if err := s.repo.Save(ctx, cart); err != nil {
		s.logger.ErrorContext(ctx, "failed to remove item from cart", "user_id", userID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "item removed from cart successfully", "user_id", userID, "sku_id", skuID)
	return nil
}

// ClearCart 清空指定用户的所有购物车商品。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回可能发生的错误。
func (s *CartService) ClearCart(ctx context.Context, userID uint64) error {
	// 获取用户的购物车。
	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		return err
	}

	// 调用购物车实体的方法清空购物车。
	cart.Clear()
	// 通过仓储接口清空数据库中的购物车记录。
	if err := s.repo.Clear(ctx, uint64(cart.ID)); err != nil {
		s.logger.ErrorContext(ctx, "failed to clear cart", "user_id", userID, "cart_id", cart.ID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "cart cleared successfully", "user_id", userID, "cart_id", cart.ID)
	return nil
}
