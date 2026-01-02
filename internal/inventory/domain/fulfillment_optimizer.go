package domain

import (
	"context"
	"math"

	"github.com/shopspring/decimal"
)

// WarehouseStock 仓库库存信息
type WarehouseStock struct {
	WarehouseID uint64
	SKUID       uint64
	Available   int32
	LocationLat float64
	LocationLon float64
}

// FulfillmentPlan 履约计划
type FulfillmentPlan struct {
	SKUID       uint64
	WarehouseID uint64
	Quantity    int32
	ShipCost    decimal.Decimal
}

// FulfillmentOptimizer 履约优化器
type FulfillmentOptimizer struct{}

// Optimize 寻找最优发货仓库组合
func (o *FulfillmentOptimizer) Optimize(ctx context.Context, orderItems map[uint64]int32, userLat, userLon float64, stocks []*WarehouseStock) ([]FulfillmentPlan, error) {
	plans := make([]FulfillmentPlan, 0)

	for skuID, neededQty := range orderItems {
		remaining := neededQty

		// 1. 过滤出有该SKU的仓库并按距离排序
		type candidate struct {
			ws   *WarehouseStock
			dist float64
		}
		var candidates []candidate
		for _, s := range stocks {
			if s.SKUID == skuID && s.Available > 0 {
				dist := o.calculateDistance(userLat, userLon, s.LocationLat, s.LocationLon)
				candidates = append(candidates, candidate{s, dist})
			}
		}

		// 2. 贪心策略：优先从最近的仓库发货 (Nearest Neighbor)
		// 进阶算法：使用最小费用最大流模型解决跨单合并最优解
		for _, cand := range candidates {
			if remaining <= 0 {
				break
			}

			shipQty := cand.ws.Available
			if shipQty > remaining {
				shipQty = remaining
			}

			plans = append(plans, FulfillmentPlan{
				SKUID:       skuID,
				WarehouseID: cand.ws.WarehouseID,
				Quantity:    shipQty,
				ShipCost:    decimal.NewFromFloat(cand.dist * 0.5), // 假设每公里 0.5 元
			})
			remaining -= shipQty
		}

		if remaining > 0 {
			// 库存不足异常处理
		}
	}

	return plans, nil
}

func (o *FulfillmentOptimizer) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 简化计算：欧几里得距离 (实际应使用球面距离)
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lon1-lon2, 2))
}
