package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository"
	"time"

	"log/slog"
)

type FlashSaleService struct {
	repo   repository.FlashSaleRepository
	logger *slog.Logger
}

func NewFlashSaleService(repo repository.FlashSaleRepository, logger *slog.Logger) *FlashSaleService {
	return &FlashSaleService{
		repo:   repo,
		logger: logger,
	}
}

// CreateFlashsale 创建秒杀活动
func (s *FlashSaleService) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*entity.Flashsale, error) {
	flashsale := entity.NewFlashsale(name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
	if err := s.repo.SaveFlashsale(ctx, flashsale); err != nil {
		s.logger.Error("failed to save flashsale", "error", err)
		return nil, err
	}
	return flashsale, nil
}

// GetFlashsale 获取秒杀活动
func (s *FlashSaleService) GetFlashsale(ctx context.Context, id uint64) (*entity.Flashsale, error) {
	return s.repo.GetFlashsale(ctx, id)
}

// ListFlashsales 获取秒杀活动列表
func (s *FlashSaleService) ListFlashsales(ctx context.Context, status *entity.FlashsaleStatus, page, pageSize int) ([]*entity.Flashsale, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListFlashsales(ctx, status, offset, pageSize)
}

// PlaceOrder 下单
func (s *FlashSaleService) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*entity.FlashsaleOrder, error) {
	// 1. Get Flashsale
	flashsale, err := s.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	// 2. Check User Limit
	boughtCount, err := s.repo.CountUserBought(ctx, userID, flashsaleID)
	if err != nil {
		return nil, err
	}

	// 3. Validate
	if err := flashsale.CanBuy(boughtCount, quantity); err != nil {
		return nil, err
	}

	// 4. Deduct Stock (Optimistic Lock handled in repo)
	if err := s.repo.UpdateStock(ctx, flashsaleID, quantity); err != nil {
		return nil, entity.ErrFlashsaleSoldOut
	}

	// 5. Create Order
	order := entity.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		s.logger.Error("failed to save order", "error", err)
		// Rollback stock? In a real system, we might need a transaction or compensation.
		// For simplicity here, we assume repo operations are atomic enough or we accept slight inconsistency.
		return nil, err
	}

	return order, nil
}

// CancelOrder 取消订单
func (s *FlashSaleService) CancelOrder(ctx context.Context, orderID uint64) error {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if order.Status != entity.FlashsaleOrderStatusPending {
		return nil // Already processed
	}

	order.Cancel()
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	// Restore stock
	// Note: This should be a negative update
	return s.repo.UpdateStock(ctx, order.FlashsaleID, -order.Quantity)
}
