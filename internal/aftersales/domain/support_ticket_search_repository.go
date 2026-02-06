// 生成摘要：定义客服工单搜索仓储接口（Elasticsearch）。
package domain

import "context"

// SupportTicketSearchRepository 定义客服工单搜索访问接口。
type SupportTicketSearchRepository interface {
	Index(ctx context.Context, ticket *SupportTicket) error
	Delete(ctx context.Context, ticketID uint64) error
	Search(ctx context.Context, userID uint64, status *int, offset, limit int) ([]*SupportTicket, int64, error)
}
