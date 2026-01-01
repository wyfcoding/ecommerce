package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	marketdatav1 "github.com/wyfcoding/financialtrading/goapi/marketdata/v1"
)

// PricingQuery 处理读操作和计算。
type PricingQuery struct {
	repo          domain.PricingRepository
	marketDataCli marketdatav1.MarketDataServiceClient
}

// NewPricingQuery creates a new PricingQuery instance.
func NewPricingQuery(repo domain.PricingRepository) *PricingQuery {
	return &PricingQuery{
		repo: repo,
	}
}

func (q *PricingQuery) SetMarketDataClient(cli marketdatav1.MarketDataServiceClient) {
	q.marketDataCli = cli
}

// CalculatePrice 根据定价规则计算商品或SKU的价格。
func (q *PricingQuery) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	rule, err := q.repo.GetActiveRule(ctx, productID, skuID)
	if err != nil {
		return 0, err
	}
	if rule == nil {
		return 0, errors.New("no active pricing rule found")
	}

	price := rule.CalculatePrice(demand, competition)
	return price, nil
}

// ConvertPrice 将价格转换为目标币种 (Cross-Project Interaction)
func (q *PricingQuery) ConvertPrice(ctx context.Context, amount uint64, baseCurrency, targetCurrency string) (float64, error) {
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
func (q *PricingQuery) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*domain.PricingRule, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRules(ctx, productID, offset, pageSize)
}

// ListHistory 获取价格历史记录列表。
func (q *PricingQuery) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*domain.PriceHistory, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListHistory(ctx, productID, skuID, offset, pageSize)
}
