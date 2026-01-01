package application

import (
	"context"
	"fmt"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	paymentv1 "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/idgen"

	"log/slog"
)

// AfterSalesService 结构体定义了售后管理相关的应用服务。
// 它协调领域层和基础设施层，处理售后申请的创建、审批、拒绝以及查询等业务流程。
type AfterSalesService struct {
	repo          domain.AfterSalesRepository
	idGenerator   idgen.Generator
	logger        *slog.Logger
	orderClient   orderv1.OrderServiceClient
	paymentClient paymentv1.PaymentServiceClient
}

// NewAfterSalesService 创建并返回一个新的 AfterSalesService 实例。
func NewAfterSalesService(
	repo domain.AfterSalesRepository,
	idGenerator idgen.Generator,
	logger *slog.Logger,
	orderClient orderv1.OrderServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
) *AfterSalesService {
	return &AfterSalesService{
		repo:          repo,
		idGenerator:   idGenerator,
		logger:        logger,
		orderClient:   orderClient,
		paymentClient: paymentClient,
	}
}

// CreateAfterSales 创建一个新的售后申请。
func (s *AfterSalesService) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType domain.AfterSalesType, reason, description string, images []string, items []*domain.AfterSalesItem,
) (*domain.AfterSales, error) {
	no := fmt.Sprintf("AS%d", s.idGenerator.Generate())
	afterSales := domain.NewAfterSales(no, orderID, orderNo, userID, asType, reason, description, images)

	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		afterSales.Items = append(afterSales.Items, item)
	}

	if err := s.repo.Create(ctx, afterSales); err != nil {
		s.logger.ErrorContext(ctx, "failed to create after-sales", "order_id", orderID, "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "after-sales request created successfully", "after_sales_id", afterSales.ID, "order_id", orderID)

	s.logOperation(ctx, uint64(afterSales.ID), "User", "Create", "", domain.AfterSalesStatusPending.String(), "Created after-sales request")

	return afterSales, nil
}

// Approve 批准一个售后申请。
func (s *AfterSalesService) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
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

// Reject 拒绝一个售后申请。
func (s *AfterSalesService) Reject(ctx context.Context, id uint64, operator, reason string) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if afterSales.Status != domain.AfterSalesStatusPending {
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

// List 获取售后申请列表，支持通过查询条件进行过滤。
func (s *AfterSalesService) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	return s.repo.List(ctx, query)
}

// GetDetails 获取售后申请的详细信息。
func (s *AfterSalesService) GetDetails(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	return s.repo.GetByID(ctx, id)
}

// logOperation 是一个辅助函数，用于记录售后操作日志。
func (s *AfterSalesService) logOperation(ctx context.Context, asID uint64, operator, action, oldStatus, newStatus, remark string) {
	log := &domain.AfterSalesLog{
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

// ProcessRefund 处理退款流程。
// 涉及跨服务调用：PaymentService 发起退款，OrderService 更新状态。
func (s *AfterSalesService) ProcessRefund(ctx context.Context, id uint64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for refund: %v", afterSales.Status)
	}

	// 1. 调用 PaymentService 发起退款
	// 注意：RequestRefundRequest 需要 PaymentTransactionId。
	// 这里假设 PaymentId 不在 AfterSales 中，我们暂不传递 PaymentTransactionId (0)，或者需要先查询 Order 获取。
	// 修正字段名：RefundAmount (int64), OrderId (uint64)
	_, err = s.paymentClient.RequestRefund(ctx, &paymentv1.RequestRefundRequest{
		OrderId:      afterSales.OrderID,
		UserId:       afterSales.UserID,
		RefundAmount: afterSales.ApprovalAmount,
		Reason:       afterSales.Reason,
	})
	if err != nil {
		s.logger.Error("payment service refund failed", "error", err)
		return fmt.Errorf("payment refund failed: %w", err)
	}

	// 2. 调用 OrderService 更新订单状态或记录退款
	_, err = s.orderClient.RequestRefund(ctx, &orderv1.RequestRefundRequest{
		OrderId:      afterSales.OrderID,
		UserId:       afterSales.UserID,
		RefundAmount: afterSales.ApprovalAmount,
		Reason:       afterSales.Reason,
	})
	if err != nil {
		// 记录错误但不阻断流程，可能需要人工介入
		s.logger.Error("order service refund update failed", "error", err)
	}

	// 3. 更新本地状态
	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	s.logOperation(ctx, id, "System", "ProcessRefund", "Approved", "Completed", "Refund processed successfully")
	return nil
}

// ProcessExchange 处理换货流程。
// 涉及跨服务调用：OrderService 创建新订单。
func (s *AfterSalesService) ProcessExchange(ctx context.Context, id uint64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != domain.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for exchange: %v", afterSales.Status)
	}

	// 1. 调用 OrderService 创建换货订单
	// 构建新订单请求 items (使用 OrderItemCreate)
	var orderItems []*orderv1.OrderItemCreate
	for _, item := range afterSales.Items {
		orderItems = append(orderItems, &orderv1.OrderItemCreate{
			ProductId: item.ProductID,
			SkuId:     item.SkuID,
			Quantity:  item.Quantity,
		})
	}

	_, err = s.orderClient.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		UserId: afterSales.UserID,
		Items:  orderItems,
		Remark: fmt.Sprintf("Exchange order for AS No: %s", afterSales.AfterSalesNo),
		// 这里可能需要特殊标记是换货订单，跳过支付
	})
	if err != nil {
		s.logger.Error("failed to create exchange order", "error", err)
		return fmt.Errorf("create exchange order failed: %w", err)
	}

	// 2. 更新本地状态
	afterSales.Status = domain.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	s.logOperation(ctx, id, "System", "ProcessExchange", "Approved", "Completed", "Exchange processed successfully")
	return nil
}

