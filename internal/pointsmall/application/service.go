package application

import (
	"context"
	"ecommerce/internal/pointsmall/domain/entity"
	"ecommerce/internal/pointsmall/domain/repository"
	"ecommerce/pkg/idgen"
	"errors"
	"fmt"
	"time"

	"log/slog"
)

type PointsService struct {
	repo   repository.PointsRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

func NewPointsService(repo repository.PointsRepository, idGen idgen.Generator, logger *slog.Logger) *PointsService {
	return &PointsService{
		repo:   repo,
		idGen:  idGen,
		logger: logger,
	}
}

// CreateProduct 创建商品
func (s *PointsService) CreateProduct(ctx context.Context, product *entity.PointsProduct) error {
	return s.repo.SaveProduct(ctx, product)
}

// ListProducts 商品列表
func (s *PointsService) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*entity.PointsProduct, int64, error) {
	offset := (page - 1) * pageSize
	var prodStatus *entity.PointsProductStatus
	if status != nil {
		s := entity.PointsProductStatus(*status)
		prodStatus = &s
	}
	return s.repo.ListProducts(ctx, prodStatus, offset, pageSize)
}

// ExchangeProduct 兑换商品
func (s *PointsService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*entity.PointsOrder, error) {
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.Status != entity.PointsProductStatusOnline {
		return nil, errors.New("product not online")
	}
	if product.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	account, err := s.repo.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalPoints := product.Points * int64(quantity)
	if account.TotalPoints-account.UsedPoints < totalPoints {
		return nil, errors.New("insufficient points")
	}

	// Transactional logic should be here (omitted for simplicity, but critical in production)
	// 1. Deduct points
	account.UsedPoints += totalPoints
	if err := s.repo.SaveAccount(ctx, account); err != nil {
		return nil, err
	}

	// 2. Deduct stock
	product.Stock -= quantity
	product.SoldCount += quantity
	if err := s.repo.SaveProduct(ctx, product); err != nil {
		return nil, err
	}

	// 3. Create order
	orderID := s.idGen.Generate()
	orderNo := fmt.Sprintf("P%s%d", time.Now().Format("20060102"), orderID)
	order := &entity.PointsOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		ProductName: product.Name,
		Quantity:    quantity,
		Points:      product.Points,
		TotalPoints: totalPoints,
		Status:      entity.PointsOrderStatusPending,
		Address:     address,
		Phone:       phone,
		Receiver:    receiver,
	}
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return nil, err
	}

	// 4. Record transaction
	tx := &entity.PointsTransaction{
		UserID:      userID,
		Type:        "spend",
		Points:      -totalPoints,
		Description: fmt.Sprintf("Exchange product: %s", product.Name),
		RefID:       orderNo,
	}
	if err := s.repo.SaveTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return order, nil
}

// GetAccount 获取账户
func (s *PointsService) GetAccount(ctx context.Context, userID uint64) (*entity.PointsAccount, error) {
	return s.repo.GetAccount(ctx, userID)
}

// AddPoints 增加积分 (Admin/System)
func (s *PointsService) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	account, err := s.repo.GetAccount(ctx, userID)
	if err != nil {
		return err
	}

	account.TotalPoints += points
	if err := s.repo.SaveAccount(ctx, account); err != nil {
		return err
	}

	tx := &entity.PointsTransaction{
		UserID:      userID,
		Type:        "earn",
		Points:      points,
		Description: description,
		RefID:       refID,
	}
	return s.repo.SaveTransaction(ctx, tx)
}

// ListOrders 订单列表
func (s *PointsService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.PointsOrder, int64, error) {
	offset := (page - 1) * pageSize
	var orderStatus *entity.PointsOrderStatus
	if status != nil {
		s := entity.PointsOrderStatus(*status)
		orderStatus = &s
	}
	return s.repo.ListOrders(ctx, userID, orderStatus, offset, pageSize)
}
