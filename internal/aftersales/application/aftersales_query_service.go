package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

// AfterSalesQueryService 处理所有售后相关的查询操作（Queries）。
type AfterSalesQueryService struct {
	repo                           domain.AfterSalesRepository
	readRepo                       domain.AfterSalesReadRepository
	searchRepo                     domain.AfterSalesSearchRepository
	supportTicketReadRepo          domain.SupportTicketReadRepository
	supportTicketSearchRepo        domain.SupportTicketSearchRepository
	supportTicketMessageSearchRepo domain.SupportTicketMessageSearchRepository
	configReadRepo                 domain.AfterSalesConfigReadRepository
}

// NewAfterSalesQueryService 构造函数。
func NewAfterSalesQueryService(
	repo domain.AfterSalesRepository,
	readRepo domain.AfterSalesReadRepository,
	searchRepo domain.AfterSalesSearchRepository,
	supportTicketReadRepo domain.SupportTicketReadRepository,
	supportTicketSearchRepo domain.SupportTicketSearchRepository,
	supportTicketMessageSearchRepo domain.SupportTicketMessageSearchRepository,
	configReadRepo domain.AfterSalesConfigReadRepository,
) *AfterSalesQueryService {
	return &AfterSalesQueryService{
		repo:                           repo,
		readRepo:                       readRepo,
		searchRepo:                     searchRepo,
		supportTicketReadRepo:          supportTicketReadRepo,
		supportTicketSearchRepo:        supportTicketSearchRepo,
		supportTicketMessageSearchRepo: supportTicketMessageSearchRepo,
		configReadRepo:                 configReadRepo,
	}
}

func (q *AfterSalesQueryService) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	if q.searchRepo != nil {
		list, total, err := q.searchRepo.Search(ctx, query)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.List(ctx, query)
}

func (q *AfterSalesQueryService) GetDetails(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	if q.readRepo != nil {
		if cached, err := q.readRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	item, err := q.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, domain.ErrAfterSalesNotFound
	}
	if q.readRepo != nil {
		_ = q.readRepo.Save(ctx, item)
	}
	return item, nil
}

func (q *AfterSalesQueryService) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	if q.supportTicketReadRepo != nil {
		if cached, err := q.supportTicketReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	ticket, err := q.repo.GetSupportTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("support ticket not found")
	}
	if q.supportTicketReadRepo != nil {
		_ = q.supportTicketReadRepo.Save(ctx, ticket)
	}
	return ticket, nil
}

func (q *AfterSalesQueryService) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	offset := (page - 1) * pageSize
	if q.supportTicketSearchRepo != nil {
		list, total, err := q.supportTicketSearchRepo.Search(ctx, userID, status, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.ListSupportTickets(ctx, userID, status, page, pageSize)
}

func (q *AfterSalesQueryService) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	if q.supportTicketMessageSearchRepo != nil {
		list, _, err := q.supportTicketMessageSearchRepo.Search(ctx, ticketID, 0, 200)
		if err == nil {
			return list, nil
		}
	}
	return q.repo.ListSupportTicketMessages(ctx, ticketID)
}

func (q *AfterSalesQueryService) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	if q.configReadRepo != nil {
		if cached, err := q.configReadRepo.GetByKey(ctx, key); err == nil && cached != nil {
			return cached, nil
		}
	}
	cfg, err := q.repo.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}
	if cfg != nil && q.configReadRepo != nil {
		_ = q.configReadRepo.Save(ctx, cfg)
	}
	return cfg, nil
}
