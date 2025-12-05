package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity"     // 导入客户服务领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/repository" // 导入客户服务领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// CustomerService 结构体定义了客户服务（工单）相关的应用服务。
// 它协调领域层和基础设施层，处理工单的创建、回复、状态变更和查询等业务逻辑。
type CustomerService struct {
	repo   repository.CustomerServiceRepository // 依赖CustomerServiceRepository接口，用于数据持久化操作。
	logger *slog.Logger                         // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewCustomerService 创建并返回一个新的 CustomerService 实例。
func NewCustomerService(repo repository.CustomerServiceRepository, logger *slog.Logger) *CustomerService {
	return &CustomerService{
		repo:   repo,
		logger: logger,
	}
}

// CreateTicket 创建一个新的客户服务工单。
// ctx: 上下文。
// userID: 提交工单的用户ID。
// subject: 工单主题。
// description: 工单详细描述。
// category: 工单类别。
// priority: 工单优先级。
// 返回created successfully的Ticket实体和可能发生的错误。
func (s *CustomerService) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority entity.TicketPriority) (*entity.Ticket, error) {
	// 生成唯一的工单编号。
	ticketNo := fmt.Sprintf("TKT%d", time.Now().UnixNano())
	// 创建Ticket实体。
	ticket := entity.NewTicket(ticketNo, userID, subject, description, category, priority)

	// 通过仓储接口保存工单。
	if err := s.repo.SaveTicket(ctx, ticket); err != nil {
		s.logger.ErrorContext(ctx, "failed to create ticket", "user_id", userID, "subject", subject, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "ticket created successfully", "ticket_id", ticket.ID, "ticket_no", ticketNo)
	return ticket, nil
}

// ReplyTicket 回复一个工单，并可能更新工单状态。
// ctx: 上下文。
// ticketID: 待回复的工单ID。
// senderID: 发送消息的用户或客服ID。
// senderType: 发送者类型（"user"或"support"）。
// content: 消息内容。
// msgType: 消息类型。
// 返回created successfully的Message实体和可能发生的错误。
func (s *CustomerService) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType entity.MessageType) (*entity.Message, error) {
	// 获取工单实体。
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// 如果消息由非用户（例如，客服或管理员）发送，并且工单状态为Open，则更新工单状态为InProgress。
	if senderType != "user" && ticket.Status == entity.TicketStatusOpen {
		ticket.Status = entity.TicketStatusInProgress
		// 更新数据库中的工单状态。
		if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
			return nil, err
		}
	}

	// 创建Message实体。
	message := entity.NewMessage(ticketID, senderID, senderType, content, msgType, false)
	// 通过仓储接口保存消息。
	if err := s.repo.SaveMessage(ctx, message); err != nil {
		s.logger.ErrorContext(ctx, "failed to save message", "ticket_id", ticketID, "sender_id", senderID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "message saved successfully", "message_id", message.ID, "ticket_id", ticketID)

	return message, nil
}

// GetTicket 获取指定ID的工单详情。
// ctx: 上下文。
// id: 工单ID。
// 返回Ticket实体和可能发生的错误。
func (s *CustomerService) GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error) {
	return s.repo.GetTicket(ctx, id)
}

// ListTickets 获取工单列表，支持通过用户ID和状态过滤。
// ctx: 上下文。
// userID: 筛选工单的用户ID。
// status: 筛选工单的状态。
// page, pageSize: 分页参数。
// 返回工单列表、总数和可能发生的错误。
func (s *CustomerService) ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, page, pageSize int) ([]*entity.Ticket, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTickets(ctx, userID, status, offset, pageSize)
}

// ListMessages 获取指定工单的所有消息列表。
// ctx: 上下文。
// ticketID: 工单ID。
// page, pageSize: 分页参数。
// 返回消息列表、总数和可能发生的错误。
func (s *CustomerService) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*entity.Message, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListMessages(ctx, ticketID, offset, pageSize)
}

// CloseTicket 关闭一个工单。
// ctx: 上下文。
// id: 工单ID。
// 返回可能发生的错误。
func (s *CustomerService) CloseTicket(ctx context.Context, id uint64) error {
	// 获取工单实体。
	ticket, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法关闭工单。
	ticket.Close()
	// 更新数据库中的工单状态。
	return s.repo.UpdateTicket(ctx, ticket)
}

// ResolveTicket 解决一个工单。
// ctx: 上下文。
// id: 工单ID。
// 返回可能发生的错误。
func (s *CustomerService) ResolveTicket(ctx context.Context, id uint64) error {
	// 获取工单实体。
	ticket, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法解决工单。
	ticket.Resolve()
	// 更新数据库中的工单状态。
	return s.repo.UpdateTicket(ctx, ticket)
}
