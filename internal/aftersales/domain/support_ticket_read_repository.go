// 生成摘要：定义售后客服工单读模型仓储接口（Redis）。
package domain

import "context"

// SupportTicketReadRepository 定义客服工单读模型接口。
type SupportTicketReadRepository interface {
	Save(ctx context.Context, ticket *SupportTicket) error
	GetByID(ctx context.Context, id uint64) (*SupportTicket, error)
	Delete(ctx context.Context, id uint64) error
}
