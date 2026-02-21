package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	marketdatav1 "github.com/wyfcoding/financialtrading/go-api/marketdata/v1"
)

// PricingQueryService 处理读操作和计算。
type PricingQueryService struct {
	repo              domain.PricingRepository
	ruleReadRepo      domain.PricingRuleReadRepository
	historySearchRepo domain.PriceHistorySearchRepository
	marketDataCli     marketdatav1.MarketDataServiceClient
}

// NewPricingQueryService creates a new PricingQueryService instance.
func NewPricingQueryService(
	repo domain.PricingRepository,
	ruleReadRepo domain.PricingRuleReadRepository,
	historySearchRepo domain.PriceHistorySearchRepository,
) *PricingQueryService {
	return &PricingQueryService{
		repo:              repo,
		ruleReadRepo:      ruleReadRepo,
		historySearchRepo: historySearchRepo,
	}
}

func (q *PricingQueryService) SetMarketDataClient(cli marketdatav1.MarketDataServiceClient) {
	q.marketDataCli = cli
}

// CalculatePrice 根据定价规则计算商品或SKU的价格。
func (q *PricingQueryService) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	// 1. 优先尝试从本地仓储获取最新动态价格
	latest, err := q.repo.GetLatestDynamicPrice(ctx, skuID)
	if err == nil && latest != nil && latest.FinalPrice > 0 {
		return uint64(latest.FinalPrice), nil
	}

	// 2. 如果没有动态价格，回退到基础定价规则
	var rule *domain.PricingRule
	if q.ruleReadRepo != nil {
		if cached, cacheErr := q.ruleReadRepo.GetActive(ctx, productID, skuID); cacheErr == nil && cached != nil {
			rule = cached
		}
	}
	if rule == nil {
		rule, err = q.repo.GetActiveRule(ctx, productID, skuID)
		if err != nil {
			return 0, err
		}
		if rule != nil && q.ruleReadRepo != nil {
			_ = q.ruleReadRepo.Save(ctx, rule)
		}
	}
	if rule == nil {
		return 0, errors.New("no active pricing rule found")
	}

	price := rule.CalculatePrice(demand, competition)
	return price, nil
}

// ConvertPrice 将价格转换为目标币种 (Cross-Project Interaction)
func (q *PricingQueryService) ConvertPrice(ctx context.Context, amount uint64, baseCurrency, targetCurrency string) (float64, error) {
	if baseCurrency == targetCurrency {
		return float64(amount), nil
	}

	if q.marketDataCli == nil {
		return 0, errors.New("market data service client not initialized")
	}

	// 构造交易对代码，例如 "USD/CNY"
	symbol := baseCurrency + "/" + targetCurrency
	resp, err := q.marketDataCli.GetLatestQuote(ctx, &marketdatav1.GetLatestQuoteRequest{
		Symbol: symbol,
	})
	if err != nil {
		return 0, err
	}

	// 使用最新报价进行转换
	convertedAmount := float64(amount) * resp.LastPrice
	return convertedAmount, nil
}

// ListRules 获取定价规则列表。
func (q *PricingQueryService) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*domain.PricingRule, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRules(ctx, productID, offset, pageSize)
}

// ListHistory 获取价格历史记录列表。
func (q *PricingQueryService) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*domain.PriceHistory, int64, error) {
	offset := (page - 1) * pageSize
	if q.historySearchRepo != nil {
		list, total, err := q.historySearchRepo.Search(ctx, productID, skuID, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
	}
	return q.repo.ListHistory(ctx, productID, skuID, offset, pageSize)
}
