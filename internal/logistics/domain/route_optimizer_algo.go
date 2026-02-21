// 变更说明：
// 从 pkg/algos/optimization/route_optimizer.go 迁移。
// 实现了配送路线优化算法，包括最近邻 (NN) 和 Clarke-Wright Savings 算法。
package domain

import (
	"math"
	"slices"

	algomath "github.com/wyfcoding/pkg/algos/math"
)

type Location struct {
	Lat    float64
	Lon    float64
	Demand float64
	ID     uint64
}

type Route struct {
	Locations []Location
	Distance  float64
}

type RouteOptimizer struct{}

func NewRouteOptimizer() *RouteOptimizer {
	return &RouteOptimizer{}
}

func (ro *RouteOptimizer) OptimizeRoute(start Location, destinations []Location) Route {
	destLen := len(destinations)
	if destLen == 0 {
		return Route{Locations: []Location{start}, Distance: 0}
	}

	visited := make(map[uint64]bool)
	route := make([]Location, 0, destLen+1)
	route = append(route, start)
	totalDistance := 0.0
	current := start

	for len(route) < destLen+1 {
		idx, dist := ro.findNearest(current, destinations, visited)
		if idx == -1 {
			break
		}
		nearest := destinations[idx]
		visited[nearest.ID] = true
		route = append(route, nearest)
		totalDistance += dist
		current = nearest
	}

	return Route{Locations: route, Distance: totalDistance}
}

func (ro *RouteOptimizer) findNearest(curr Location, dests []Location, visited map[uint64]bool) (int, float64) {
	minDist := math.MaxFloat64
	bestIdx := -1
	for idx, dest := range dests {
		if visited[dest.ID] {
			continue
		}
		dist := algomath.HaversineDistance(curr.Lat, curr.Lon, dest.Lat, dest.Lon)
		if dist < minDist {
			minDist = dist
			bestIdx = idx
		}
	}
	return bestIdx, minDist
}

type saving struct {
	val  float64
	i, j int
}

func (ro *RouteOptimizer) ClarkeWrightVRP(start Location, destinations []Location, capacity float64) []Route {
	n := len(destinations)
	if n == 0 {
		return nil
	}
	routes := make([][]int, n)
	for i := range n {
		routes[i] = []int{i}
	}

	savings := ro.calculateSavings(start, destinations)
	slices.SortFunc(savings, func(a, b saving) int {
		if a.val > b.val {
			return -1
		}
		return 1
	})

	mergedRoutes := ro.processMerging(destinations, routes, savings, capacity)
	return ro.buildFinalRoutes(start, destinations, mergedRoutes)
}

func (ro *RouteOptimizer) calculateSavings(start Location, dests []Location) []saving {
	n := len(dests)
	distToStart := make([]float64, n)
	for i := range n {
		distToStart[i] = algomath.HaversineDistance(start.Lat, start.Lon, dests[i].Lat, dests[i].Lon)
	}

	res := make([]saving, 0)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			distIJ := algomath.HaversineDistance(dests[i].Lat, dests[i].Lon, dests[j].Lat, dests[j].Lon)
			val := distToStart[i] + distToStart[j] - distIJ
			if val > 0 {
				res = append(res, saving{i: i, j: j, val: val})
			}
		}
	}
	return res
}

func (ro *RouteOptimizer) processMerging(dests []Location, curr [][]int, savings []saving, maxCapacity float64) [][]int {
	res := curr
	for _, s := range savings {
		r1, p1 := ro.findRouteIdx(res, s.i)
		r2, p2 := ro.findRouteIdx(res, s.j)
		if r1 == -1 || r2 == -1 || r1 == r2 {
			continue
		}
		if p1 == -1 || p2 == -1 {
			continue
		}

		if ro.getDemand(dests, res[r1])+ro.getDemand(dests, res[r2]) <= maxCapacity {
			merged := ro.merge(res[r1], res[r2], p1, p2)
			if merged != nil {
				res[r1] = merged
				res = append(res[:r2], res[r2+1:]...)
			}
		}
	}
	return res
}

func (ro *RouteOptimizer) findRouteIdx(routes [][]int, nodeIdx int) (int, int) {
	for rIdx, r := range routes {
		for pIdx, node := range r {
			if node == nodeIdx {
				if pIdx == 0 {
					return rIdx, 0
				}
				if pIdx == len(r)-1 {
					return rIdx, 1
				}
				return rIdx, -1
			}
		}
	}
	return -1, -1
}

func (ro *RouteOptimizer) getDemand(dests []Location, route []int) float64 {
	var d float64
	for _, idx := range route {
		d += dests[idx].Demand
	}
	return d
}

func (ro *RouteOptimizer) merge(r1, r2 []int, p1, p2 int) []int {
	if p1 == 1 && p2 == 0 {
		return append(r1, r2...)
	}
	if p1 == 0 && p2 == 1 {
		return append(r2, r1...)
	}
	if p1 == 0 && p2 == 0 {
		res := make([]int, 0, len(r1)+len(r2))
		for i := len(r1) - 1; i >= 0; i-- {
			res = append(res, r1[i])
		}
		return append(res, r2...)
	}
	if p1 == 1 && p2 == 1 {
		res := make([]int, 0, len(r1)+len(r2))
		res = append(res, r1...)
		for i := len(r2) - 1; i >= 0; i-- {
			res = append(res, r2[i])
		}
		return res
	}
	return nil
}

func (ro *RouteOptimizer) buildFinalRoutes(start Location, dests []Location, routes [][]int) []Route {
	final := make([]Route, len(routes))
	for i, nodes := range routes {
		locs := make([]Location, 0, len(nodes)+1)
		locs = append(locs, start)
		var dist float64
		curr := start
		for _, nIdx := range nodes {
			d := dests[nIdx]
			locs = append(locs, d)
			dist += algomath.HaversineDistance(curr.Lat, curr.Lon, d.Lat, d.Lon)
			curr = d
		}
		dist += algomath.HaversineDistance(curr.Lat, curr.Lon, start.Lat, start.Lon)
		final[i] = Route{Locations: locs, Distance: dist}
	}
	return final
}
