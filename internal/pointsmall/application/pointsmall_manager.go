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

// PointsManager 处理积分商城的写操作。
type PointsManager struct {
	repo   domain.PointsRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

// NewPointsManager creates a new PointsManager instance.
func NewPointsManager(repo domain.PointsRepository, idGen idgen.Generator, logger *slog.Logger) *PointsManager {
	return &PointsManager{
		repo:   repo,
		idGen:  idGen,
		logger: logger,
	}
}

// CreateProduct 创建积分商品。
func (m *PointsManager) CreateProduct(ctx context.Context, product *domain.PointsProduct) error {
	return m.repo.SaveProduct(ctx, product)
}

// ExchangeProduct 兑换商品。
func (m *PointsManager) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*domain.PointsOrder, error) {
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

	// 3. 检查积分
	totalPoints := product.Points * int64(quantity)
	if account.TotalPoints-account.UsedPoints < totalPoints {
		return nil, errors.New("insufficient points")
	}

	// 4. 扣减积分
	account.UsedPoints += totalPoints
	if err := m.repo.SaveAccount(ctx, account); err != nil {
		return nil, err
	}

	// 5. 扣减库存
	product.Stock -= quantity
	product.SoldCount += quantity
	if err := m.repo.SaveProduct(ctx, product); err != nil {
		// 如果需要回滚（为简洁起见省略）
		return nil, err
	}

	// 6. 创建订单
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
	if err := m.repo.SaveOrder(ctx, order); err != nil {
		return nil, err
	}

	// 7. 记录流水
	tx := &domain.PointsTransaction{
		UserID:      userID,
		Type:        "spend",
		Points:      -totalPoints,
		Description: fmt.Sprintf("Exchange product: %s", product.Name),
		RefID:       orderNo,
	}
	if err := m.repo.SaveTransaction(ctx, tx); err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "product exchanged successfully", "user_id", userID, "product_id", productID, "order_no", orderNo)
	return order, nil
}

// AddPoints 增加用户积分。
func (m *PointsManager) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	account, err := m.repo.GetAccount(ctx, userID)
	if err != nil {
		return err
	}
	// 注意：如果 repo 实现中不存在，GetAccount 会自动创建

	account.TotalPoints += points
	if err := m.repo.SaveAccount(ctx, account); err != nil {
		return err
	}

	tx := &domain.PointsTransaction{
		UserID:      userID,
		Type:        "earn",
		Points:      points,
		Description: description,
		RefID:       refID,
	}
	return m.repo.SaveTransaction(ctx, tx)
}
