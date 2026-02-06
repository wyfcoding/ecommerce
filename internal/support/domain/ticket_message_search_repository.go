// 生成摘要：定义工单消息搜索仓储接口（Elasticsearch）。
package domain

import "context"

// TicketMessageSearchRepository 定义工单消息搜索接口。
type TicketMessageSearchRepository interface {
	Index(ctx context.Context, message *Message) error
	Delete(ctx context.Context, messageID uint64) error
	Search(ctx context.Context, ticketID uint64, offset, limit int) ([]*Message, int64, error)
}
