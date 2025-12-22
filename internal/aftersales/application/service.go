package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/entity"     // 导入售后领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain/repository" // 导入售后领域的仓储接口。
	"github.com/wyfcoding/pkg/idgen"                                       // 导入ID生成器接口。

	"log/slog" // 导入结构化日志库。
)

// AfterSalesService 结构体定义了售后管理相关的应用服务。
// 它协调领域层和基础设施层，处理售后申请的创建、审批、拒绝以及查询等业务流程。
type AfterSalesService struct {
	repo        repository.AfterSalesRepository // 依赖AfterSalesRepository接口，用于数据持久化操作。
	idGenerator idgen.Generator                 // 依赖ID生成器接口，用于生成唯一的售后单号。
	logger      *slog.Logger                    // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewAfterSalesService 创建并返回一个新的 AfterSalesService 实例。
func NewAfterSalesService(repo repository.AfterSalesRepository, idGenerator idgen.Generator, logger *slog.Logger) *AfterSalesService {
	return &AfterSalesService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// CreateAfterSales 创建一个新的售后申请。
// ctx: 上下文。
// orderID: 关联的订单ID。
// orderNo: 关联的订单号。
// userID: 发起售后申请的用户ID。
// asType: 售后类型（例如，退货、退款、换货）。
// reason: 售后原因。
// description: 售后描述。
// images: 相关的图片URL列表。
// items: 申请售后的商品列表。
// 返回created successfully的售后实体和可能发生的错误。
func (s *AfterSalesService) CreateAfterSales(ctx context.Context, orderID uint64, orderNo string, userID uint64,
	asType entity.AfterSalesType, reason, description string, images []string, items []*entity.AfterSalesItem,
) (*entity.AfterSales, error) {
	// 生成唯一的售后单号。
	no := fmt.Sprintf("AS%d", s.idGenerator.Generate())
	// 创建售后实体。
	afterSales := entity.NewAfterSales(no, orderID, orderNo, userID, asType, reason, description, images)

	// 添加申请售后的商品项，并计算每个商品的总价。
	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		afterSales.Items = append(afterSales.Items, item)
	}

	// 通过仓储接口将售后实体保存到数据库。
	if err := s.repo.Create(ctx, afterSales); err != nil {
		s.logger.ErrorContext(ctx, "failed to create after-sales", "order_id", orderID, "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "after-sales request created successfully", "after_sales_id", afterSales.ID, "order_id", orderID)

	// 记录售后操作日志。
	s.logOperation(ctx, uint64(afterSales.ID), "User", "Create", "", entity.AfterSalesStatusPending.String(), "Created after-sales request")

	return afterSales, nil
}

// Approve 批准一个售后申请。
// ctx: 上下文。
// id: 售后申请的ID。
// operator: 执行批准操作的人员（例如，管理员用户名）。
// amount: 批准的退款金额或补偿金额。
// 返回可能发生的错误。
func (s *AfterSalesService) Approve(ctx context.Context, id uint64, operator string, amount int64) error {
	// 根据ID获取售后申请。
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查售后申请当前的状态是否允许批准操作。
	if afterSales.Status != entity.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String() // 记录旧状态。
	afterSales.Approve(operator, amount)    // 调用实体方法批准售后。

	// 更新数据库中的售后申请状态。
	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	// 记录售后操作日志。
	s.logOperation(ctx, id, operator, "Approve", oldStatus, afterSales.Status.String(), fmt.Sprintf("Approved amount: %d", amount))
	return nil
}

// Reject 拒绝一个售后申请。
// ctx: 上下文。
// id: 售后申请的ID。
// operator: 执行拒绝操作的人员。
// reason: 拒绝的原因。
// 返回可能发生的错误。
func (s *AfterSalesService) Reject(ctx context.Context, id uint64, operator, reason string) error {
	// 根据ID获取售后申请。
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查售后申请当前的状态是否允许拒绝操作。
	if afterSales.Status != entity.AfterSalesStatusPending {
		return fmt.Errorf("invalid status: %v", afterSales.Status)
	}

	oldStatus := afterSales.Status.String() // 记录旧状态。
	afterSales.Reject(operator, reason)     // 调用实体方法拒绝售后。

	// 更新数据库中的售后申请状态。
	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	// 记录售后操作日志。
	s.logOperation(ctx, id, operator, "Reject", oldStatus, afterSales.Status.String(), reason)
	return nil
}

// List 获取售后申请列表，支持通过查询条件进行过滤。
// ctx: 上下文。
// query: 包含过滤条件和分页参数的查询对象。
// 返回售后申请列表、总数和可能发生的错误。
func (s *AfterSalesService) List(ctx context.Context, query *repository.AfterSalesQuery) ([]*entity.AfterSales, int64, error) {
	return s.repo.List(ctx, query)
}

// GetDetails 获取售后申请的详细信息。
// ctx: 上下文。
// id: 售后申请的ID。
// 返回售后实体和可能发生的错误。
func (s *AfterSalesService) GetDetails(ctx context.Context, id uint64) (*entity.AfterSales, error) {
	return s.repo.GetByID(ctx, id)
}

// logOperation 是一个辅助函数，用于记录售后操作日志。
// ctx: 上下文。
// asID: 售后申请ID。
// operator: 操作人员。
// action: 操作类型。
// oldStatus: 操作前的状态。
// newStatus: 操作后的状态。
// remark: 操作备注。
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

// ProcessRefund 处理退款流程。
// ctx: 上下文。
// id: 售后申请ID。
// 返回可能发生的错误。
func (s *AfterSalesService) ProcessRefund(ctx context.Context, id uint64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != entity.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for refund: %v", afterSales.Status)
	}

	// 模拟逻辑 process
	afterSales.Status = entity.AfterSalesStatusCompleted
	now := time.Now()
	afterSales.CompletedAt = &now

	if err := s.repo.Update(ctx, afterSales); err != nil {
		return err
	}

	s.logOperation(ctx, id, "System", "ProcessRefund", "Approved", "Completed", "Refund processed successfully")
	return nil
}

// ProcessExchange 处理换货流程。
// ctx: 上下文。
// id: 售后申请ID。
// 返回可能发生的错误。
func (s *AfterSalesService) ProcessExchange(ctx context.Context, id uint64) error {
	afterSales, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if afterSales.Status != entity.AfterSalesStatusApproved {
		return fmt.Errorf("invalid status for exchange: %v", afterSales.Status)
	}

	// 模拟逻辑 process
	afterSales.Status = entity.AfterSalesStatusCompleted
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
func (s *AfterSalesService) CreateSupportTicket(ctx context.Context, userID, orderID uint64, subject, description, category string, priority int8) (*entity.SupportTicket, error) {
	ticketNo := fmt.Sprintf("TCK%d", s.idGenerator.Generate())
	ticket := &entity.SupportTicket{
		TicketNo:    ticketNo,
		UserID:      userID,
		OrderID:     orderID,
		Subject:     subject,
		Description: description,
		Status:      entity.SupportTicketStatusOpen,
		Priority:    priority,
		Category:    category,
		Messages:    []*entity.SupportTicketMessage{},
	}

	if err := s.repo.CreateSupportTicket(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// GetSupportTicket 获取工单详情。
func (s *AfterSalesService) GetSupportTicket(ctx context.Context, id uint64) (*entity.SupportTicket, error) {
	return s.repo.GetSupportTicket(ctx, id)
}

// UpdateSupportTicketStatus 更新工单状态。
func (s *AfterSalesService) UpdateSupportTicketStatus(ctx context.Context, id uint64, status entity.SupportTicketStatus) error {
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
func (s *AfterSalesService) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.SupportTicket, int64, error) {
	return s.repo.ListSupportTickets(ctx, userID, status, page, pageSize)
}

// CreateSupportTicketMessage 为工单添加一条新消息（回复）。
func (s *AfterSalesService) CreateSupportTicketMessage(ctx context.Context, ticketID, senderID uint64, senderType, content string) (*entity.SupportTicketMessage, error) {
	msg := &entity.SupportTicketMessage{
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
func (s *AfterSalesService) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*entity.SupportTicketMessage, error) {
	return s.repo.ListSupportTicketMessages(ctx, ticketID)
}

// --- AfterSales Config Service Methods ---

// GetConfig 根据键获取售后配置项。
func (s *AfterSalesService) GetConfig(ctx context.Context, key string) (*entity.AfterSalesConfig, error) {
	return s.repo.GetConfig(ctx, key)
}

// SetConfig 设置（保存或更新）售后配置项。
func (s *AfterSalesService) SetConfig(ctx context.Context, key, value, description string) (*entity.AfterSalesConfig, error) {
	config := &entity.AfterSalesConfig{
		Key:         key,
		Value:       value,
		Description: description,
	}
	if err := s.repo.SetConfig(ctx, config); err != nil {
		return nil, err
	}
	return config, nil
}
