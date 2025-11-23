package algorithm

import (
	"sync"
)

// ============================================================================
// 9. 线段树 (Segment Tree) - 区间操作优化
// ============================================================================

// SegmentTree 线段树
type SegmentTree struct {
	tree []int64
	n    int
	mu   sync.RWMutex
}

// NewSegmentTree 创建线段树
func NewSegmentTree(n int) *SegmentTree {
	return &SegmentTree{
		tree: make([]int64, 4*n),
		n:    n,
	}
}

// Update 单点更新
func (st *SegmentTree) Update(idx int, val int64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.update(1, 0, st.n-1, idx, val)
}

// update 递归更新
func (st *SegmentTree) update(node, start, end, idx int, val int64) {
	if start == end {
		st.tree[node] = val
		return
	}

	mid := (start + end) / 2
	if idx <= mid {
		st.update(2*node, start, mid, idx, val)
	} else {
		st.update(2*node+1, mid+1, end, idx, val)
	}

	st.tree[node] = st.tree[2*node] + st.tree[2*node+1]
}

// Query 区间查询
// 应用: 库存区间查询、销量统计
func (st *SegmentTree) Query(left, right int) int64 {
	st.mu.RLock()
	defer st.mu.RUnlock()

	return st.query(1, 0, st.n-1, left, right)
}

// query 递归查询
func (st *SegmentTree) query(node, start, end, left, right int) int64 {
	if right < start || end < left {
		return 0
	}

	if left <= start && end <= right {
		return st.tree[node]
	}

	mid := (start + end) / 2
	leftSum := st.query(2*node, start, mid, left, right)
	rightSum := st.query(2*node+1, mid+1, end, left, right)

	return leftSum + rightSum
}
