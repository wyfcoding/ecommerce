package category

import (
	"github.com/wyfcoding/pkg/algorithm"
)

// CategoryInfo 类目精简信息
type CategoryInfo struct {
	ID       int
	ParentID int
}

// HierarchyAnalyzer 类目层级分析器
type HierarchyAnalyzer struct {
	lca *algorithm.TreeLCA
}

// Build 预处理类目树
func (a *HierarchyAnalyzer) Build(categories []CategoryInfo) {
	nodes := make([]algorithm.LCANode, len(categories))
	maxID := 0
	for i, c := range categories {
		nodes[i] = algorithm.LCANode{ID: c.ID, ParentID: c.ParentID}
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	// 简单起见，假设 ID 是连续的且从 0 开始。实际中可能需要映射 map。
	a.lca = algorithm.NewTreeLCA(maxID+1, nodes)
}

// FindCommonParent 找到两个类目的最近公共祖先
func (a *HierarchyAnalyzer) FindCommonParent(id1, id2 int) int {
	if a.lca == nil {
		return -1
	}
	return a.lca.GetLCA(id1, id2)
}

// GetPathDistance 获取类目间的距离
func (a *HierarchyAnalyzer) GetPathDistance(id1, id2 int) int {
	if a.lca == nil {
		return -1
	}
	return a.lca.GetDistance(id1, id2)
}
