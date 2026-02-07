package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	risksecurityv1 "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// FlashSaleCommandService 负责处理 Flashsale 相关的写操作和业务逻辑。
type FlashSaleCommandService struct {
	repo       domain.FlashSaleRepository
	cache      domain.FlashSaleCache
	publisher  domain.EventPublisher
	idGen      idgen.Generator
	logger     *slog.Logger
	riskClient risksecurityv1.RiskSecurityServiceClient
}

// NewFlashSaleCommandService 构造函数。
func NewFlashSaleCommandService(
	repo domain.FlashSaleRepository,
	cache domain.FlashSaleCache,
	publisher domain.EventPublisher,
	idGen idgen.Generator,
	logger *slog.Logger,
	riskClient risksecurityv1.RiskSecurityServiceClient,
) *FlashSaleCommandService {
	return &FlashSaleCommandService{
		repo:       repo,
		cache:      cache,
		publisher:  publisher,
		idGen:      idGen,
		logger:     logger,
		riskClient: riskClient,
	}
}

// CreateFlashsale 创建一个新的秒杀活动。
func (m *FlashSaleCommandService) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*domain.Flashsale, error) {
	flashsale := domain.NewFlashsale(name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveFlashsaleInTx(ctx, tx, flashsale); err != nil {
			m.logger.ErrorContext(ctx, "failed to save flashsale", "name", name, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.FlashSaleEventCreatedEvent{
			EventID:   flashsale.ID,
			Name:      name,
			StartTime: startTime,
			EndTime:   endTime,
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.FlashsaleCreatedEventType, fmt.Sprintf("%d", flashsale.ID), event)
	}); err != nil {
		return nil, err
	}

	if m.cache != nil {
		if err := m.cache.SetStock(ctx, uint64(flashsale.ID), totalStock); err != nil {
			m.logger.WarnContext(ctx, "failed to pre-warm cache", "flashsale_id", flashsale.ID, "error", err)
		}
	}

	m.logger.InfoContext(ctx, "flashsale created successfully", "flashsale_id", flashsale.ID, "name", name)
	return flashsale, nil
}

// PlaceOrder 下达一个秒杀订单。
func (m *FlashSaleCommandService) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*domain.FlashsaleOrder, error) {
	flashsale, err := m.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}
	if flashsale == nil {
		return nil, domain.ErrFlashsaleNotFound
	}

	now := time.Now()
	if now.Before(flashsale.StartTime) {
		return nil, domain.ErrFlashsaleNotStarted
	}
	if now.After(flashsale.EndTime) {
		return nil, domain.ErrFlashsaleEnded
	}

	// 2. 风控检查
	if m.riskClient != nil {
		riskResp, err := m.riskClient.EvaluateRisk(ctx, &risksecurityv1.EvaluateRiskRequest{
			UserId: userID,
			Amount: int64(flashsale.FlashPrice) * int64(quantity),
		})
		if err == nil && riskResp.Result != nil && riskResp.Result.RiskLevel > 2 {
			m.logger.WarnContext(ctx, "risk check rejected", "user_id", userID, "risk_level", riskResp.Result.RiskLevel)
			return nil, errors.New("risk check rejected")
		}
	}

	// 3. Redis 预扣减 (高性能屏障)
	if m.cache == nil {
		return nil, errors.New("flashsale cache not initialized")
	}
	success, err := m.cache.DeductStock(ctx, flashsaleID, userID, quantity, flashsale.LimitPerUser)
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, domain.ErrFlashsaleSoldOut
	}

	// 4. 本地事务落库 + Outbox
	orderID := uint64(m.idGen.Generate())
	order := domain.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	order.ID = uint(orderID)

	err = m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveOrderInTx(ctx, tx, order); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.FlashSaleOrderCreatedEvent{
			OrderID:     orderID,
			FlashsaleID: flashsaleID,
			UserID:      userID,
			ProductID:   flashsale.ProductID,
			SkuID:       flashsale.SkuID,
			Quantity:    quantity,
			Price:       flashsale.FlashPrice,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.FlashsaleOrderCreatedEventType, fmt.Sprintf("%d", orderID), event)
	})
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to commit flashsale transaction", "order_id", orderID, "error", err)
		// 容错：DB 失败必须回滚 Redis
		if revertErr := m.cache.RevertStock(ctx, flashsaleID, userID, quantity); revertErr != nil {
			m.logger.ErrorContext(ctx, "CRITICAL: failed to revert redis stock after DB failure", "flashsale_id", flashsaleID, "error", revertErr)
		}
		return nil, fmt.Errorf("system error during order creation: %w", err)
	}

	m.logger.InfoContext(ctx, "flashsale order accepted", "order_id", orderID, "user_id", userID)
	return order, nil
}

// CancelOrder 撤销一个秒杀订单（支持库存回滚）。
func (m *FlashSaleCommandService) CancelOrder(ctx context.Context, orderID uint64) error {
	order, err := m.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return nil
	}
	if order.Status != domain.FlashsaleOrderStatusPending {
		return nil
	}

	err = m.repo.WithTx(ctx, func(tx any) error {
		order.Cancel()
		if err := m.repo.SaveOrderInTx(ctx, tx, order); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.FlashSaleOrderCancelledEvent{
			OrderID:     uint64(orderID),
			FlashsaleID: order.FlashsaleID,
			UserID:      order.UserID,
			Quantity:    order.Quantity,
			Timestamp:   time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.FlashsaleOrderCancelledEventType, fmt.Sprintf("%d", orderID), event)
	})
	if err != nil {
		return err
	}

	if m.cache != nil {
		_ = m.cache.RevertStock(ctx, order.FlashsaleID, order.UserID, order.Quantity)
	}
	return nil
}

func (m *FlashSaleCommandService) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return m.repo.UpdateStock(ctx, id, quantity)
}
