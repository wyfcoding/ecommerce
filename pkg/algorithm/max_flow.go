package algorithm

import (
	"math"
	"sync"
)

// ============================================================================
// 1. 最大流算法 (Max Flow) - 库存分配优化
// ============================================================================

// MaxFlowGraph 最大流图
type MaxFlowGraph struct {
	capacity map[string]map[string]int64 // 容量矩阵
	flow     map[string]map[string]int64 // 流量矩阵
	nodes    map[string]bool
	mu       sync.RWMutex
}

// NewMaxFlowGraph 创建最大流图
func NewMaxFlowGraph() *MaxFlowGraph {
	return &MaxFlowGraph{
		capacity: make(map[string]map[string]int64),
		flow:     make(map[string]map[string]int64),
		nodes:    make(map[string]bool),
	}
}

// AddEdge 添加边（容量）
func (g *MaxFlowGraph) AddEdge(from, to string, cap int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.capacity[from] == nil {
		g.capacity[from] = make(map[string]int64)
		g.flow[from] = make(map[string]int64)
	}
	if g.capacity[to] == nil {
		g.capacity[to] = make(map[string]int64)
		g.flow[to] = make(map[string]int64)
	}

	g.capacity[from][to] += cap
	g.nodes[from] = true
	g.nodes[to] = true
}

// FordFulkerson 福特-富尔克森算法（最大流）
// 应用: 多仓库到多订单的最优库存分配
func (g *MaxFlowGraph) FordFulkerson(source, sink string) int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	maxFlow := int64(0)

	for {
		// BFS找增广路径
		parent := g.bfs(source, sink)
		if parent == nil {
			break
		}

		// 找路径上的最小容量
		pathFlow := int64(math.MaxInt64)
		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			residual := g.capacity[u][v] - g.flow[u][v]
			if residual < pathFlow {
				pathFlow = residual
			}
		}

		// 更新流量
		for v := sink; v != source; v = parent[v] {
			u := parent[v]
			g.flow[u][v] += pathFlow
			g.flow[v][u] -= pathFlow
		}

		maxFlow += pathFlow
	}

	return maxFlow
}

// bfs 广度优先搜索找增广路径
func (g *MaxFlowGraph) bfs(source, sink string) map[string]string {
	visited := make(map[string]bool)
	parent := make(map[string]string)
	queue := []string{source}
	visited[source] = true

	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]

		if u == sink {
			return parent
		}

		for v := range g.capacity[u] {
			if !visited[v] && g.capacity[u][v] > g.flow[u][v] {
				visited[v] = true
				parent[v] = u
				queue = append(queue, v)
			}
		}
	}

	return nil
}
