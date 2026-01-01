package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/idgen"
)

// AfterSalesManager 处理所有售后相关的写入操作（Commands）。
type AfterSalesManager struct {
	repo          domain.AfterSalesRepository
	idGenerator   idgen.Generator
	logger        *slog.Logger
	orderClient   orderv1.OrderServiceClient
	paymentClient paymentv1.PaymentServiceClient
}

// NewAfterSalesManager 构造函数。
func NewAfterSalesManager(
	repo domain.AfterSalesRepository,
	idGenerator idgen.Generator,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
) *AfterSalesManager {
	return &AfterSalesManager{
		repo:          repo,
		idGenerator:   idGenerator,
		logger:        logger,
		orderClient:   orderClient,
		paymentClient: paymentClient,
	}
}

func (m *AfterSalesManager) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType domain.AfterSalesType, reason, description string, images []string, items []*domain.AfterSalesItem,
) (*domain.AfterSales, error) {
	no := fmt.Sprintf("AS%d", m.idGenerator.Generate())
	afterSales := domain.NewAfterSales(no, orderID, orderNo, userID, asType, reason, description, images)

	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		afterSales.Items = append(afterSales.Items, item)
	}

	if err := m.repo.Create(ctx, afterSales); err != nil {
		m.logger.ErrorContext(ctx, "failed to create after-sales", "order_id", orderID, "user_id", userID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "after-sales request created successfully", "after_sales_id", afterSales.ID, "order_id", orderID)

	m.LogOperation(ctx, uint64(afterSales.ID), "User", "Create", "", domain.AfterSalesStatusPending.String(), "Created after-sales request")

	return afterSales, nil
}

func (m *AfterSalesManager) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String()
	afterSales.Approve(operator, amount)

	if err := m.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	m.LogOperation(ctx, id, operator, "Approve", oldStatus, afterSales.Status.String(), fmt.Sprintf("Approved amount: %d", amount))
	return nil
}

func (m *AfterSalesManager) Reject(ctx context.Context, id uint64, operator, reason string) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String()
	afterSales.Reject(operator, reason)

	if err := m.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	m.LogOperation(ctx, id, operator, "Reject", oldStatus, afterSales.Status.String(), reason)
	return nil
}

func (m *AfterSalesManager) ProcessRefund(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for refund: %v", afterSales.Status)
	}

	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	if err := m.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	m.LogOperation(ctx, id, "System", "ProcessRefund", "Approved", "Completed", "Refund processed successfully")
	return nil
}

func (m *AfterSalesManager) ProcessExchange(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for exchange: %v", afterSales.Status)
	}

	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	if err := m.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	m.LogOperation(ctx, id, "System", "ProcessExchange", "Approved", "Completed", "Exchange processed successfully")
	return nil
}

func (m *AfterSalesManager) CreateSupportTicket(ctx context.Context, userID, orderID uint64, subject, description, category string, priority int8) (*domain.SupportTicket, error) {
	ticketNo := fmt.Sprintf("TCK%d", m.idGenerator.Generate())
	ticket := &domain.SupportTicket{
		TicketNo:    ticketNo,
		UserID:      userID,
		OrderID:     orderID,
		Subject:     subject,
		Description: description,
		Status:      domain.SupportTicketStatusOpen,
		Priority:    priority,
		Category:    category,
		Messages:    []*domain.SupportTicketMessage{},
	}

	if err := m.repo.CreateSupportTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

func (m *AfterSalesManager) UpdateSupportTicketStatus(ctx context.Context, id uint64, status domain.SupportTicketStatus) error {
	ticket, err := m.repo.GetSupportTicket(ctx, id)
	if err != nil {
		return err
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	ticket.Status = status
	return m.repo.UpdateSupportTicket(ctx, ticket)
}

func (m *AfterSalesManager) CreateSupportTicketMessage(ctx context.Context, ticketID, senderID uint64, senderType, content string) (*domain.SupportTicketMessage, error) {
	msg := &domain.SupportTicketMessage{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		IsRead:     false,
	}
	if err := m.repo.CreateSupportTicketMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *AfterSalesManager) SetConfig(ctx context.Context, key, value, description string) (*domain.AfterSalesConfig, error) {
	config := &domain.AfterSalesConfig{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := m.repo.SetConfig(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}

func (m *AfterSalesManager) LogOperation(ctx context.Context, asID uint64, operator, action, oldStatus, newStatus, remark string) {
	log := &domain.AfterSalesLog{
		AfterSalesID: asID,
		Operator:     operator,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Remark:       remark,
	}
	if err := m.repo.CreateLog(ctx, log); err != nil {
		m.logger.WarnContext(ctx, "failed to create after-sales log", "after_sales_id", asID, "error", err)
	}
}
