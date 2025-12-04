package application

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity"     // 导入定价领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/repository" // 导入定价领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// PricingService 结构体定义了动态定价相关的应用服务。
// 它协调领域层和基础设施层，处理定价规则的创建、价格计算以及价格历史记录等业务逻辑。
type PricingService struct {
	repo   repository.PricingRepository // 依赖PricingRepository接口，用于数据持久化操作。
	logger *slog.Logger                 // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewPricingService 创建并返回一个新的 PricingService 实例。
func NewPricingService(repo repository.PricingRepository, logger *slog.Logger) *PricingService {
	return &PricingService{
		repo:   repo,
		logger: logger,
	}
}

// CreateRule 创建一个新的定价规则。
// ctx: 上下文。
// rule: 待创建的PricingRule实体。
// 返回可能发生的错误。
func (s *PricingService) CreateRule(ctx context.Context, rule *entity.PricingRule) error {
	if err := s.repo.SaveRule(ctx, rule); err != nil {
		s.logger.ErrorContext(ctx, "failed to create pricing rule", "rule_id", rule.ID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "pricing rule created successfully", "rule_id", rule.ID)
	return nil
}

// CalculatePrice 根据定价规则计算商品或SKU的价格。
// ctx: 上下文。
// productID: 商品ID。
// skuID: SKU ID。
// demand: 当前需求系数。
// competition: 当前竞争系数。
// 返回计算后的价格（单位：分）和可能发生的错误。
func (s *PricingService) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	// 获取商品或SKU的活跃定价规则。
	rule, err := s.repo.GetActiveRule(ctx, productID, skuID)
	if err != nil {
		return 0, err
	}
	if rule == nil {
		// 如果没有找到活跃规则，返回0或默认价格。
		// 在此示例中，0表示未计算出特定规则下的价格。
		return 0, errors.New("no active pricing rule found")
	}

	// 调用定价规则实体的方法来计算价格。
	price := rule.CalculatePrice(demand, competition)
	return price, nil
}

// RecordHistory 记录价格变动历史。
// ctx: 上下文。
// productID: 商品ID。
// skuID: SKU ID。
// price: 新的价格。
// oldPrice: 旧的价格。
// reason: 价格变动原因。
// 返回可能发生的错误。
func (s *PricingService) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
	var changeRate float64
	if oldPrice > 0 {
		// 计算价格变动率。
		changeRate = float64(price-oldPrice) / float64(oldPrice) * 100
	}

	history := &entity.PriceHistory{
		ProductID:  productID,
		SkuID:      skuID,
		Price:      price,
		OldPrice:   oldPrice,
		ChangeRate: changeRate,
		Reason:     reason,
	}
	// 通过仓储接口保存价格历史记录。
	if err := s.repo.SaveHistory(ctx, history); err != nil {
		s.logger.ErrorContext(ctx, "failed to record price history", "product_id", productID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "price history recorded successfully", "product_id", productID, "sku_id", skuID, "price", price)
	return nil
}

// ListRules 获取定价规则列表。
// ctx: 上下文。
// productID: 筛选规则的商品ID。
// page, pageSize: 分页参数。
// 返回定价规则列表、总数和可能发生的错误。
func (s *PricingService) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*entity.PricingRule, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRules(ctx, productID, offset, pageSize)
}

// ListHistory 获取价格历史记录列表。
// ctx: 上下文。
// productID: 筛选历史记录的商品ID。
// skuID: 筛选历史记录的SKU ID。
// page, pageSize: 分页参数。
// 返回价格历史记录列表、总数和可能发生的错误。
func (s *PricingService) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*entity.PriceHistory, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListHistory(ctx, productID, skuID, offset, pageSize)
}
