// 生成摘要：定义工单搜索仓储接口（Elasticsearch）。
package domain

import "context"

// TicketSearchRepository 定义工单搜索访问接口。
type TicketSearchRepository interface {
	Index(ctx context.Context, ticket *Ticket) error
	Delete(ctx context.Context, ticketID uint64) error
	Search(ctx context.Context, userID uint64, status *TicketStatus, offset, limit int) ([]*Ticket, int64, error)
}
