package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity" // 导入客户服务领域的实体定义。
)

// CustomerServiceRepository 是客服模块的仓储接口。
// 它定义了对工单和消息实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type CustomerServiceRepository interface {
	// --- Ticket methods ---

	// SaveTicket 将工单实体保存到数据存储中。
	// 如果工单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// ticket: 待保存的工单实体。
	SaveTicket(ctx context.Context, ticket *entity.Ticket) error
	// GetTicket 根据ID获取工单实体。
	GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error)
	// GetTicketByNo 根据工单编号获取工单实体。
	GetTicketByNo(ctx context.Context, ticketNo string) (*entity.Ticket, error)
	// UpdateTicket 更新工单实体的信息。
	UpdateTicket(ctx context.Context, ticket *entity.Ticket) error
	// ListTickets 列出所有工单实体，支持通过用户ID和状态过滤，并支持分页。
	ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, offset, limit int) ([]*entity.Ticket, int64, error)

	// --- Message methods ---

	// SaveMessage 将消息实体保存到数据存储中。
	SaveMessage(ctx context.Context, message *entity.Message) error
	// ListMessages 列出指定工单的所有消息实体，支持分页。
	ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*entity.Message, int64, error)
}
