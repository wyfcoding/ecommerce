package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashsaleService 结构体定义了秒杀活动相关的应用服务。
// 它作为一个外观（Facade），组合了 FlashsaleManager 和 FlashsaleQuery。
type FlashsaleService struct {
	manager *FlashsaleManager
	query   *FlashsaleQuery
}

// NewFlashsaleService 创建秒杀服务门面实例。
func NewFlashsaleService(manager *FlashsaleManager, query *FlashsaleQuery) *FlashsaleService {
	return &FlashsaleService{
		manager: manager,
		query:   query,
	}
}

// CreateFlashsale 创建一个新的秒杀活动。
func (s *FlashsaleService) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*domain.Flashsale, error) {
	return s.manager.CreateFlashsale(ctx, name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
}

// GetFlashsale 获取指定ID的秒杀活动详情。
func (s *FlashsaleService) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	return s.query.GetFlashsale(ctx, id)
}

// ListFlashsales 列出秒杀活动（按状态筛选）。
func (s *FlashsaleService) ListFlashsales(ctx context.Context, status *domain.FlashsaleStatus, page, pageSize int) ([]*domain.Flashsale, int64, error) {
	return s.query.ListFlashsales(ctx, status, page, pageSize)
}

// PlaceOrder 秒杀下单（包含库存预扣逻辑）。
func (s *FlashsaleService) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*domain.FlashsaleOrder, error) {
	return s.manager.PlaceOrder(ctx, userID, flashsaleID, quantity)
}

// CancelOrder 取消秒杀订单。
func (s *FlashsaleService) CancelOrder(ctx context.Context, orderID uint64) error {
	return s.manager.CancelOrder(ctx, orderID)
}

// GetOrder 获取秒杀订单详情。
func (s *FlashsaleService) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	return s.query.GetOrder(ctx, id)
}

// SaveOrder 保存秒杀订单状态。
func (s *FlashsaleService) SaveOrder(ctx context.Context, order *domain.FlashsaleOrder) error {
	return s.manager.SaveOrder(ctx, order)
}

// UpdateStock 更新秒杀活动的库存数据。
func (s *FlashsaleService) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return s.manager.UpdateStock(ctx, id, quantity)
}

// CountUserBought 获取指定用户在某次秒杀活动中已购买的商品数量。
func (s *FlashsaleService) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	return s.query.CountUserBought(ctx, userID, flashsaleID)
}
