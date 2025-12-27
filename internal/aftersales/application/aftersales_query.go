package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

// AfterSalesQuery 处理所有售后相关的查询操作（Queries）。
type AfterSalesQuery struct {
	repo domain.AfterSalesRepository
}

// NewAfterSalesQuery 构造函数。
func NewAfterSalesQuery(repo domain.AfterSalesRepository) *AfterSalesQuery {
	return &AfterSalesQuery{repo: repo}
}

func (q *AfterSalesQuery) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	return q.repo.List(ctx, query)
}

func (q *AfterSalesQuery) GetDetails(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	return q.repo.GetByID(ctx, id)
}

func (q *AfterSalesQuery) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	return q.repo.GetSupportTicket(ctx, id)
}

func (q *AfterSalesQuery) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	return q.repo.ListSupportTickets(ctx, userID, status, page, pageSize)
}

func (q *AfterSalesQuery) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	return q.repo.ListSupportTicketMessages(ctx, ticketID)
}

func (q *AfterSalesQuery) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	return q.repo.GetConfig(ctx, key)
}
