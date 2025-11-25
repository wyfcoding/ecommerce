package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"
)

// FlashSaleRepository 秒杀仓储接口
type FlashSaleRepository interface {
	// 活动管理
	SaveFlashsale(ctx context.Context, flashsale *entity.Flashsale) error
	GetFlashsale(ctx context.Context, id uint64) (*entity.Flashsale, error)
	ListFlashsales(ctx context.Context, status *entity.FlashsaleStatus, offset, limit int) ([]*entity.Flashsale, int64, error)
	UpdateStock(ctx context.Context, id uint64, quantity int32) error

	// 订单管理
	SaveOrder(ctx context.Context, order *entity.FlashsaleOrder) error
	GetOrder(ctx context.Context, id uint64) (*entity.FlashsaleOrder, error)
	GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*entity.FlashsaleOrder, error)
	CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error)
}
