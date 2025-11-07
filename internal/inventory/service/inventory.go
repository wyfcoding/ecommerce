package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ecommerce/internal/inventory/model"
	"ecommerce/internal/inventory/repository"
)

var (
	ErrInsufficientStock = errors.New("库存不足")
	ErrStockNotFound     = errors.New("库存记录不存在")
	ErrInvalidQuantity   = errors.New("无效的数量")
)

// InventoryService 库存服务接口
type InventoryService interface {
	// 库存查询
	GetStock(ctx context.Context, skuID uint64) (*model.Stock, error)
	BatchGetStock(ctx context.Context, skuIDs []uint64) (map[uint64]*model.Stock, error)
	
	// 库存操作
	LockStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error
	UnlockStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error
	DeductStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error
	RestoreStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error
	
	// 库存调整
	AdjustStock(ctx context.Context, skuID uint64, quantity int32, reason string, operatorID uint64) error
	
	// 库存预警
	GetLowStockProducts(ctx context.Context, threshold uint32) ([]*model.Stock, error)
	
	// 库存日志
	GetStockLogs(ctx context.Context, skuID uint64, pageSize, pageNum int32) ([]*model.StockLog, int64, error)
}

type inventoryService struct {
	repo   repository.InventoryRepo
	logger *zap.Logger
}

// NewInventoryService 创建库存服务实例
func NewInventoryService(repo repository.InventoryRepo, logger *zap.Logger) InventoryService {
	return &inventoryService{
		repo:   repo,
		logger: logger,
	}
}

// GetStock 获取SKU库存信息
func (s *inventoryService) GetStock(ctx context.Context, skuID uint64) (*model.Stock, error) {
	stock, err := s.repo.GetStockBySKUID(ctx, skuID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStockNotFound
		}
		s.logger.Error("获取库存失败", zap.Error(err), zap.Uint64("skuID", skuID))
		return nil, err
	}
	return stock, nil
}

// BatchGetStock 批量获取库存信息
func (s *inventoryService) BatchGetStock(ctx context.Context, skuIDs []uint64) (map[uint64]*model.Stock, error) {
	stocks, err := s.repo.BatchGetStock(ctx, skuIDs)
	if err != nil {
		s.logger.Error("批量获取库存失败", zap.Error(err))
		return nil, err
	}
	
	result := make(map[uint64]*model.Stock)
	for _, stock := range stocks {
		result[stock.SKUID] = stock
	}
	return result, nil
}

// LockStock 锁定库存（下单时）
func (s *inventoryService) LockStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error {
	if quantity == 0 {
		return ErrInvalidQuantity
	}

	// 使用事务确保原子性
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		// 获取当前库存
		stock, err := s.repo.GetStockBySKUID(ctx, skuID)
		if err != nil {
			return err
		}

		// 检查可用库存
		if stock.AvailableStock < quantity {
			return ErrInsufficientStock
		}

		// 更新库存
		stock.AvailableStock -= quantity
		stock.LockedStock += quantity
		stock.UpdatedAt = time.Now()

		if err := s.repo.UpdateStock(ctx, stock); err != nil {
			return err
		}

		// 记录库存日志
		log := &model.StockLog{
			SKUID:      skuID,
			Type:       model.StockLogTypeLock,
			Quantity:   int32(quantity),
			BeforeQty:  stock.AvailableStock + quantity,
			AfterQty:   stock.AvailableStock,
			OrderID:    orderID,
			OperatorID: 0, // 系统操作
			Reason:     "订单锁定库存",
			CreatedAt:  time.Now(),
		}
		return s.repo.CreateStockLog(ctx, log)
	})

	if err != nil {
		s.logger.Error("锁定库存失败",
			zap.Error(err),
			zap.Uint64("skuID", skuID),
			zap.Uint32("quantity", quantity),
			zap.String("orderID", orderID))
		return err
	}

	s.logger.Info("锁定库存成功",
		zap.Uint64("skuID", skuID),
		zap.Uint32("quantity", quantity),
		zap.String("orderID", orderID))
	return nil
}

// UnlockStock 解锁库存（取消订单时）
func (s *inventoryService) UnlockStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error {
	if quantity == 0 {
		return ErrInvalidQuantity
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		stock, err := s.repo.GetStockBySKUID(ctx, skuID)
		if err != nil {
			return err
		}

		if stock.LockedStock < quantity {
			return fmt.Errorf("锁定库存不足: 需要 %d, 实际 %d", quantity, stock.LockedStock)
		}

		stock.AvailableStock += quantity
		stock.LockedStock -= quantity
		stock.UpdatedAt = time.Now()

		if err := s.repo.UpdateStock(ctx, stock); err != nil {
			return err
		}

		log := &model.StockLog{
			SKUID:      skuID,
			Type:       model.StockLogTypeUnlock,
			Quantity:   int32(quantity),
			BeforeQty:  stock.AvailableStock - quantity,
			AfterQty:   stock.AvailableStock,
			OrderID:    orderID,
			OperatorID: 0,
			Reason:     "订单取消，解锁库存",
			CreatedAt:  time.Now(),
		}
		return s.repo.CreateStockLog(ctx, log)
	})

	if err != nil {
		s.logger.Error("解锁库存失败",
			zap.Error(err),
			zap.Uint64("skuID", skuID),
			zap.Uint32("quantity", quantity))
		return err
	}

	s.logger.Info("解锁库存成功",
		zap.Uint64("skuID", skuID),
		zap.Uint32("quantity", quantity))
	return nil
}

