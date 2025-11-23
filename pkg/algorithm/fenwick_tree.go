package algorithm

import (
	"sync"
)

// ============================================================================
// 8. 树状数组 (Fenwick Tree) - 区间查询优化
// ============================================================================

// FenwickTree 树状数组
type FenwickTree struct {
	tree []int64
	n    int
	mu   sync.RWMutex
}

// NewFenwickTree 创建树状数组
func NewFenwickTree(n int) *FenwickTree {
	return &FenwickTree{
		tree: make([]int64, n+1),
		n:    n,
	}
}

// Update 更新元素
func (ft *FenwickTree) Update(idx int, delta int64) {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	idx++ // 1-indexed
	for idx <= ft.n {
		ft.tree[idx] += delta
		idx += idx & (-idx)
	}
}

// Query 查询前缀和
func (ft *FenwickTree) Query(idx int) int64 {
	ft.mu.RLock()
	defer ft.mu.RUnlock()

	idx++ // 1-indexed
	var sum int64 = 0
	for idx > 0 {
		sum += ft.tree[idx]
		idx -= idx & (-idx)
	}
	return sum
}

// RangeQuery 区间查询
// 应用: 库存区间查询、销量统计
func (ft *FenwickTree) RangeQuery(left, right int) int64 {
	if left == 0 {
		return ft.Query(right)
	}
	return ft.Query(right) - ft.Query(left-1)
}
