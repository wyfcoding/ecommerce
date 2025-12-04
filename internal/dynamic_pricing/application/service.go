package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"     // 导入动态定价领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/repository" // 导入动态定价领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// DynamicPricingService 结构体定义了动态定价相关的应用服务。
// 它协调领域层和基础设施层，处理动态价格的计算、定价策略的管理等业务逻辑。
type DynamicPricingService struct {
	repo   repository.PricingRepository // 依赖PricingRepository接口，用于数据持久化操作。
	logger *slog.Logger                 // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewDynamicPricingService 创建并返回一个新的 DynamicPricingService 实例。
func NewDynamicPricingService(repo repository.PricingRepository, logger *slog.Logger) *DynamicPricingService {
	return &DynamicPricingService{
		repo:   repo,
		logger: logger,
	}
}

// CalculatePrice 计算给定SKU的动态价格。
// 它会根据定价策略和实时因素（库存、需求、竞品价格等）调整商品价格。
// ctx: 上下文。
// req: 包含计算价格所需所有输入参数的结构体。
// 返回计算出的DynamicPrice实体和可能发生的错误。
func (s *DynamicPricingService) CalculatePrice(ctx context.Context, req *entity.PricingRequest) (*entity.DynamicPrice, error) {
	// 1. 获取定价策略。
	strategy, err := s.repo.GetPricingStrategy(ctx, req.SKUID)
	if err != nil {
		// 如果没有找到特定策略，使用默认的固定价格策略。
		strategy = &entity.PricingStrategy{
			StrategyType: "fixed",
			MinPrice:     req.BasePrice,
			MaxPrice:     req.BasePrice,
		}
	}

	// 2. 计算各项影响价格的因子。
	// 库存因子：库存越低价格越高，库存越高价格越低。
	inventoryFactor := 1.0
	if req.TotalStock > 0 {
		ratio := float64(req.CurrentStock) / float64(req.TotalStock)
		if ratio < 0.2 { // 低于20%库存，价格上调10%。
			inventoryFactor = 1.1
		} else if ratio > 0.8 { // 高于80%库存，价格下调10%。
			inventoryFactor = 0.9
		}
	}

	// 需求因子：需求越高价格越高，需求越低价格越低。
	demandFactor := 1.0
	if req.AverageDailyDemand > 0 { // 避免除以零。
		ratio := float64(req.DailyDemand) / float64(req.AverageDailyDemand)
		if ratio > 1.5 { // 实际需求高于平均需求1.5倍，价格上调10%。
			demandFactor = 1.1
		} else if ratio < 0.5 { // 实际需求低于平均需求0.5倍，价格下调10%。
			demandFactor = 0.9
		}
	}

	// 竞品价格因子：根据竞品价格调整自身价格。
	competitorFactor := 1.0
	if req.CompetitorPrice > 0 {
		if req.CompetitorPrice < req.BasePrice { // 如果竞品价格低于我方基础价格，我方适当降价。
			competitorFactor = 0.95
		}
	}

	// 3. 计算最终价格。
	adjustment := inventoryFactor * demandFactor * competitorFactor // 综合所有因子。
	finalPrice := int64(float64(req.BasePrice) * adjustment)

	// 应用价格约束（最小价格和最大价格）。
	if strategy.MinPrice > 0 && finalPrice < strategy.MinPrice {
		finalPrice = strategy.MinPrice
	}
	if strategy.MaxPrice > 0 && finalPrice > strategy.MaxPrice {
		finalPrice = strategy.MaxPrice
	}

	// 创建DynamicPrice实体。
	price := &entity.DynamicPrice{
		SKUID:            req.SKUID,
		BasePrice:        req.BasePrice,
		FinalPrice:       finalPrice,
		PriceAdjustment:  adjustment,
		InventoryFactor:  inventoryFactor,
		DemandFactor:     demandFactor,
		CompetitorFactor: competitorFactor,
		EffectiveTime:    time.Now(),
		ExpiryTime:       time.Now().Add(24 * time.Hour), // 价格有效期默认为24小时。
	}

	// 通过仓储接口保存动态价格记录。
	if err := s.repo.SaveDynamicPrice(ctx, price); err != nil {
		s.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "final_price", finalPrice)

	return price, nil
}

// GetLatestPrice 获取指定SKU的最新动态价格。
// ctx: 上下文。
// skuID: 待查询的SKU ID。
// 返回最新动态价格实体和可能发生的错误。
func (s *DynamicPricingService) GetLatestPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error) {
	return s.repo.GetLatestDynamicPrice(ctx, skuID)
}

// SaveStrategy 保存（创建或更新）一个定价策略。
// ctx: 上下文。
// strategy: 待保存的PricingStrategy实体。
// 返回可能发生的错误。
func (s *DynamicPricingService) SaveStrategy(ctx context.Context, strategy *entity.PricingStrategy) error {
	return s.repo.SavePricingStrategy(ctx, strategy)
}

// ListStrategies 获取定价策略列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回定价策略列表、总数和可能发生的错误。
func (s *DynamicPricingService) ListStrategies(ctx context.Context, page, pageSize int) ([]*entity.PricingStrategy, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPricingStrategies(ctx, offset, pageSize)
}
