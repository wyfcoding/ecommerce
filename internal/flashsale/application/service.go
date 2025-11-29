package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/idgen"
	"github.com/wyfcoding/ecommerce/pkg/messagequeue/kafka"

	"log/slog"
)

type FlashSaleService struct {
	repo     repository.FlashSaleRepository
	cache    repository.FlashSaleCache
	producer *kafka.Producer
	idGen    idgen.Generator
	logger   *slog.Logger
}

func NewFlashSaleService(
	repo repository.FlashSaleRepository,
	cache repository.FlashSaleCache,
	producer *kafka.Producer,
	idGen idgen.Generator,
	logger *slog.Logger,
) *FlashSaleService {
	return &FlashSaleService{
		repo:     repo,
		cache:    cache,
		producer: producer,
		idGen:    idGen,
		logger:   logger,
	}
}

// CreateFlashsale 创建秒杀活动
func (s *FlashSaleService) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*entity.Flashsale, error) {
	flashsale := entity.NewFlashsale(name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
	if err := s.repo.SaveFlashsale(ctx, flashsale); err != nil {
		s.logger.Error("failed to save flashsale", "error", err)
		return nil, err
	}

	// Pre-warm Cache
	if err := s.cache.SetStock(ctx, uint64(flashsale.ID), totalStock); err != nil {
		s.logger.Error("failed to pre-warm cache", "error", err)
		// Should we fail? Ideally yes, or retry.
		return nil, fmt.Errorf("failed to pre-warm cache: %w", err)
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

// PlaceOrder 下单 (High Concurrency Version)
func (s *FlashSaleService) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*entity.FlashsaleOrder, error) {
	// 1. Get Flashsale (Can be cached too, but let's assume it's fast or cached in repo)
	flashsale, err := s.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	// 2. Validate Time
	now := time.Now()
	if now.Before(flashsale.StartTime) || now.After(flashsale.EndTime) {
		return nil, errors.New("flashsale is not active")
	}

	// 3. Deduct Stock in Redis (Atomic Check & Deduct)
	success, err := s.cache.DeductStock(ctx, flashsaleID, userID, quantity, flashsale.LimitPerUser)
	if err != nil {
		s.logger.Error("failed to deduct stock in cache", "error", err)
		return nil, err
	}
	if !success {
		return nil, entity.ErrFlashsaleSoldOut // Or limit exceeded
	}

	// 4. Generate Order ID
	orderID := s.idGen.Generate()

	// 5. Create Order Object (Pending)
	order := entity.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	order.ID = uint(orderID)
	order.Status = entity.FlashsaleOrderStatusPending

	// 6. Publish to MQ
	event := map[string]interface{}{
		"order_id":     orderID,
		"flashsale_id": flashsaleID,
		"user_id":      userID,
		"product_id":   flashsale.ProductID,
		"sku_id":       flashsale.SkuID,
		"quantity":     quantity,
		"price":        flashsale.FlashPrice,
		"created_at":   order.CreatedAt,
	}
	payload, _ := json.Marshal(event)

	// Use Order ID as key for ordering if needed, or Flashsale ID
	if err := s.producer.Publish(ctx, []byte(fmt.Sprintf("%d", orderID)), payload); err != nil {
		s.logger.Error("failed to publish order event", "error", err)
		// Rollback Redis Stock
		_ = s.cache.RevertStock(ctx, flashsaleID, userID, quantity)
		return nil, fmt.Errorf("failed to publish order: %w", err)
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

	// Restore stock in Redis
	return s.cache.RevertStock(ctx, order.FlashsaleID, order.UserID, order.Quantity)
}
