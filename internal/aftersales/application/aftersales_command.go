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

// AfterSalesCommandService 处理所有售后相关的写入操作（Commands）。
type AfterSalesCommandService struct {
	repo          domain.AfterSalesRepository
	publisher     domain.EventPublisher
	idGenerator   idgen.Generator
	logger        *slog.Logger
	orderClient   orderv1.OrderServiceClient
	paymentClient paymentv1.PaymentServiceClient
	dtmServer     string
	orderSvcURL   string
	paymentSvcURL string
	aftersalesURL string // 本服务回调地址
}

// NewAfterSalesCommandService 构造函数。
func NewAfterSalesCommandService(
	repo domain.AfterSalesRepository,
	publisher domain.EventPublisher,
	idGenerator idgen.Generator,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	dtmServer, orderSvcURL, paymentSvcURL, aftersalesURL string,
) *AfterSalesCommandService {
	return &AfterSalesCommandService{
		repo:          repo,
		publisher:     publisher,
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

func (m *AfterSalesCommandService) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType domain.AfterSalesType, reason, description string, images []string, items []*domain.AfterSalesItem,
) (*domain.AfterSales, error) {
	no := fmt.Sprintf("AS%d", m.idGenerator.Generate())
	afterSales := domain.NewAfterSales(no, orderID, orderNo, userID, asType, reason, description, images)

	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		afterSales.Items = append(afterSales.Items, item)
	}

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.CreateInTx(ctx, tx, afterSales); err != nil {
			m.logger.ErrorContext(ctx, "failed to create after-sales", "order_id", orderID, "user_id", userID, "error", err)
			return err
		}
		if err := m.logOperationInTx(ctx, tx, uint64(afterSales.ID), "User", "Create", "", domain.AfterSalesStatusPending.String(), "Created after-sales request"); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.AfterSalesCreatedEvent{
				AfterSalesID: afterSales.ID,
				AfterSalesNo: afterSales.AfterSalesNo,
				OrderID:      afterSales.OrderID,
				UserID:       afterSales.UserID,
				Type:         afterSales.Type,
				Status:       afterSales.Status,
				Timestamp:    time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.AfterSalesCreatedEventType, fmt.Sprintf("%d", afterSales.ID), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	m.logger.InfoContext(ctx, "after-sales request created successfully", "after_sales_id", afterSales.ID, "order_id", orderID)

	return afterSales, nil
}

func (m *AfterSalesCommandService) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales == nil {
		return domain.ErrAfterSalesNotFound
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status
	oldStatusStr := oldStatus.String()
	afterSales.Approve(operator, amount)

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateInTx(ctx, tx, afterSales); err != nil {
			return err
		}
		if err := m.logOperationInTx(ctx, tx, id, operator, "Approve", oldStatusStr, afterSales.Status.String(), fmt.Sprintf("Approved amount: %d", amount)); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, afterSales, oldStatus, operator)
	})
}

func (m *AfterSalesCommandService) Reject(ctx context.Context, id uint64, operator, reason string) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales == nil {
		return domain.ErrAfterSalesNotFound
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status
	oldStatusStr := oldStatus.String()
	afterSales.Reject(operator, reason)

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateInTx(ctx, tx, afterSales); err != nil {
			return err
		}
		if err := m.logOperationInTx(ctx, tx, id, operator, "Reject", oldStatusStr, afterSales.Status.String(), reason); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, afterSales, oldStatus, operator)
	})
}

// Saga 状态回调实现

// SagaMarkRefundCompleted 正向确认成功
func (m *AfterSalesCommandService) SagaMarkRefundCompleted(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales == nil {
		return err
	}
	if afterSales.Status == domain.AfterSalesStatusCompleted {
		return nil
	}

	oldStatus := afterSales.Status
	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateInTx(ctx, tx, afterSales); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, afterSales, oldStatus, "System")
	})
}

// SagaMarkRefundFailed 补偿标记失败
func (m *AfterSalesCommandService) SagaMarkRefundFailed(ctx context.Context, id uint64, reason string) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales == nil {
		return err
	}

	oldStatus := afterSales.Status
	afterSales.Status = domain.AfterSalesStatusRejected
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateInTx(ctx, tx, afterSales); err != nil {
			return err
		}
		if err := m.logOperationInTx(ctx, tx, id, "System", "SagaCompensation", oldStatus.String(), "FAILED", reason); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, afterSales, oldStatus, "System")
	})
}

