package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashsaleQuery 负责处理 Flashsale 相关的读操作和查询逻辑。
type FlashsaleQuery struct {
	repo domain.FlashSaleRepository
}

// NewFlashsaleQuery 负责处理 NewFlashsale 相关的读操作和查询逻辑。
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
