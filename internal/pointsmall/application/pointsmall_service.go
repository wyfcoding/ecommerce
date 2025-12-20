package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

// PointsmallService 作为积分商城操作的门面。
type PointsmallService struct {
	manager *PointsManager
	query   *PointsQuery
}

// NewPointsmallService creates a new PointsmallService facade.
func NewPointsmallService(manager *PointsManager, query *PointsQuery) *PointsmallService {
	return &PointsmallService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *PointsmallService) CreateProduct(ctx context.Context, product *domain.PointsProduct) error {
	return s.manager.CreateProduct(ctx, product)
}

func (s *PointsmallService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*domain.PointsOrder, error) {
	return s.manager.ExchangeProduct(ctx, userID, productID, quantity, address, phone, receiver)
}

func (s *PointsmallService) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	return s.manager.AddPoints(ctx, userID, points, description, refID)
}

// --- 读操作（委托给 Query）---

func (s *PointsmallService) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*domain.PointsProduct, int64, error) {
	return s.query.ListProducts(ctx, status, page, pageSize)
}

func (s *PointsmallService) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	return s.query.GetProduct(ctx, id)
}

func (s *PointsmallService) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	return s.query.GetAccount(ctx, userID)
}

func (s *PointsmallService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.PointsOrder, int64, error) {
	return s.query.ListOrders(ctx, userID, status, page, pageSize)
}
