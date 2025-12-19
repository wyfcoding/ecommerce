package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue/kafka"

	"log/slog"
)

type FlashsaleManager struct {
	repo     domain.FlashSaleRepository
	cache    domain.FlashSaleCache
	producer *kafka.Producer
	idGen    idgen.Generator
	logger   *slog.Logger
}

func NewFlashsaleManager(
	repo domain.FlashSaleRepository,
	cache domain.FlashSaleCache,
	producer *kafka.Producer,
	idGen idgen.Generator,
	logger *slog.Logger,
) *FlashsaleManager {
	return &FlashsaleManager{
		repo:     repo,
		cache:    cache,
		producer: producer,
		idGen:    idGen,
		logger:   logger,
	}
}

// CreateFlashsale 创建一个新的秒杀活动。
func (m *FlashsaleManager) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*domain.Flashsale, error) {
	flashsale := domain.NewFlashsale(name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
	if err := m.repo.SaveFlashsale(ctx, flashsale); err != nil {
		m.logger.ErrorContext(ctx, "failed to save flashsale", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "flashsale created successfully", "flashsale_id", flashsale.ID, "name", name)

	if err := m.cache.SetStock(ctx, uint64(flashsale.ID), totalStock); err != nil {
		m.logger.ErrorContext(ctx, "failed to pre-warm cache", "flashsale_id", flashsale.ID, "error", err)
		return nil, fmt.Errorf("failed to pre-warm cache: %w", err)
	}

	return flashsale, nil
}

// PlaceOrder 下达一个秒杀订单。
func (m *FlashsaleManager) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*domain.FlashsaleOrder, error) {
	flashsale, err := m.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if now.Before(flashsale.StartTime) || now.After(flashsale.EndTime) {
		return nil, errors.New("flashsale is not active")
	}

	success, err := m.cache.DeductStock(ctx, flashsaleID, userID, quantity, flashsale.LimitPerUser)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to deduct stock in cache", "flashsale_id", flashsaleID, "user_id", userID, "error", err)
		return nil, err
	}
	if !success {
		return nil, domain.ErrFlashsaleSoldOut
	}

	orderID := m.idGen.Generate()
	order := domain.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	order.ID = uint(orderID)
	order.Status = domain.FlashsaleOrderStatusPending

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

	if err := m.producer.Publish(ctx, []byte(fmt.Sprintf("%d", orderID)), payload); err != nil {
		m.logger.ErrorContext(ctx, "failed to publish order event", "order_id", orderID, "error", err)
		_ = m.cache.RevertStock(ctx, flashsaleID, userID, quantity)
		return nil, fmt.Errorf("failed to publish order: %w", err)
	}

	return order, nil
}

// CancelOrder 取消一个秒杀订单。
func (m *FlashsaleManager) CancelOrder(ctx context.Context, orderID uint64) error {
	order, err := m.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	if order.Status != domain.FlashsaleOrderStatusPending {
		return nil
	}

	order.Cancel()
	if err := m.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	return m.cache.RevertStock(ctx, order.FlashsaleID, order.UserID, order.Quantity)
}

func (m *FlashsaleManager) SaveOrder(ctx context.Context, order *domain.FlashsaleOrder) error {
	return m.repo.SaveOrder(ctx, order)
}

func (m *FlashsaleManager) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return m.repo.UpdateStock(ctx, id, quantity)
}
