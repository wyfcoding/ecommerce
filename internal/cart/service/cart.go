package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ecommerce/internal/cart/model"
	"ecommerce/internal/cart/repository"
)

var (
	ErrCartNotFound     = errors.New("购物车不存在")
	ErrCartItemNotFound = errors.New("购物车项不存在")
	ErrInvalidQuantity  = errors.New("无效的数量")
	ErrProductNotFound  = errors.New("商品不存在")
	ErrStockInsufficient = errors.New("库存不足")
)

// CartService 购物车服务接口
type CartService interface {
	// 购物车操作
	AddToCart(ctx context.Context, userID, skuID uint64, quantity uint32) (*model.CartItem, error)
	UpdateCartItem(ctx context.Context, userID, itemID uint64, quantity uint32) (*model.CartItem, error)
	RemoveFromCart(ctx context.Context, userID, itemID uint64) error
	ClearCart(ctx context.Context, userID uint64) error
	
	// 购物车查询
	GetCart(ctx context.Context, userID uint64) ([]*model.CartItem, error)
	GetCartItemCount(ctx context.Context, userID uint64) (int64, error)
	
	// 购物车选中
	SelectCartItem(ctx context.Context, userID, itemID uint64, selected bool) error
	SelectAllCartItems(ctx context.Context, userID uint64, selected bool) error
	
	// 购物车合并（登录后）
	MergeCart(ctx context.Context, guestUserID, loginUserID uint64) error
}

type cartService struct {
	repo          repository.CartRepo
	productClient ProductClient // 商品服务客户端
	logger        *zap.Logger
}

// ProductClient 商品服务客户端接口
type ProductClient interface {
	GetSKUInfo(ctx context.Context, skuID uint64) (*SKUInfo, error)
	CheckStock(ctx context.Context, skuID uint64, quantity uint32) (bool, error)
}

// SKUInfo SKU信息
type SKUInfo struct {
	ID          uint64
	Name        string
	Price       uint64
	Stock       uint32
	ImageURL    string
	IsAvailable bool
}

// NewCartService 创建购物车服务实例
func NewCartService(repo repository.CartRepo, productClient ProductClient, logger *zap.Logger) CartService {
	return &cartService{
		repo:          repo,
		productClient: productClient,
		logger:        logger,
	}
}

