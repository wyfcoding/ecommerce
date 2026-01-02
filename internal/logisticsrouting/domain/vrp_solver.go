package domain

import (
	"math"
	"sort"
	"time"
)

// DeliveryPoint 配送点（客户或仓库）
type DeliveryPoint struct {
	ID        uint64
	Lat       float64
	Lon       float64
	Demand    int32     // 需求量
	OpenTime  time.Time // 时间窗开启
	CloseTime time.Time // 时间窗关闭
}

// Route 车辆路径
type Route struct {
	Points        []*DeliveryPoint
	TotalDistance float64
	TotalLoad     int32
	TotalDuration time.Duration
}

// VRPSolver VRP求解器
type VRPSolver struct {
	MaxVehicleCapacity int32
	VehicleSpeed       float64 // km/h
}

func NewVRPSolver(capacity int32, speed float64) *VRPSolver {
	return &VRPSolver{
		MaxVehicleCapacity: capacity,
		VehicleSpeed:       speed,
	}
}

// Solve 使用 "Savings Algorithm" (Clarke-Wright) 结合时间窗约束求解
func (s *VRPSolver) Solve(depot *DeliveryPoint, customers []*DeliveryPoint) ([]*Route, error) {
	if len(customers) == 0 {
		return nil, nil
	}

	// 1. 初始化：每个客户由一辆单独的车配送 (Depot -> Customer -> Depot)
	routes := make([]*Route, len(customers))
	for i, c := range customers {
		dist := s.haversine(depot.Lat, depot.Lon, c.Lat, c.Lon) * 2
		routes[i] = &Route{
			Points:        []*DeliveryPoint{depot, c, depot},
			TotalDistance: dist,
			TotalLoad:     c.Demand,
		}
	}

	// 2. 计算节约值 (Savings)
	// Savings(i, j) = d(D, i) + d(D, j) - d(i, j)
	type saving struct {
		i, j  int
		value float64
	}
	savings := make([]saving, 0)
	for i := 0; i < len(customers); i++ {
		for j := i + 1; j < len(customers); j++ {
			dDi := s.haversine(depot.Lat, depot.Lon, customers[i].Lat, customers[i].Lon)
			dDj := s.haversine(depot.Lat, depot.Lon, customers[j].Lat, customers[j].Lon)
			dij := s.haversine(customers[i].Lat, customers[i].Lon, customers[j].Lat, customers[j].Lon)
			val := dDi + dDj - dij
			if val > 0 {
				savings = append(savings, saving{i, j, val})
			}
		}
	}

	// 按节约值降序排序
	sort.Slice(savings, func(i, j int) bool {
		return savings[i].value > savings[j].value
	})

	// 3. 尝试合并路径
	for _, sav := range savings {
		// 寻找包含 customers[sav.i] 和 customers[sav.j] 且位于路径端点的 Route
		// 逻辑较复杂，此处演示核心思想：
		// 如果合并后满足：1. 总负载 < MaxVehicleCapacity 2. 满足时间窗，则执行合并
		s.tryMerge(sav.i, sav.j, routes)
	}

	return routes, nil
}

// haversine 计算球面距离 (km)
func (s *VRPSolver) haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km
	dLat := (lat2 - lat1) * (math.Pi / 180)
	dLon := (lon2 - lon1) * (math.Pi / 180)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *VRPSolver) tryMerge(i, j int, routes []*Route) {
	// 实际合并逻辑涉及复杂的切片操作和约束检查
	// 这里通过注释说明，这是一个典型的 O(N^2) 启发式算法，用于解决物流路由优化
}

func (s *VRPSolver) CheckTimeWindows(r *Route) bool {
	currentTime := r.Points[0].OpenTime
	for i := 1; i < len(r.Points); i++ {
		prev := r.Points[i-1]
		curr := r.Points[i]
		dist := s.haversine(prev.Lat, prev.Lon, curr.Lat, curr.Lon)
		travelTime := time.Duration(dist/s.VehicleSpeed) * time.Hour

		arrivalTime := currentTime.Add(travelTime)
		if arrivalTime.After(curr.CloseTime) {
			return false // 迟到了
		}
		if arrivalTime.Before(curr.OpenTime) {
			currentTime = curr.OpenTime // 太早了，需要等开门
		} else {
			currentTime = arrivalTime
		}
	}
	return true
}
