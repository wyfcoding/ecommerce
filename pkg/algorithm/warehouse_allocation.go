package algorithm

import (
	"math"
	"sort"
)

// WarehouseInfo 仓库信息
type WarehouseInfo struct {
	ID        uint64
	Lat       float64 // 纬度
	Lon       float64 // 经度
	Stock     int32   // 库存
	Priority  int     // 优先级
	ShipCost  int64   // 配送成本（分）
}

// OrderItem 订单商品
type OrderItem struct {
	SkuID    uint64
	Quantity int32
}

// AllocationResult 分配结果
type AllocationResult struct {
	WarehouseID uint64
	Items       []OrderItem
	TotalCost   int64
	Distance    float64
}

// WarehouseAllocator 仓库分配器
type WarehouseAllocator struct{}

// NewWarehouseAllocator 创建仓库分配器
func NewWarehouseAllocator() *WarehouseAllocator {
	return &WarehouseAllocator{}
}

// AllocateOptimal 最优仓库分配（考虑距离、库存、成本）
// 策略：优先选择距离最近且库存充足的仓库，如果单仓库不足则拆分订单
func (wa *WarehouseAllocator) AllocateOptimal(
	userLat, userLon float64,
	items []OrderItem,
	warehouses map[uint64]map[uint64]*WarehouseInfo, // warehouseID -> skuID -> info
) []AllocationResult {
	
	results := make([]AllocationResult, 0)
	remainingItems := make(map[uint64]int32) // skuID -> quantity
	
	for _, item := range items {
		remainingItems[item.SkuID] = item.Quantity
	}

	// 按距离排序仓库
	type warehouseScore struct {
		warehouseID uint64
		distance    float64
		totalStock  int32
		avgCost     int64
		score       float64
	}

	scores := make([]warehouseScore, 0)
	
	for warehouseID, skuMap := range warehouses {
		if len(skuMap) == 0 {
			continue
		}

		// 获取第一个SKU的仓库信息（假设同一仓库的位置相同）
		var warehouseInfo *WarehouseInfo
		for _, info := range skuMap {
			warehouseInfo = info
			break
		}

		distance := haversineDistance(userLat, userLon, warehouseInfo.Lat, warehouseInfo.Lon)
		
		// 计算该仓库可满足的总库存和平均成本
		var totalStock int32
		var totalCost int64
		var count int
		
		for skuID, qty := range remainingItems {
			if info, exists := skuMap[skuID]; exists {
				totalStock += info.Stock
				totalCost += info.ShipCost
				count++
			}
		}

		avgCost := int64(0)
		if count > 0 {
			avgCost = totalCost / int64(count)
		}

		// 综合评分：距离权重0.4，库存权重0.3，成本权重0.3
		// 距离越近越好，库存越多越好，成本越低越好
		distanceScore := 1.0 / (1.0 + distance/1000.0) // 归一化距离
		stockScore := float64(totalStock) / 1000.0
		costScore := 1.0 / (1.0 + float64(avgCost)/100.0)
		
		score := 0.4*distanceScore + 0.3*stockScore + 0.3*costScore

		scores = append(scores, warehouseScore{
			warehouseID: warehouseID,
			distance:    distance,
			totalStock:  totalStock,
			avgCost:     avgCost,
			score:       score,
		})
	}

	// 按评分降序排序
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 贪心分配：从评分最高的仓库开始分配
	for _, ws := range scores {
		if len(remainingItems) == 0 {
			break
		}

		warehouseID := ws.warehouseID
		skuMap := warehouses[warehouseID]
		
		allocatedItems := make([]OrderItem, 0)
		totalCost := int64(0)

		for skuID, needQty := range remainingItems {
			if info, exists := skuMap[skuID]; exists && info.Stock > 0 {
				// 分配库存
				allocQty := needQty
				if info.Stock < needQty {
					allocQty = info.Stock
				}

				allocatedItems = append(allocatedItems, OrderItem{
					SkuID:    skuID,
					Quantity: allocQty,
				})

				totalCost += info.ShipCost * int64(allocQty)

				// 更新剩余需求
				remainingItems[skuID] -= allocQty
				if remainingItems[skuID] <= 0 {
					delete(remainingItems, skuID)
				}
			}
		}

		if len(allocatedItems) > 0 {
			results = append(results, AllocationResult{
				WarehouseID: warehouseID,
				Items:       allocatedItems,
				TotalCost:   totalCost,
				Distance:    ws.distance,
			})
		}
	}

	return results
}

// AllocateByDistance 按距离分配（最近优先）
func (wa *WarehouseAllocator) AllocateByDistance(
	userLat, userLon float64,
	items []OrderItem,
	warehouses map[uint64]map[uint64]*WarehouseInfo,
) []AllocationResult {
	
	type warehouseDist struct {
		warehouseID uint64
		distance    float64
	}

	distances := make([]warehouseDist, 0)

	for warehouseID, skuMap := range warehouses {
		if len(skuMap) == 0 {
			continue
		}

		var warehouseInfo *WarehouseInfo
		for _, info := range skuMap {
			warehouseInfo = info
			break
		}

		distance := haversineDistance(userLat, userLon, warehouseInfo.Lat, warehouseInfo.Lon)
		distances = append(distances, warehouseDist{warehouseID, distance})
	}

	sort.Slice(distances, func(i, j int) bool {
		return distances[i].distance < distances[j].distance
	})

	results := make([]AllocationResult, 0)
	remainingItems := make(map[uint64]int32)
	
	for _, item := range items {
		remainingItems[item.SkuID] = item.Quantity
	}

	for _, wd := range distances {
		if len(remainingItems) == 0 {
			break
		}

		warehouseID := wd.warehouseID
		skuMap := warehouses[warehouseID]
		
		allocatedItems := make([]OrderItem, 0)
		totalCost := int64(0)

		for skuID, needQty := range remainingItems {
			if info, exists := skuMap[skuID]; exists && info.Stock > 0 {
				allocQty := needQty
				if info.Stock < needQty {
					allocQty = info.Stock
				}

				allocatedItems = append(allocatedItems, OrderItem{
					SkuID:    skuID,
					Quantity: allocQty,
				})

				totalCost += info.ShipCost * int64(allocQty)

				remainingItems[skuID] -= allocQty
				if remainingItems[skuID] <= 0 {
					delete(remainingItems, skuID)
				}
			}
		}

		if len(allocatedItems) > 0 {
			results = append(results, AllocationResult{
				WarehouseID: warehouseID,
				Items:       allocatedItems,
				TotalCost:   totalCost,
				Distance:    wd.distance,
			})
		}
	}

	return results
}

// haversineDistance 计算两点间距离（米）
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0 // 地球半径（米）

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
