// 变更说明：
// 促销引擎应用层的 Query (读) 服务，贯彻 CQRS 架构模式。
// 直接将繁重的读取压力分配到从库（Read Replica）或 Redis 中，不干预写模型的性能。
package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/promotion/domain"
)

type PromotionQueryService struct {
	readRepo domain.PromotionReadRepository
	cache    domain.PromotionCache // 只读缓存
	logger   *slog.Logger
}

func NewPromotionQueryService(readRepo domain.PromotionReadRepository, cache domain.PromotionCache, logger *slog.Logger) *PromotionQueryService {
	return &PromotionQueryService{
		readRepo: readRepo,
		cache:    cache,
		logger:   logger,
	}
}

// GetProductActivePromotions 查询指定商品详情页能够展示的活动“横幅”（Banner）。
// 使用缓存穿透保护：优先读 Redis。
func (s *PromotionQueryService) GetProductActivePromotions(ctx context.Context, productID uint64) ([]*domain.Promotion, error) {
	now := time.Now()

	// 1. 尝试从高速缓存读取商品的关联促销ID映射
	promos, err := s.cache.GetActiveByProduct(ctx, productID)
	if err == nil && len(promos) > 0 {
		return promos, nil
	}

	// 2. 缓存 Miss：下沉到 Read 节点查 DB 并进行装配
	dbPromos, err := s.readRepo.ListActiveByProduct(ctx, productID, now)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to query active promotions from DB", "product_id", productID, "error", err)
		return nil, err
	}

	// 3. 异步回填缓存 (防御缓存击穿可使用 singleflight，这里简写回填)
	if len(dbPromos) > 0 {
		_ = s.cache.SetProductPromotions(ctx, productID, dbPromos)
	}

	return dbPromos, nil
}
