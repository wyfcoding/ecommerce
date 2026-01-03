package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	risksecurityv1 "github.com/wyfcoding/ecommerce/goapi/risksecurity/v1"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type cachedFlashsale struct {
	Data      *domain.Flashsale
	ExpiresAt time.Time
}

// FlashsaleManager 负责处理 Flashsale 相关的写操作和业务逻辑。
type FlashsaleManager struct {
	repo       domain.FlashSaleRepository
	cache      domain.FlashSaleCache
	producer   *kafka.Producer
	outbox     *outbox.Manager
	db         *gorm.DB
	idGen      idgen.Generator
	logger     *slog.Logger
	riskClient risksecurityv1.RiskSecurityServiceClient
	localCache sync.Map
}

// NewFlashsaleManager 负责处理 NewFlashsale 相关的写操作和业务逻辑。
func NewFlashsaleManager(
	repo domain.FlashSaleRepository,
	cache domain.FlashSaleCache,
	producer *kafka.Producer,
	outboxMgr *outbox.Manager,
	db *gorm.DB,
	idGen idgen.Generator,
	logger *slog.Logger,
	riskClient risksecurityv1.RiskSecurityServiceClient,
) *FlashsaleManager {
	return &FlashsaleManager{
		repo:       repo,
		cache:      cache,
		producer:   producer,
		outbox:     outboxMgr,
		db:         db,
		idGen:      idGen,
		logger:     logger,
		riskClient: riskClient,
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

	// Update local cache
	m.localCache.Store(uint64(flashsale.ID), cachedFlashsale{
		Data:      flashsale,
		ExpiresAt: time.Now().Add(1 * time.Minute),
	})

	return flashsale, nil
}

// getFlashsaleWithCache retrieves flashsale from local cache or DB.
func (m *FlashsaleManager) getFlashsaleWithCache(ctx context.Context, flashsaleID uint64) (*domain.Flashsale, error) {
	if val, ok := m.localCache.Load(flashsaleID); ok {
		cached := val.(cachedFlashsale)
		if time.Now().Before(cached.ExpiresAt) {
			return cached.Data, nil
		}
	}

	// Cache miss or expired
	flashsale, err := m.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	m.localCache.Store(flashsaleID, cachedFlashsale{
		Data:      flashsale,
		ExpiresAt: time.Now().Add(5 * time.Second), // Short TTL for high concurrency
	})

	return flashsale, nil
}

// PlaceOrder 下达一个秒杀订单。
func (m *FlashsaleManager) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*domain.FlashsaleOrder, error) {
	// 1. 获取活动信息 (Local Cache)
	flashsale, err := m.getFlashsaleWithCache(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if now.Before(flashsale.StartTime) || now.After(flashsale.EndTime) {
		return nil, errors.New("flashsale is not active")
	}

	// 2. 风控检查 (Risk Control)
	riskResp, err := m.riskClient.EvaluateRisk(ctx, &risksecurityv1.EvaluateRiskRequest{
		UserId: userID,
		Amount: int64(flashsale.FlashPrice) * int64(quantity),
	})
	if err != nil {
		// 风控调用失败，为了安全起见，可以选择拒绝或降级 (此处选择 Log & Continue 或 Reject)
		m.logger.WarnContext(ctx, "risk evaluation failed", "error", err)
	} else if riskResp.Result != nil && riskResp.Result.RiskLevel > 2 {
		// 假设 RiskLevel > 2 代表 High 或 Critical
		m.logger.WarnContext(ctx, "risk check rejected", "user_id", userID, "risk_level", riskResp.Result.RiskLevel)
		return nil, errors.New("risk check rejected")
	}

	// 3. 核心步骤 A: Redis 预扣减 (高性能屏障)
	success, err := m.cache.DeductStock(ctx, flashsaleID, userID, quantity, flashsale.LimitPerUser)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to deduct stock in cache", "flashsale_id", flashsaleID, "user_id", userID, "error", err)
		return nil, err
	}
	if !success {
		return nil, domain.ErrFlashsaleSoldOut
	}

	// 4. 核心步骤 B: 本地事务落库 + Outbox (可靠性保证)
	orderID := m.idGen.Generate()
	order := domain.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	order.ID = uint(orderID)
	order.Status = domain.FlashsaleOrderStatusPending

	err = m.db.Transaction(func(tx *gorm.DB) error {
		// 4.1 保存订单占位符 (状态为 PENDING)
		if err := m.repo.SaveOrder(ctx, order); err != nil {
			return err
		}

		// 4.2 写入 Outbox 事件
		event := map[string]any{
			"order_id":     orderID,
			"flashsale_id": flashsaleID,
			"user_id":      userID,
			"product_id":   flashsale.ProductID,
			"sku_id":       flashsale.SkuID,
			"quantity":     quantity,
			"price":        flashsale.FlashPrice,
			"created_at":   order.CreatedAt,
		}

		return m.outbox.PublishInTx(tx, "flashsale.order.created", fmt.Sprintf("%d", orderID), event)
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
