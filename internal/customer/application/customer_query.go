package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer/domain"
)

// CustomerQuery 处理客户服务的读操作。
type CustomerQuery struct {
	repo domain.CustomerRepository
}

// NewCustomerQuery 创建并返回一个新的 CustomerQuery 实例。
func NewCustomerQuery(repo domain.CustomerRepository) *CustomerQuery {
	return &CustomerQuery{
		repo: repo,
	}
}

// GetTicket 获取指定ID的工单详情。
func (q *CustomerQuery) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	return q.repo.GetTicket(ctx, id)
}

// ListTickets 获取工单列表，支持通过用户ID和状态过滤。
func (q *CustomerQuery) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListTickets(ctx, userID, status, offset, pageSize)
}

// ListMessages 获取指定工单的所有消息列表。
func (q *CustomerQuery) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListMessages(ctx, ticketID, offset, pageSize)
}