// --- Support Ticket Service Methods ---

// CreateSupportTicket 创建客服工单。
func (s *AfterSalesService) CreateSupportTicket(ctx context.Context, userID, orderID uint64, subject, description, category string, priority int8) (*domain.SupportTicket, error) {
	ticketNo := fmt.Sprintf("TCK%d", s.idGenerator.Generate())
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

	if err := s.repo.CreateSupportTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// GetSupportTicket 获取工单详情。
func (s *AfterSalesService) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	return s.repo.GetSupportTicket(ctx, id)
}

// UpdateSupportTicketStatus 更新工单状态。
func (s *AfterSalesService) UpdateSupportTicketStatus(ctx context.Context, id uint64, status domain.SupportTicketStatus) error {
	ticket, err := s.repo.GetSupportTicket(ctx, id)
	if err != nil {
		return err
	}
	if ticket == nil {
		return fmt.Errorf("ticket not found")
	}

	ticket.Status = status
	return s.repo.UpdateSupportTicket(ctx, ticket)
}

// ListSupportTickets 获取用户的工单列表。
func (s *AfterSalesService) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	return s.repo.ListSupportTickets(ctx, userID, status, page, pageSize)
}

// CreateSupportTicketMessage 为工单添加一条新消息（回复）。
func (s *AfterSalesService) CreateSupportTicketMessage(ctx context.Context, ticketID, senderID uint64, senderType, content string) (*domain.SupportTicketMessage, error) {
	msg := &domain.SupportTicketMessage{
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		IsRead:     false,
	}
	if err := s.repo.CreateSupportTicketMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// ListSupportTicketMessages 获取指定工单的所有消息记录。
func (s *AfterSalesService) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	return s.repo.ListSupportTicketMessages(ctx, ticketID)
}

// --- domain.AfterSales Config Service Methods ---

// GetConfig 根据键获取售后配置项。
func (s *AfterSalesService) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	return s.repo.GetConfig(ctx, key)
}

// SetConfig 设置（保存或更新）售后配置项。
func (s *AfterSalesService) SetConfig(ctx context.Context, key, value, description string) (*domain.AfterSalesConfig, error) {
	config := &domain.AfterSalesConfig{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := s.repo.SetConfig(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}
