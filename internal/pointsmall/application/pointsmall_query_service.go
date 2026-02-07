package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

// PointsmallQueryService 处理积分商城的读操作。
type PointsmallQueryService struct {
	repo              domain.PointsRepository
	productReadRepo   domain.PointsProductReadRepository
	orderReadRepo     domain.PointsOrderReadRepository
	accountReadRepo   domain.PointsAccountReadRepository
	productSearchRepo domain.PointsProductSearchRepository
	orderSearchRepo   domain.PointsOrderSearchRepository
	logger            *slog.Logger
}

// NewPointsmallQueryService creates a new PointsmallQueryService instance.
func NewPointsmallQueryService(
	repo domain.PointsRepository,
	productReadRepo domain.PointsProductReadRepository,
	orderReadRepo domain.PointsOrderReadRepository,
	accountReadRepo domain.PointsAccountReadRepository,
	productSearchRepo domain.PointsProductSearchRepository,
	orderSearchRepo domain.PointsOrderSearchRepository,
	logger *slog.Logger,
) *PointsmallQueryService {
	return &PointsmallQueryService{
		repo:              repo,
		productReadRepo:   productReadRepo,
		orderReadRepo:     orderReadRepo,
		accountReadRepo:   accountReadRepo,
		productSearchRepo: productSearchRepo,
		orderSearchRepo:   orderSearchRepo,
		logger:            logger,
	}
}

// ListProducts 获取积分商品列表。
func (q *PointsmallQueryService) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*domain.PointsProduct, int64, error) {
	var prodStatus *domain.PointsProductStatus
	if status != nil {
		s := domain.PointsProductStatus(*status)
		prodStatus = &s
	}
	query := &domain.PointsProductQuery{
		Status:   prodStatus,
		Page:     page,
		PageSize: pageSize,
	}
	return q.SearchProducts(ctx, query)
}

// GetProduct 获取商品详情
func (q *PointsmallQueryService) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	if q.productReadRepo != nil {
		if cached, err := q.productReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	product, err := q.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	if product != nil && q.productReadRepo != nil {
		_ = q.productReadRepo.Save(ctx, product)
	}
	return product, nil
}

// GetAccount 获取用户积分账户信息。
func (q *PointsmallQueryService) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	if q.accountReadRepo != nil {
		if cached, err := q.accountReadRepo.GetByUserID(ctx, userID); err == nil && cached != nil {
			return cached, nil
		}
	}
	account, err := q.repo.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account = &domain.PointsAccount{UserID: userID}
		if err := q.repo.SaveAccount(ctx, account); err != nil {
			return nil, err
		}
	}
	if q.accountReadRepo != nil {
		_ = q.accountReadRepo.Save(ctx, account)
	}
	return account, nil
}

// ListOrders 获取积分订单列表。
func (q *PointsmallQueryService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.PointsOrder, int64, error) {
	var orderStatus *domain.PointsOrderStatus
	if status != nil {
		s := domain.PointsOrderStatus(*status)
		orderStatus = &s
	}
	query := &domain.PointsOrderQuery{
		UserID:   userID,
		Status:   orderStatus,
		Page:     page,
		PageSize: pageSize,
	}
	return q.SearchOrders(ctx, query)
}

// SearchProducts 搜索积分商品（优先 ES）。
func (q *PointsmallQueryService) SearchProducts(ctx context.Context, query *domain.PointsProductQuery) ([]*domain.PointsProduct, int64, error) {
	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.productSearchRepo != nil {
		list, total, err := q.productSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "product search fallback to mysql", "error", err)
	}
	return q.repo.ListProducts(ctx, query)
}

// SearchOrders 搜索积分订单（优先 ES）。
func (q *PointsmallQueryService) SearchOrders(ctx context.Context, query *domain.PointsOrderQuery) ([]*domain.PointsOrder, int64, error) {
	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.orderSearchRepo != nil {
		list, total, err := q.orderSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "order search fallback to mysql", "error", err)
	}
	return q.repo.ListOrders(ctx, query)
}
