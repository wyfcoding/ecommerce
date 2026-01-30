package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

// AfterSalesQueryService 处理所有售后相关的查询操作（Queries）。
type AfterSalesQueryService struct {
	repo domain.AfterSalesRepository
}

// NewAfterSalesQueryService 构造函数。
func NewAfterSalesQueryService(repo domain.AfterSalesRepository) *AfterSalesQueryService {
	return &AfterSalesQueryService{repo: repo}
}

func (q *AfterSalesQueryService) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	return q.repo.List(ctx, query)
}

func (q *AfterSalesQueryService) GetDetails(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	return q.repo.GetByID(ctx, id)
}

func (q *AfterSalesQueryService) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	return q.repo.GetSupportTicket(ctx, id)
}

func (q *AfterSalesQueryService) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	return q.repo.ListSupportTickets(ctx, userID, status, page, pageSize)
}

func (q *AfterSalesQueryService) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	return q.repo.ListSupportTicketMessages(ctx, ticketID)
}

func (q *AfterSalesQueryService) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	return q.repo.GetConfig(ctx, key)
}
