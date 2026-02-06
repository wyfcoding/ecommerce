// 生成摘要：定义工单读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// TicketReadRepository 定义工单读模型接口。
type TicketReadRepository interface {
	Save(ctx context.Context, ticket *Ticket) error
	GetByID(ctx context.Context, id uint64) (*Ticket, error)
	GetByNo(ctx context.Context, ticketNo string) (*Ticket, error)
	Delete(ctx context.Context, id uint64, ticketNo string) error
}