// ProcessRefund 执行退款 (生产级 100% 可靠编排)
func (m *AfterSalesCommandService) ProcessRefund(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil || afterSales == nil {
		return err
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
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

	if err := saga.Submit(ctx); err != nil {
		return fmt.Errorf("failed to submit saga: %w", err)
	}

	return nil
}

func (m *AfterSalesCommandService) ProcessExchange(ctx context.Context, id uint64) error {
	afterSales, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales == nil {
		return domain.ErrAfterSalesNotFound
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for exchange: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status
	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateInTx(ctx, tx, afterSales); err != nil {
			return err
		}
		if err := m.logOperationInTx(ctx, tx, id, "System", "ProcessExchange", "Approved", "Completed", "Exchange processed successfully"); err != nil {
			return err
		}
		return m.publishStatusUpdated(ctx, tx, afterSales, oldStatus, "System")
	})
}

func (m *AfterSalesCommandService) CreateSupportTicket(ctx context.Context, userID, orderID uint64, subject, description, category string, priority int8) (*domain.SupportTicket, error) {
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

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.CreateSupportTicketInTx(ctx, tx, ticket); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.SupportTicketCreatedEvent{
				TicketID:  ticket.ID,
				TicketNo:  ticket.TicketNo,
				UserID:    ticket.UserID,
				OrderID:   ticket.OrderID,
				Timestamp: time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.AfterSalesSupportTicketCreatedType, fmt.Sprintf("%d", ticket.ID), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ticket, nil
}

func (m *AfterSalesCommandService) UpdateSupportTicketStatus(ctx context.Context, id uint64, status domain.SupportTicketStatus) error {
	ticket, err := m.repo.GetSupportTicket(ctx, id)
	if err != nil {
		return err
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	ticket.Status = status
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateSupportTicketInTx(ctx, tx, ticket); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.SupportTicketUpdatedEvent{
				TicketID:  ticket.ID,
				Status:    ticket.Status,
				Timestamp: time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.AfterSalesSupportTicketUpdatedType, fmt.Sprintf("%d", ticket.ID), event)
		}
		return nil
	})
}

func (m *AfterSalesCommandService) CreateSupportTicketMessage(ctx context.Context, ticketID, senderID uint64, senderType, content string) (*domain.SupportTicketMessage, error) {
	msg := &domain.SupportTicketMessage{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		IsRead:     false,
	}
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.CreateSupportTicketMessageInTx(ctx, tx, msg); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.SupportTicketMessageCreatedEvent{
				MessageID: msg.ID,
				TicketID:  msg.TicketID,
				SenderID:  msg.SenderID,
				Timestamp: time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.AfterSalesSupportTicketMessageType, fmt.Sprintf("%d", msg.ID), event)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *AfterSalesCommandService) SetConfig(ctx context.Context, key, value, description string) (*domain.AfterSalesConfig, error) {
	config := &domain.AfterSalesConfig{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SetConfigInTx(ctx, tx, config); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.AfterSalesConfigUpdatedEvent{
				Key:       config.Key,
				Timestamp: time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.AfterSalesConfigUpdatedEventType, config.Key, event)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return config, nil
}

func (m *AfterSalesCommandService) LogOperation(ctx context.Context, asID uint64, operator, action, oldStatus, newStatus, remark string) {
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

func (m *AfterSalesCommandService) logOperationInTx(ctx context.Context, tx any, asID uint64, operator, action, oldStatus, newStatus, remark string) error {
	log := &domain.AfterSalesLog{
		AfterSalesID: asID,
		Operator:     operator,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Remark:       remark,
	}
	return m.repo.CreateLogInTx(ctx, tx, log)
}

func (m *AfterSalesCommandService) publishStatusUpdated(ctx context.Context, tx any, afterSales *domain.AfterSales, oldStatus domain.AfterSalesStatus, operator string) error {
	if m.publisher == nil || afterSales == nil {
		return nil
	}
	event := &domain.AfterSalesStatusUpdatedEvent{
		AfterSalesID: afterSales.ID,
		AfterSalesNo: afterSales.AfterSalesNo,
		OldStatus:    oldStatus,
		NewStatus:    afterSales.Status,
		Operator:     operator,
		Timestamp:    time.Now(),
	}
	return m.publisher.PublishInTx(ctx, tx, domain.AfterSalesStatusUpdatedEventType, fmt.Sprintf("%d", afterSales.ID), event)
}
