package application

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository"
	"github.com/wyfcoding/ecommerce/pkg/idgen"

	"log/slog"
)

type AfterSalesService struct {
	repo        repository.AfterSalesRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

func NewAfterSalesService(repo repository.AfterSalesRepository, idGenerator idgen.Generator, logger *slog.Logger) *AfterSalesService {
	return &AfterSalesService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateAfterSales 创建售后申请
func (s *AfterSalesService) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType entity.AfterSalesType, reason, description string, images []string, items []*entity.AfterSalesItem) (*entity.AfterSales, error) {

	no := fmt.Sprintf("AS%d", s.idGenerator.Generate())
	afterSales := entity.NewAfterSales(no, orderID, orderNo, userID, asType, reason, description, images)

	// Add items
	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		afterSales.Items = append(afterSales.Items, item)
	}

	if err := s.repo.Create(ctx, afterSales); err != nil {
		s.logger.ErrorContext(ctx, "failed to create after-sales", "order_id", orderID, "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "after-sales request created successfully", "after_sales_id", afterSales.ID, "order_id", orderID)

	// Log creation
	s.logOperation(ctx, uint64(afterSales.ID), "User", "Create", "", entity.AfterSalesStatusPending.String(), "Created after-sales request")

	return afterSales, nil
}

// Approve 批准售后
func (s *AfterSalesService) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != entity.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String()
	afterSales.Approve(operator, amount)

	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	s.logOperation(ctx, id, operator, "Approve", oldStatus, afterSales.Status.String(), fmt.Sprintf("Approved amount: %d", amount))
	return nil
}

// Reject 拒绝售后
func (s *AfterSalesService) Reject(ctx context.Context, id uint64, operator, reason string) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != entity.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String()
	afterSales.Reject(operator, reason)

	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	s.logOperation(ctx, id, operator, "Reject", oldStatus, afterSales.Status.String(), reason)
	return nil
}

// List 获取列表
func (s *AfterSalesService) List(ctx context.Context, query *repository.AfterSalesQuery) ([]*entity.AfterSales, int64, error) {
	return s.repo.List(ctx, query)
}

// GetDetails 获取详情
func (s *AfterSalesService) GetDetails(ctx context.Context, id uint64) (*entity.AfterSales, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AfterSalesService) logOperation(ctx context.Context, asID uint64, operator, action, oldStatus, newStatus, remark string) {
	log := &entity.AfterSalesLog{
		AfterSalesID: asID,
		Operator:     operator,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Remark:       remark,
	}
	if err := s.repo.CreateLog(ctx, log); err != nil {
		s.logger.WarnContext(ctx, "failed to create after-sales log", "after_sales_id", asID, "error", err)
	}
}
