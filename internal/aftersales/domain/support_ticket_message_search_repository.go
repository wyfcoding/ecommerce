// 生成摘要：定义客服工单消息搜索仓储接口（Elasticsearch）。
package domain

import "context"

// SupportTicketMessageSearchRepository 定义客服工单消息搜索接口。
type SupportTicketMessageSearchRepository interface {
	Index(ctx context.Context, message *SupportTicketMessage) error
	Delete(ctx context.Context, messageID uint64) error
	Search(ctx context.Context, ticketID uint64, offset, limit int) ([]*SupportTicketMessage, int64, error)
}
