package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// WarehouseCommandService 处理所有仓库相关的写入操作。
type WarehouseCommandService struct {
	repo      domain.WarehouseRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

// NewWarehouseCommandService 构造函数。
func NewWarehouseCommandService(
	repo domain.WarehouseRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *WarehouseCommandService {
	return &WarehouseCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// DeductStock 扣减库存 (用于 Saga)。
func (s *WarehouseCommandService) DeductStock(ctx context.Context, barrier any, orderID, skuID, warehouseID uint64, quantity int32) error {
	return s.repo.WithBarrier(ctx, barrier, func(tx any) error {
		// 1. 获取库存并加行锁
		stock, err := s.repo.GetStockWithLock(ctx, tx, warehouseID, skuID)
		if err != nil {
			return err
		}
		if stock == nil {
			return fmt.Errorf("stock not found for SKU %d in warehouse %d", skuID, warehouseID)
		}

		// 2. 检查可用库存
		if stock.AvailableStock() < quantity {
			return fmt.Errorf("insufficient stock for SKU %d: available %d, req %d", skuID, stock.AvailableStock(), quantity)
		}

		// 3. 执行扣减
		stock.Stock -= quantity
		if err := s.repo.UpdateStockInTx(ctx, tx, stock); err != nil {
			return err
		}

		// 4. 发布事件 (Outbox)
		event := &domain.StockDeductedEvent{
			OrderID:     orderID,
			SkuID:       skuID,
			Quantity:    quantity,
			WarehouseID: warehouseID,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.StockDeductedEventType, fmt.Sprintf("%d-%d", orderID, skuID), event)
	})
}

// RevertStock 回滚库存 (用于 Saga 补偿)。
func (s *WarehouseCommandService) RevertStock(ctx context.Context, barrier any, orderID, skuID, warehouseID uint64, quantity int32) error {
	return s.repo.WithBarrier(ctx, barrier, func(tx any) error {
		stock, err := s.repo.GetStockWithLock(ctx, tx, warehouseID, skuID)
		if err != nil {
			return err
		}
		if stock == nil {
			s.logger.WarnContext(ctx, "stock not found during revert", "order_id", orderID, "sku_id", skuID)
			return nil
		}

		stock.Stock += quantity
		if err := s.repo.UpdateStockInTx(ctx, tx, stock); err != nil {
			return err
		}

		event := &domain.StockRevertedEvent{
			OrderID:     orderID,
			SkuID:       skuID,
			Quantity:    quantity,
			WarehouseID: warehouseID,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.StockRevertedEventType, fmt.Sprintf("%d-%d", orderID, skuID), event)
	})
}

// CreateWarehouse 创建。
func (s *WarehouseCommandService) CreateWarehouse(ctx context.Context, w *domain.Warehouse) (*domain.Warehouse, error) {
	if err := s.repo.Save(ctx, w); err != nil {
		return nil, err
	}
	err := s.publisher.Publish(ctx, domain.WarehouseCreatedEventType, w.Code, &domain.WarehouseCreatedEvent{
		WarehouseID: uint64(w.ID),
		Code:        w.Code,
		Name:        w.Name,
		Timestamp:   time.Now(),
	})
	return w, err
}

// AdjustStock 人工调整。
func (s *WarehouseCommandService) AdjustStock(ctx context.Context, warehouseID, skuID uint64, newQty int32, reason string) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		stock, err := s.repo.GetStockWithLock(ctx, tx, warehouseID, skuID)
		if err != nil {
			return err
		}
		oldQty := int32(0)
		if stock == nil {
			stock = &domain.WarehouseStock{
				WarehouseID: warehouseID,
				SkuID:       skuID,
				Stock:       newQty,
			}
			if err := s.repo.SaveStockInTx(ctx, tx, stock); err != nil {
				return err
			}
		} else {
			oldQty = stock.Stock
			stock.Stock = newQty
			if err := s.repo.UpdateStockInTx(ctx, tx, stock); err != nil {
				return err
			}
		}

		return s.publisher.PublishInTx(ctx, tx, domain.StockAdjustedEventType, fmt.Sprintf("%d-%d", warehouseID, skuID), &domain.StockAdjustedEvent{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			OldQty:      oldQty,
			NewQty:      newQty,
			Reason:      reason,
			Timestamp:   time.Now(),
		})
	})
}

// CreateTransfer 创建调拨单。
func (s *WarehouseCommandService) CreateTransfer(ctx context.Context, fromID, toID, skuID uint64, qty int32, createdBy uint64) (*domain.StockTransfer, error) {
	transfer := &domain.StockTransfer{
		TransferNo:      fmt.Sprintf("TR%d%d", time.Now().Unix(), skuID),
		FromWarehouseID: fromID,
		ToWarehouseID:   toID,
		SkuID:           skuID,
		Quantity:        qty,
		Status:          domain.StockTransferStatusPending,
		CreatedBy:       createdBy,
	}
	if err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveTransferInTx(ctx, tx, transfer); err != nil {
			return err
		}
		event := &domain.StockTransferCreatedEvent{
			TransferID:      uint64(transfer.ID),
			TransferNo:      transfer.TransferNo,
			FromWarehouseID: transfer.FromWarehouseID,
			ToWarehouseID:   transfer.ToWarehouseID,
			SkuID:           transfer.SkuID,
			Quantity:        transfer.Quantity,
			Status:          transfer.Status,
			CreatedBy:       transfer.CreatedBy,
			Timestamp:       time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.StockTransferCreatedEventType, fmt.Sprintf("%d", transfer.ID), event)
	}); err != nil {
		return nil, err
	}
	return transfer, nil
}

// CompleteTransfer 完成调拨。
func (s *WarehouseCommandService) CompleteTransfer(ctx context.Context, id uint64) error {
	return s.repo.WithTx(ctx, func(tx any) error {
		transfer, err := s.repo.FindTransferByID(ctx, id)
		if err != nil {
			return err
		}
		if transfer == nil {
			return fmt.Errorf("transfer not found")
		}
		transfer.Status = domain.StockTransferStatusCompleted
		now := time.Now()
		transfer.CompletedAt = &now
		if err := s.repo.UpdateTransferInTx(ctx, tx, transfer); err != nil {
			return err
		}
		event := &domain.StockTransferCompletedEvent{
			TransferID:  uint64(transfer.ID),
			TransferNo:  transfer.TransferNo,
			Status:      transfer.Status,
			CompletedAt: now,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.StockTransferCompletedEventType, fmt.Sprintf("%d", transfer.ID), event)
	})
}
