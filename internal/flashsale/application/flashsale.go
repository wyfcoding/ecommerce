package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashSale 门面服务，整合 CommandService 和 Query。
type FlashSale struct {
	command *FlashSaleCommandService
	query   *FlashSaleQuery
}

// NewFlashSale 构造函数。
func NewFlashSale(command *FlashSaleCommandService, query *FlashSaleQuery) *FlashSale {
	return &FlashSale{
		command: command,
		query:   query,
	}
}

// --- Commands (Writes) ---

func (s *FlashSale) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*domain.Flashsale, error) {
	return s.command.CreateFlashsale(ctx, name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime)
}

func (s *FlashSale) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*domain.FlashsaleOrder, error) {
	return s.command.PlaceOrder(ctx, userID, flashsaleID, quantity)
}

func (s *FlashSale) CancelOrder(ctx context.Context, orderID uint64) error {
	return s.command.CancelOrder(ctx, orderID)
}

func (s *FlashSale) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return s.command.UpdateStock(ctx, id, quantity)
}

// --- Query (Reads) ---

func (s *FlashSale) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	return s.query.GetFlashsale(ctx, id)
}

func (s *FlashSale) ListFlashsales(ctx context.Context, status *domain.FlashsaleStatus, page, pageSize int) ([]*domain.Flashsale, int64, error) {
	return s.query.ListFlashsales(ctx, status, page, pageSize)
}

func (s *FlashSale) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	return s.query.GetOrder(ctx, id)
}

func (s *FlashSale) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	return s.query.CountUserBought(ctx, userID, flashsaleID)
}
