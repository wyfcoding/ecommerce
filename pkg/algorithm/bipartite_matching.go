package algorithm

import (
	"sync"
)

// ============================================================================
// 2. 二分图匹配算法 (Bipartite Matching) - 配送员与订单匹配
// ============================================================================

// BipartiteGraph 二分图
type BipartiteGraph struct {
	// 左侧节点（配送员）到右侧节点（订单）的边
	edges map[string][]string
	match map[string]string // 匹配结果
	mu    sync.RWMutex
}

// NewBipartiteGraph 创建二分图
func NewBipartiteGraph() *BipartiteGraph {
	return &BipartiteGraph{
		edges: make(map[string][]string),
		match: make(map[string]string),
	}
}

// AddEdge 添加边（配送员可以配送的订单）
func (bg *BipartiteGraph) AddEdge(deliveryMan, order string) {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	bg.edges[deliveryMan] = append(bg.edges[deliveryMan], order)
}

// MaxMatching 最大匹配（匈牙利算法）
// 应用: 配送员与订单的最优匹配
func (bg *BipartiteGraph) MaxMatching() map[string]string {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	result := make(map[string]string)
	visited := make(map[string]bool)

	for deliveryMan := range bg.edges {
		visited = make(map[string]bool) // 重置visited
		bg.dfs(deliveryMan, visited, result)
	}

	return result
}

// dfs 深度优先搜索找增广路径
func (bg *BipartiteGraph) dfs(u string, visited map[string]bool, match map[string]string) bool {
	for _, v := range bg.edges[u] {
		if visited[v] {
			continue
		}
		visited[v] = true

		// 如果v未匹配或能为v的匹配找到增广路径
		if match[v] == "" || bg.dfs(match[v], visited, match) {
			match[v] = u
			return true
		}
	}

	return false
}
