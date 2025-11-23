package algorithm

import (
	"sync"
)

// ============================================================================
// 5. 匹配算法 - 用户与商品匹配
// ============================================================================

// HungarianAlgorithm 匈牙利算法（最优分配）
type HungarianAlgorithm struct {
	cost    [][]int64
	n       int
	match   []int
	visited []bool
	mu      sync.RWMutex
}

// NewHungarianAlgorithm 创建匈牙利算法求解器
func NewHungarianAlgorithm(costMatrix [][]int64) *HungarianAlgorithm {
	n := len(costMatrix)
	return &HungarianAlgorithm{
		cost:    costMatrix,
		n:       n,
		match:   make([]int, n),
		visited: make([]bool, n),
	}
}

// Solve 求解最优分配
// 应用: 用户与推荐商品的最优匹配
func (ha *HungarianAlgorithm) Solve() []int {
	ha.mu.Lock()
	defer ha.mu.Unlock()

	for i := 0; i < ha.n; i++ {
		ha.match[i] = -1
	}

	for i := 0; i < ha.n; i++ {
		for j := 0; j < ha.n; j++ {
			ha.visited[j] = false
		}
		ha.dfsHungarian(i)
	}

	return ha.match
}

// dfsHungarian 深度优先搜索
func (ha *HungarianAlgorithm) dfsHungarian(u int) bool {
	for v := 0; v < ha.n; v++ {
		if !ha.visited[v] && ha.cost[u][v] > 0 {
			ha.visited[v] = true

			if ha.match[v] == -1 || ha.dfsHungarian(ha.match[v]) {
				ha.match[v] = u
				return true
			}
		}
	}

	return false
}
