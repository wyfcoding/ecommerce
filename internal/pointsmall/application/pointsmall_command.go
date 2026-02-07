package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// PointsmallCommandService 处理积分商城的写操作。
type PointsmallCommandService struct {
	repo      domain.PointsRepository
	publisher domain.EventPublisher
	idGen     idgen.Generator
	logger    *slog.Logger
}

// NewPointsmallCommandService creates a new PointsmallCommandService instance.
func NewPointsmallCommandService(repo domain.PointsRepository, publisher domain.EventPublisher, idGen idgen.Generator, logger *slog.Logger) *PointsmallCommandService {
	return &PointsmallCommandService{
		repo:      repo,
		publisher: publisher,
		idGen:     idGen,
		logger:    logger,
	}
}

// CreateProduct 创建积分商品。
func (m *PointsmallCommandService) CreateProduct(ctx context.Context, product *domain.PointsProduct) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveProductInTx(ctx, tx, product); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PointsProductCreatedEventType, fmt.Sprintf("%d", product.ID), &domain.PointsProductCreatedEvent{
			ProductID: uint64(product.ID),
			Status:    product.Status,
			Stock:     product.Stock,
			Timestamp: time.Now(),
		})
	})
}

// ExchangeProduct 兑换商品。
func (m *PointsmallCommandService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*domain.PointsOrder, error) {
	// 1. 获取商品信息
	product, err := m.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.Status != domain.PointsProductStatusOnline {
		return nil, errors.New("product not online")
	}
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// 2. 获取用户积分账户
	account, err := m.repo.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account = &domain.PointsAccount{UserID: userID}
	}

	// 3. 检查积分
	totalPoints := product.Points * int64(quantity)
	if account.TotalPoints-account.UsedPoints < totalPoints {
		return nil, errors.New("insufficient points")
	}

	// 4. 事务内更新积分、库存与订单
	orderID := m.idGen.Generate()
	orderNo := fmt.Sprintf("P%s%d", time.Now().Format("20060102"), orderID)
	order := &domain.PointsOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		ProductName: product.Name,
		Quantity:    quantity,
		Points:      product.Points,
		TotalPoints: totalPoints,
		Status:      domain.PointsOrderStatusPending,
		Address:     address,
		Phone:       phone,
		Receiver:    receiver,
	}

	err = m.repo.WithTx(ctx, func(tx any) error {
		// 扣减积分
		account.UsedPoints += totalPoints
		if err := m.repo.SaveAccountInTx(ctx, tx, account); err != nil {
			return err
		}

		// 扣减库存
		product.Stock -= quantity
		product.SoldCount += quantity
		if err := m.repo.SaveProductInTx(ctx, tx, product); err != nil {
			return err
		}

		// 创建订单
		if err := m.repo.SaveOrderInTx(ctx, tx, order); err != nil {
			return err
		}

		// 记录流水
		txRecord := &domain.PointsTransaction{
			UserID:      userID,
			Type:        "spend",
			Points:      -totalPoints,
			Description: fmt.Sprintf("Exchange product: %s", product.Name),
			RefID:       orderNo,
		}
		if err := m.repo.SaveTransactionInTx(ctx, tx, txRecord); err != nil {
			return err
		}

		// 发布事件
		if m.publisher == nil {
			return nil
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.PointsAccountUpdatedEventType, fmt.Sprintf("%d", account.UserID), &domain.PointsAccountUpdatedEvent{
			UserID:      account.UserID,
			TotalPoints: account.TotalPoints,
			UsedPoints:  account.UsedPoints,
			Timestamp:   time.Now(),
		}); err != nil {
			return err
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.PointsStockUpdatedEventType, fmt.Sprintf("%d", product.ID), &domain.PointsStockUpdatedEvent{
			ItemID:    uint64(product.ID),
			NewStock:  product.Stock,
			Timestamp: time.Now(),
		}); err != nil {
			return err
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.PointsOrderCreatedEventType, fmt.Sprintf("%d", order.ID), &domain.PointsOrderCreatedEvent{
			OrderID:     uint64(order.ID),
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			ProductID:   order.ProductID,
			Quantity:    order.Quantity,
			TotalPoints: order.TotalPoints,
			Timestamp:   time.Now(),
		}); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PointItemExchangedEventType, fmt.Sprintf("%d", order.ID), &domain.PointItemExchangedEvent{
			ExchangeID: uint64(order.ID),
			UserID:     userID,
			ItemID:     productID,
			Points:     int32(totalPoints),
			Timestamp:  time.Now(),
		})
	})
	if err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "product exchanged successfully", "user_id", userID, "product_id", productID, "order_no", orderNo)
	return order, nil
}

// AddPoints 增加用户积分。
func (m *PointsmallCommandService) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	account, err := m.repo.GetAccount(ctx, userID)
	if err != nil {
		return err
	}
	// 注意：如果 repo 实现中不存在，GetAccount 会自动创建

	if account == nil {
		account = &domain.PointsAccount{UserID: userID}
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		account.TotalPoints += points
		if err := m.repo.SaveAccountInTx(ctx, tx, account); err != nil {
			return err
		}

		txRecord := &domain.PointsTransaction{
			UserID:      userID,
			Type:        "earn",
			Points:      points,
			Description: description,
			RefID:       refID,
		}
		if err := m.repo.SaveTransactionInTx(ctx, tx, txRecord); err != nil {
			return err
		}

		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PointsAccountUpdatedEventType, fmt.Sprintf("%d", account.UserID), &domain.PointsAccountUpdatedEvent{
			UserID:      account.UserID,
			TotalPoints: account.TotalPoints,
			UsedPoints:  account.UsedPoints,
			Timestamp:   time.Now(),
		})
	})
}
