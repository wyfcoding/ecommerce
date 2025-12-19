package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

type FlashsaleQuery struct {
	repo domain.FlashSaleRepository
}

func NewFlashsaleQuery(repo domain.FlashSaleRepository) *FlashsaleQuery {
	return &FlashsaleQuery{
		repo: repo,
	}
}

// GetFlashsale 获取指定ID的秒杀活动详情。
func (q *FlashsaleQuery) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	return q.repo.GetFlashsale(ctx, id)
}

// ListFlashsales 获取秒杀活动列表。
func (q *FlashsaleQuery) ListFlashsales(ctx context.Context, status *domain.FlashsaleStatus, page, pageSize int) ([]*domain.Flashsale, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListFlashsales(ctx, status, offset, pageSize)
}

func (q *FlashsaleQuery) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	return q.repo.GetOrder(ctx, id)
}

func (q *FlashsaleQuery) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	return q.repo.CountUserBought(ctx, userID, flashsaleID)
}
