// 变更说明：
// 从 pkg/algos/optimization/warehouse_allocator.go 迁移。
// 实现了订单商品仓库分配优化，综合考虑距离、库存量和发货成本。
package domain

import (
	"slices"

	algomath "github.com/wyfcoding/pkg/algos/math"
)

type WarehouseInfo struct {
	Lat      float64
	Lon      float64
	ID       uint64
	ShipCost int64
	Stock    int32
	Priority int
}

type OrderItem struct {
	SkuID    uint64
	Quantity int32
}

type AllocationResult struct {
	Items       []OrderItem
	Distance    float64
	WarehouseID uint64
	TotalCost   int64
}

type WarehouseAllocator struct{}

func NewWarehouseAllocator() *WarehouseAllocator {
	return &WarehouseAllocator{}
}

type warehouseScore struct {
	score       float64
	distance    float64
	warehouseID uint64
}

func (wa *WarehouseAllocator) AllocateOptimal(
	userLat, userLon float64,
	items []OrderItem,
	warehouses map[uint64]map[uint64]*WarehouseInfo,
) []AllocationResult {
	results := make([]AllocationResult, 0)
	remaining := make(map[uint64]int32)
	for _, item := range items {
		remaining[item.SkuID] = item.Quantity
	}

	scores := wa.calculateScores(userLat, userLon, remaining, warehouses)
	slices.SortFunc(scores, func(a, b warehouseScore) int {
		if a.score > b.score {
			return -1
		}
		return 1
	})

	for _, ws := range scores {
		if len(remaining) == 0 {
			break
		}
		res := wa.allocateFromWarehouse(ws.warehouseID, ws.distance, remaining, warehouses)
		if len(res.Items) > 0 {
			results = append(results, res)
		}
	}
	return results
}

func (wa *WarehouseAllocator) calculateScores(
	lat, lon float64,
	rem map[uint64]int32,
	warehouses map[uint64]map[uint64]*WarehouseInfo,
) []warehouseScore {
	scores := make([]warehouseScore, 0)
	for wID, skuMap := range warehouses {
		var wInfo *WarehouseInfo
		for _, info := range skuMap {
			wInfo = info
			break
		}
		if wInfo == nil {
			continue
		}

		dist := algomath.HaversineDistance(lat, lon, wInfo.Lat, wInfo.Lon)
		var totalStock int32
		var coverage int
		for sID := range rem {
			if info, ok := skuMap[sID]; ok {
				totalStock += info.Stock
				coverage++
			}
		}

		// 综合分：距离 (40%) + 库存覆盖 (30%) + 总库存量 (30%)
		sScore := float64(coverage) / float64(len(rem))
		dScore := 1.0 / (1.0 + dist/1000.0)
		final := 0.4*dScore + 0.3*sScore + 0.3*(float64(totalStock)/1000.0)
		scores = append(scores, warehouseScore{warehouseID: wID, distance: dist, score: final})
	}
	return scores
}

func (wa *WarehouseAllocator) allocateFromWarehouse(
	wID uint64,
	dist float64,
	rem map[uint64]int32,
	warehouses map[uint64]map[uint64]*WarehouseInfo,
) AllocationResult {
	skuMap := warehouses[wID]
	allocated := make([]OrderItem, 0)
	var cost int64
	for sID, need := range rem {
		if info, ok := skuMap[sID]; ok && info.Stock > 0 {
			qty := min(info.Stock, need)
			allocated = append(allocated, OrderItem{SkuID: sID, Quantity: qty})
			cost += info.ShipCost * int64(qty)
			rem[sID] -= qty
			if rem[sID] <= 0 {
				delete(rem, sID)
			}
		}
	}
	return AllocationResult{WarehouseID: wID, Items: allocated, TotalCost: cost, Distance: dist}
}
