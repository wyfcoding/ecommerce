package algorithm

import (
	"math"
	"sync"
)

// ============================================================================
// 4. 网络流最小成本 (Min Cost Max Flow) - 配送成本优化
// ============================================================================

// MinCostMaxFlowGraph 最小成本最大流图
type MinCostMaxFlowGraph struct {
	capacity map[string]map[string]int64
	cost     map[string]map[string]int64
	flow     map[string]map[string]int64
	nodes    map[string]bool
	mu       sync.RWMutex
}

// NewMinCostMaxFlowGraph 创建最小成本最大流图
func NewMinCostMaxFlowGraph() *MinCostMaxFlowGraph {
	return &MinCostMaxFlowGraph{
		capacity: make(map[string]map[string]int64),
		cost:     make(map[string]map[string]int64),
		flow:     make(map[string]map[string]int64),
		nodes:    make(map[string]bool),
	}
}

// AddEdge 添加边（容量和成本）
func (g *MinCostMaxFlowGraph) AddEdge(from, to string, cap, c int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.capacity[from] == nil {
		g.capacity[from] = make(map[string]int64)
		g.cost[from] = make(map[string]int64)
		g.flow[from] = make(map[string]int64)
	}
	if g.capacity[to] == nil {
		g.capacity[to] = make(map[string]int64)
		g.cost[to] = make(map[string]int64)
		g.flow[to] = make(map[string]int64)
	}

	g.capacity[from][to] += cap
	g.cost[from][to] = c
	g.cost[to][from] = -c
	g.nodes[from] = true
	g.nodes[to] = true
}

// MinCostMaxFlow 最小成本最大流（SPFA算法）
// 应用: 多仓库配送成本最小化
func (g *MinCostMaxFlowGraph) MinCostMaxFlow(source, sink string, maxFlow int64) (int64, int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	totalFlow := int64(0)
	totalCost := int64(0)

	for totalFlow < maxFlow {
		// SPFA找最短路径
		dist, parent := g.spfa(source)

		if dist[sink] == math.MaxInt64 {
			break
		}

		// 找路径上的最小容量
		pathFlow := maxFlow - totalFlow
		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			residual := g.capacity[u][v] - g.flow[u][v]
			if residual < pathFlow {
				pathFlow = residual
			}
		}

		// 更新流量和成本
		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			g.flow[u][v] += pathFlow
			g.flow[v][u] -= pathFlow
			totalCost += pathFlow * g.cost[u][v]
		}

		totalFlow += pathFlow
	}

	return totalFlow, totalCost
}

// spfa 最短路径快速算法
func (g *MinCostMaxFlowGraph) spfa(source string) (map[string]int64, map[string]string) {
	dist := make(map[string]int64)
	parent := make(map[string]string)
	inQueue := make(map[string]bool)

	for node := range g.nodes {
		dist[node] = math.MaxInt64
	}
	dist[source] = 0

	queue := []string{source}
	inQueue[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		inQueue[u] = false

		for v := range g.capacity[u] {
			if g.capacity[u][v] > g.flow[u][v] {
				newDist := dist[u] + g.cost[u][v]
				if newDist < dist[v] {
					dist[v] = newDist
					parent[v] = u
					if !inQueue[v] {
						queue = append(queue, v)
						inQueue[v] = true
					}
				}
			}
		}
	}

	return dist, parent
}