// DeductStock 扣减库存（支付成功后）
func (s *inventoryService) DeductStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error {
	if quantity == 0 {
		return ErrInvalidQuantity
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		stock, err := s.repo.GetStockBySKUID(ctx, skuID)
		if err != nil {
			return err
		}

		if stock.LockedStock < quantity {
			return fmt.Errorf("锁定库存不足")
		}

		stock.LockedStock -= quantity
		stock.TotalStock -= quantity
		stock.UpdatedAt = time.Now()

		if err := s.repo.UpdateStock(ctx, stock); err != nil {
			return err
		}

		log := &model.StockLog{
			SKUID:      skuID,
			Type:       model.StockLogTypeDeduct,
			Quantity:   -int32(quantity),
			BeforeQty:  stock.TotalStock + quantity,
			AfterQty:   stock.TotalStock,
			OrderID:    orderID,
			OperatorID: 0,
			Reason:     "订单支付成功，扣减库存",
			CreatedAt:  time.Now(),
		}
		return s.repo.CreateStockLog(ctx, log)
	})

	if err != nil {
		s.logger.Error("扣减库存失败",
			zap.Error(err),
			zap.Uint64("skuID", skuID),
			zap.Uint32("quantity", quantity))
		return err
	}

	s.logger.Info("扣减库存成功",
		zap.Uint64("skuID", skuID),
		zap.Uint32("quantity", quantity))
	return nil
}

// RestoreStock 恢复库存（退款时）
func (s *inventoryService) RestoreStock(ctx context.Context, skuID uint64, quantity uint32, orderID string) error {
	if quantity == 0 {
		return ErrInvalidQuantity
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		stock, err := s.repo.GetStockBySKUID(ctx, skuID)
		if err != nil {
			return err
		}

		stock.TotalStock += quantity
		stock.AvailableStock += quantity
		stock.UpdatedAt = time.Now()

		if err := s.repo.UpdateStock(ctx, stock); err != nil {
			return err
		}

		log := &model.StockLog{
			SKUID:      skuID,
			Type:       model.StockLogTypeRestore,
			Quantity:   int32(quantity),
			BeforeQty:  stock.TotalStock - quantity,
			AfterQty:   stock.TotalStock,
			OrderID:    orderID,
			OperatorID: 0,
			Reason:     "订单退款，恢复库存",
			CreatedAt:  time.Now(),
		}
		return s.repo.CreateStockLog(ctx, log)
	})

	if err != nil {
		s.logger.Error("恢复库存失败",
			zap.Error(err),
			zap.Uint64("skuID", skuID),
			zap.Uint32("quantity", quantity))
		return err
	}

	s.logger.Info("恢复库存成功",
		zap.Uint64("skuID", skuID),
		zap.Uint32("quantity", quantity))
	return nil
}

// AdjustStock 调整库存（人工操作）
func (s *inventoryService) AdjustStock(ctx context.Context, skuID uint64, quantity int32, reason string, operatorID uint64) error {
	if quantity == 0 {
		return ErrInvalidQuantity
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		stock, err := s.repo.GetStockBySKUID(ctx, skuID)
		if err != nil {
			return err
		}

		beforeQty := stock.TotalStock
		
		// 调整总库存和可用库存
		if quantity > 0 {
			stock.TotalStock += uint32(quantity)
			stock.AvailableStock += uint32(quantity)
		} else {
			absQty := uint32(-quantity)
			if stock.AvailableStock < absQty {
				return ErrInsufficientStock
			}
			stock.TotalStock -= absQty
			stock.AvailableStock -= absQty
		}
		stock.UpdatedAt = time.Now()

		if err := s.repo.UpdateStock(ctx, stock); err != nil {
			return err
		}

		log := &model.StockLog{
			SKUID:      skuID,
			Type:       model.StockLogTypeAdjust,
			Quantity:   quantity,
			BeforeQty:  beforeQty,
			AfterQty:   stock.TotalStock,
			OperatorID: operatorID,
			Reason:     reason,
			CreatedAt:  time.Now(),
		}
		return s.repo.CreateStockLog(ctx, log)
	})

	if err != nil {
		s.logger.Error("调整库存失败",
			zap.Error(err),
			zap.Uint64("skuID", skuID),
			zap.Int32("quantity", quantity))
		return err
	}

	s.logger.Info("调整库存成功",
		zap.Uint64("skuID", skuID),
		zap.Int32("quantity", quantity),
		zap.Uint64("operatorID", operatorID))
	return nil
}

// GetLowStockProducts 获取低库存商品
func (s *inventoryService) GetLowStockProducts(ctx context.Context, threshold uint32) ([]*model.Stock, error) {
	stocks, err := s.repo.GetLowStockProducts(ctx, threshold)
	if err != nil {
		s.logger.Error("获取低库存商品失败", zap.Error(err))
		return nil, err
	}
	return stocks, nil
}

// GetStockLogs 获取库存日志
func (s *inventoryService) GetStockLogs(ctx context.Context, skuID uint64, pageSize, pageNum int32) ([]*model.StockLog, int64, error) {
	logs, total, err := s.repo.GetStockLogs(ctx, skuID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取库存日志失败", zap.Error(err), zap.Uint64("skuID", skuID))
		return nil, 0, err
	}
	return logs, total, nil
}