// AddToCart 添加商品到购物车
func (s *cartService) AddToCart(ctx context.Context, userID, skuID uint64, quantity uint32) (*model.CartItem, error) {
	if quantity == 0 {
		return nil, ErrInvalidQuantity
	}

	// 1. 验证商品信息
	skuInfo, err := s.productClient.GetSKUInfo(ctx, skuID)
	if err != nil {
		s.logger.Error("获取SKU信息失败", zap.Error(err), zap.Uint64("skuID", skuID))
		return nil, ErrProductNotFound
	}

	if !skuInfo.IsAvailable {
		return nil, fmt.Errorf("商品已下架")
	}

	// 2. 检查库存
	hasStock, err := s.productClient.CheckStock(ctx, skuID, quantity)
	if err != nil {
		s.logger.Error("检查库存失败", zap.Error(err))
		return nil, err
	}
	if !hasStock {
		return nil, ErrStockInsufficient
	}

	// 3. 检查购物车中是否已存在该商品
	existingItem, err := s.repo.GetCartItemByUserAndSKU(ctx, userID, skuID)
	if err == nil && existingItem != nil {
		// 已存在，更新数量
		newQuantity := existingItem.Quantity + quantity
		
		// 再次检查库存
		hasStock, err := s.productClient.CheckStock(ctx, skuID, newQuantity)
		if err != nil || !hasStock {
			return nil, ErrStockInsufficient
		}

		existingItem.Quantity = newQuantity
		existingItem.UpdatedAt = time.Now()
		
		if err := s.repo.UpdateCartItem(ctx, existingItem); err != nil {
			s.logger.Error("更新购物车项失败", zap.Error(err))
			return nil, err
		}
		
		return existingItem, nil
	}

	// 4. 创建新的购物车项
	cartItem := &model.CartItem{
		UserID:      userID,
		SKUID:       skuID,
		ProductName: skuInfo.Name,
		Price:       skuInfo.Price,
		ImageURL:    skuInfo.ImageURL,
		Quantity:    quantity,
		Selected:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateCartItem(ctx, cartItem); err != nil {
		s.logger.Error("创建购物车项失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("添加到购物车成功",
		zap.Uint64("userID", userID),
		zap.Uint64("skuID", skuID),
		zap.Uint32("quantity", quantity))

	return cartItem, nil
}

// UpdateCartItem 更新购物车项数量
func (s *cartService) UpdateCartItem(ctx context.Context, userID, itemID uint64, quantity uint32) (*model.CartItem, error) {
	if quantity == 0 {
		// 数量为0，删除该项
		return nil, s.RemoveFromCart(ctx, userID, itemID)
	}

	// 1. 获取购物车项
	item, err := s.repo.GetCartItemByID(ctx, itemID)
	if err != nil {
		return nil, ErrCartItemNotFound
	}

	// 验证所属用户
	if item.UserID != userID {
		return nil, fmt.Errorf("无权操作该购物车项")
	}

	// 2. 检查库存
	hasStock, err := s.productClient.CheckStock(ctx, item.SKUID, quantity)
	if err != nil || !hasStock {
		return nil, ErrStockInsufficient
	}

	// 3. 更新数量
	item.Quantity = quantity
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdateCartItem(ctx, item); err != nil {
		s.logger.Error("更新购物车项失败", zap.Error(err))
		return nil, err
	}

	return item, nil
}

// RemoveFromCart 从购物车移除商品
func (s *cartService) RemoveFromCart(ctx context.Context, userID, itemID uint64) error {
	// 验证购物车项所属用户
	item, err := s.repo.GetCartItemByID(ctx, itemID)
	if err != nil {
		return ErrCartItemNotFound
	}

	if item.UserID != userID {
		return fmt.Errorf("无权操作该购物车项")
	}

	if err := s.repo.DeleteCartItem(ctx, itemID); err != nil {
		s.logger.Error("删除购物车项失败", zap.Error(err))
		return err
	}

	s.logger.Info("从购物车移除成功",
		zap.Uint64("userID", userID),
		zap.Uint64("itemID", itemID))

	return nil
}

// ClearCart 清空购物车
func (s *cartService) ClearCart(ctx context.Context, userID uint64) error {
	if err := s.repo.ClearCart(ctx, userID); err != nil {
		s.logger.Error("清空购物车失败", zap.Error(err), zap.Uint64("userID", userID))
		return err
	}

	s.logger.Info("清空购物车成功", zap.Uint64("userID", userID))
	return nil
}

// GetCart 获取购物车列表
func (s *cartService) GetCart(ctx context.Context, userID uint64) ([]*model.CartItem, error) {
	items, err := s.repo.GetCartByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("获取购物车失败", zap.Error(err), zap.Uint64("userID", userID))
		return nil, err
	}

	// TODO: 可以在这里批量获取最新的商品信息和价格，更新购物车

	return items, nil
}

// GetCartItemCount 获取购物车商品数量
func (s *cartService) GetCartItemCount(ctx context.Context, userID uint64) (int64, error) {
	count, err := s.repo.GetCartItemCount(ctx, userID)
	if err != nil {
		s.logger.Error("获取购物车数量失败", zap.Error(err), zap.Uint64("userID", userID))
		return 0, err
	}
	return count, nil
}

// SelectCartItem 选中/取消选中购物车项
func (s *cartService) SelectCartItem(ctx context.Context, userID, itemID uint64, selected bool) error {
	item, err := s.repo.GetCartItemByID(ctx, itemID)
	if err != nil {
		return ErrCartItemNotFound
	}

	if item.UserID != userID {
		return fmt.Errorf("无权操作该购物车项")
	}

	item.Selected = selected
	item.UpdatedAt = time.Now()

	if err := s.repo.UpdateCartItem(ctx, item); err != nil {
		s.logger.Error("更新购物车项选中状态失败", zap.Error(err))
		return err
	}

	return nil
}

// SelectAllCartItems 全选/取消全选购物车
func (s *cartService) SelectAllCartItems(ctx context.Context, userID uint64, selected bool) error {
	if err := s.repo.SelectAllCartItems(ctx, userID, selected); err != nil {
		s.logger.Error("全选购物车失败", zap.Error(err))
		return err
	}
	return nil
}

// MergeCart 合并购物车（游客购物车合并到登录用户购物车）
func (s *cartService) MergeCart(ctx context.Context, guestUserID, loginUserID uint64) error {
	// 1. 获取游客购物车
	guestItems, err := s.repo.GetCartByUserID(ctx, guestUserID)
	if err != nil {
		return err
	}

	if len(guestItems) == 0 {
		return nil // 游客购物车为空，无需合并
	}

	// 2. 获取登录用户购物车
	loginItems, err := s.repo.GetCartByUserID(ctx, loginUserID)
	if err != nil {
		return err
	}

	// 3. 构建登录用户购物车的SKU映射
	loginItemMap := make(map[uint64]*model.CartItem)
	for _, item := range loginItems {
		loginItemMap[item.SKUID] = item
	}

	// 4. 合并购物车
	for _, guestItem := range guestItems {
		if loginItem, exists := loginItemMap[guestItem.SKUID]; exists {
			// SKU已存在，合并数量
			loginItem.Quantity += guestItem.Quantity
			loginItem.UpdatedAt = time.Now()
			if err := s.repo.UpdateCartItem(ctx, loginItem); err != nil {
				s.logger.Error("合并购物车项失败", zap.Error(err))
				continue
			}
		} else {
			// SKU不存在，添加到登录用户购物车
			guestItem.UserID = loginUserID
			guestItem.UpdatedAt = time.Now()
			if err := s.repo.CreateCartItem(ctx, guestItem); err != nil {
				s.logger.Error("添加购物车项失败", zap.Error(err))
				continue
			}
		}
	}

	// 5. 清空游客购物车
	if err := s.repo.ClearCart(ctx, guestUserID); err != nil {
		s.logger.Error("清空游客购物车失败", zap.Error(err))
	}

	s.logger.Info("合并购物车成功",
		zap.Uint64("guestUserID", guestUserID),
		zap.Uint64("loginUserID", loginUserID))

	return nil
}
