package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	aftersalesv1 "github.com/wyfcoding/ecommerce/goapi/aftersales/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/dtm"
	"github.com/wyfcoding/pkg/idgen"
)

// AfterSalesManager 处理所有售后相关的写入操作（Commands）。
type AfterSalesManager struct {
	repo          domain.AfterSalesRepository
	idGenerator   idgen.Generator
	logger        *slog.Logger
	orderClient   orderv1.OrderServiceClient
	paymentClient paymentv1.PaymentServiceClient
	dtmServer     string
	orderSvcURL   string
	paymentSvcURL string
	aftersalesURL string // 本服务回调地址
}

// NewAfterSalesManager 构造函数。
func NewAfterSalesManager(
	repo domain.AfterSalesRepository,
	idGenerator idgen.Generator,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	dtmServer, orderSvcURL, paymentSvcURL, aftersalesURL string,
) *AfterSalesManager {
	return &AfterSalesManager{
		repo:          repo,
		idGenerator:   idGenerator,
		logger:        logger,
		orderClient:   orderClient,
		paymentClient: paymentClient,
		dtmServer:     dtmServer,
		orderSvcURL:   orderSvcURL,
		paymentSvcURL: paymentSvcURL,
		aftersalesURL: aftersalesURL,
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

// Saga 状态回调实现

// SagaMarkRefundCompleted 正向确认成功
func (m *AfterSalesManager) SagaMarkRefundCompleted(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales == nil { return err }
	if afterSales.Status == domain.AfterSalesStatusCompleted { return nil }
	
	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now
	return m.repo.Update(ctx, afterSales)
}

// SagaMarkRefundFailed 补偿标记失败
func (m *AfterSalesManager) SagaMarkRefundFailed(ctx context.Context, id uint64, reason string) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales == nil { return err }
	
	afterSales.Status = domain.AfterSalesStatusRejected
	m.LogOperation(ctx, id, "System", "SagaCompensation", "", "FAILED", reason)
	return m.repo.Update(ctx, afterSales)
}

// ProcessRefund 执行退款 (生产级 100% 可靠编排)
func (m *AfterSalesManager) ProcessRefund(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("request not ready for refund")
	}

	m.logger.InfoContext(ctx, "starting full saga refund orchestration", "as_no", afterSales.AfterSalesNo)

	gid := fmt.Sprintf("SAGA-AS-REFUND-%s", afterSales.AfterSalesNo)
	saga := dtm.NewSaga(ctx, m.dtmServer, gid)

	paymentSvc := m.paymentSvcURL + "/api.payment.v1.PaymentService"
	orderSvc := m.orderSvcURL + "/api.order.v1.OrderService"
	aftersalesSvc := m.aftersalesURL + "/api.aftersales.v1.AftersalesService"

	// 1. 状态追踪桩
	saga.Add("", aftersalesSvc+"/SagaMarkRefundFailed", &aftersalesv1.SagaAftersalesRequest{
		AftersalesId: uint64(afterSales.ID),
		Reason:       "Transaction Rolled Back",
	})

	// 2. 资金退回
	saga.Add(paymentSvc+"/SagaRefund", paymentSvc+"/SagaCancelRefund", &paymentv1.SagaRefundRequest{
		UserId: afterSales.UserID, OrderId: afterSales.OrderID, RefundAmount: afterSales.RefundAmount,
	})

	// 3. 订单状态变更 (改为 CANCELLED/REFUNDED)
	saga.Add(orderSvc+"/SagaCancelOrder", "", &orderv1.SagaOrderRequest{
		UserId: afterSales.UserID, OrderId: afterSales.OrderID, Reason: "Aftersales Refund",
	})

	// 4. 最终状态确认
	saga.Add(aftersalesSvc+"/SagaMarkRefundCompleted", "", &aftersalesv1.SagaAftersalesRequest{
		AftersalesId: uint64(afterSales.ID),
	})

	if err := saga.Submit(); err != nil {
		return fmt.Errorf("failed to submit saga: %w", err)
	}

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
