package domain

import (
	"context"
	"fmt"
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

// FulfillmentOptimizer 领域服务：负责根据用户地理位置与各分仓库存分布，计算最优的订单履约方案。
// 算法策略：当前采用 Nearest Neighbor（最近邻）贪心算法，旨在降低物流成本并提升配送时效。
type FulfillmentOptimizer struct{}

// Optimize 寻找最优发货仓库组合。
// 流程：按 SKU 遍历 -> 候选仓库筛选与距离排序 -> 贪心分配库存 -> 生成计划。
func (o *FulfillmentOptimizer) Optimize(ctx context.Context, orderItems map[uint64]int32, userLat, userLon float64, stocks []*WarehouseStock) ([]FulfillmentPlan, error) {
	plans := make([]FulfillmentPlan, 0)

	for skuID, neededQty := range orderItems {
		remaining := neededQty

		// 1. 筛选并计算距离
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

		// 2. 贪心策略：优先从物理距离最近的仓库抓取库存
		for _, cand := range candidates {
			if remaining <= 0 {
				break
			}

			shipQty := min(cand.ws.Available, remaining)

			plans = append(plans, FulfillmentPlan{
				SKUID:       skuID,
				WarehouseID: cand.ws.WarehouseID,
				Quantity:    shipQty,
				ShipCost:    decimal.NewFromFloat(cand.dist * 0.5), // 模拟成本计算
			})
			remaining -= shipQty
		}

		// 3. 严格校验：若遍历完所有候选仓后仍有缺口，判定为该 SKU 库存不足
		if remaining > 0 {
			return nil, fmt.Errorf("insufficient multi-warehouse inventory for SKU %d: lacking %d", skuID, remaining)
		}
	}

	return plans, nil
}

// calculateDistance 基于经纬度计算两点间的直线距离。
func (o *FulfillmentOptimizer) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	return math.Sqrt(math.Pow(lat1-lat2, 2) + math.Pow(lon1-lon2, 2))
}
