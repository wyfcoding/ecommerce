package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashSaleQueryService 负责处理 Flashsale 相关的读操作和查询逻辑。
type FlashSaleQueryService struct {
	repo            domain.FlashSaleRepository
	flashsaleRead   domain.FlashsaleReadRepository
	orderRead       domain.FlashsaleOrderReadRepository
	flashsaleSearch domain.FlashsaleSearchRepository
	orderSearch     domain.FlashsaleOrderSearchRepository
	logger          *slog.Logger
}

// NewFlashSaleQueryService 负责处理 Flashsale 相关的读操作和查询逻辑。
func NewFlashSaleQueryService(
	repo domain.FlashSaleRepository,
	flashsaleRead domain.FlashsaleReadRepository,
	orderRead domain.FlashsaleOrderReadRepository,
	flashsaleSearch domain.FlashsaleSearchRepository,
	orderSearch domain.FlashsaleOrderSearchRepository,
	logger *slog.Logger,
) *FlashSaleQueryService {
	return &FlashSaleQueryService{
		repo:            repo,
		flashsaleRead:   flashsaleRead,
		orderRead:       orderRead,
		flashsaleSearch: flashsaleSearch,
		orderSearch:     orderSearch,
		logger:          logger,
	}
}

// GetFlashsale 获取指定ID的秒杀活动详情。
func (q *FlashSaleQueryService) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	if q.flashsaleRead != nil {
		if cached, err := q.flashsaleRead.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	flashsale, err := q.repo.GetFlashsale(ctx, id)
	if err != nil {
		return nil, err
	}
	if flashsale != nil && q.flashsaleRead != nil {
		_ = q.flashsaleRead.Save(ctx, flashsale)
	}
	return flashsale, nil
}

// ListFlashsales 获取秒杀活动列表。
func (q *FlashSaleQueryService) ListFlashsales(ctx context.Context, status *domain.FlashsaleStatus, page, pageSize int) ([]*domain.Flashsale, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query := &domain.FlashsaleQuery{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	}
	offset := (page - 1) * pageSize
	if q.flashsaleSearch != nil {
		list, total, err := q.flashsaleSearch.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "flashsale search fallback to mysql", "error", err)
		}
	}
	return q.repo.ListFlashsales(ctx, query)
}

func (q *FlashSaleQueryService) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	if q.orderRead != nil {
		if cached, err := q.orderRead.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	order, err := q.repo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order != nil && q.orderRead != nil {
		_ = q.orderRead.Save(ctx, order)
	}
	return order, nil
}

func (q *FlashSaleQueryService) ListOrders(ctx context.Context, query *domain.FlashsaleOrderQuery) ([]*domain.FlashsaleOrder, int64, error) {
	if query == nil {
		query = &domain.FlashsaleOrderQuery{}
	}
	page := query.Page
	pageSize := query.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	query.Page = page
	query.PageSize = pageSize
	offset := (page - 1) * pageSize
	if q.orderSearch != nil {
		list, total, err := q.orderSearch.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "flashsale order search fallback to mysql", "error", err)
		}
	}
	return q.repo.ListOrders(ctx, query)
}

func (q *FlashSaleQueryService) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	return q.repo.CountUserBought(ctx, userID, flashsaleID)
}
