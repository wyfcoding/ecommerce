// Package mysql 提供了商品基础设施的持久化逻辑与数据分析辅助工具。
package mysql

import (
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/algorithm"
)

// CategoryInfo 封装了用于拓扑分析的类目精简层级信息。
type CategoryInfo struct {
	ID       int // 类目唯一 ID
	ParentID int // 父类目 ID
}

// HierarchyAnalyzer 提供了针对商品类目树的深度分析能力，支持最近公共祖先 (LCA) 查询及层级距离计算。
// 核心价值：辅助实现关联推荐、相似品类识别等业务。
type HierarchyAnalyzer struct {
	lca *algorithm.TreeLCA // 基于离线倍增算法的 LCA 执行器
}

// Build 构造类目层级索引。
// 流程：转换节点格式 -> 识别树规模 -> 预处理倍增表。
func (a *HierarchyAnalyzer) Build(categories []CategoryInfo) {
	start := time.Now()
	nodes := make([]algorithm.LCANode, len(categories))
	maxID := 0
	for i, c := range categories {
		nodes[i] = algorithm.LCANode{ID: c.ID, ParentID: c.ParentID}
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	// 执行预处理，构建 O(N log N) 的查询结构
	a.lca = algorithm.NewTreeLCA(maxID+1, nodes)

	slog.Info("category hierarchy analyzer index built successfully",
		"nodes_count", len(categories),
		"max_depth", maxID,
		"duration", time.Since(start),
	)
}

// FindCommonParent 查找两个类目在树结构上的最近公共祖先 ID。
func (a *HierarchyAnalyzer) FindCommonParent(id1, id2 int) int {
	if a.lca == nil {
		return -1
	}
	return a.lca.GetLCA(id1, id2)
}

// GetPathDistance 计算两个类目节点之间的路径步数距离。
func (a *HierarchyAnalyzer) GetPathDistance(id1, id2 int) int {
	if a.lca == nil {
		return -1
	}
	return a.lca.GetDistance(id1, id2)
}
