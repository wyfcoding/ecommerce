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

// NewPointsmallService 创建积分商城服务门面实例。
func NewPointsmallService(manager *PointsManager, query *PointsQuery) *PointsmallService {
	return &PointsmallService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateProduct 在积分商城中上架一个新的积分商品。
func (s *PointsmallService) CreateProduct(ctx context.Context, product *domain.PointsProduct) error {
	return s.manager.CreateProduct(ctx, product)
}

// ExchangeProduct 用户使用积分兑换指定的商品。
func (s *PointsmallService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, address, phone, receiver string) (*domain.PointsOrder, error) {
	return s.manager.ExchangeProduct(ctx, userID, productID, quantity, address, phone, receiver)
}

// AddPoints 为用户手动增加积分（如活动奖励、补偿等）。
func (s *PointsmallService) AddPoints(ctx context.Context, userID uint64, points int64, description, refID string) error {
	return s.manager.AddPoints(ctx, userID, points, description, refID)
}

// --- 读操作（委托给 Query）---

// ListProducts 分页获取积分商品列表（可按状态筛选）。
func (s *PointsmallService) ListProducts(ctx context.Context, status *int, page, pageSize int) ([]*domain.PointsProduct, int64, error) {
	return s.query.ListProducts(ctx, status, page, pageSize)
}

// GetProduct 获取指定ID的积分商品详情。
func (s *PointsmallService) GetProduct(ctx context.Context, id uint64) (*domain.PointsProduct, error) {
	return s.query.GetProduct(ctx, id)
}

// GetAccount 获取指定用户的积分账户资产信息。
func (s *PointsmallService) GetAccount(ctx context.Context, userID uint64) (*domain.PointsAccount, error) {
	return s.query.GetAccount(ctx, userID)
}

// ListOrders 分页获取用户的积分兑换订单列表。
func (s *PointsmallService) ListOrders(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.PointsOrder, int64, error) {
	return s.query.ListOrders(ctx, userID, status, page, pageSize)
}
