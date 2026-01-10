// Package mysql 提供了商品基础设施的持久化逻辑与数据分析辅助工具。
package mysql

import (
	"log/slog"
	"time"

	algorithm "github.com/wyfcoding/pkg/algorithm/graph"
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
	maxID := 0
	for _, c := range categories {
		if c.ID > maxID {
			maxID = c.ID
		}
	}

	// 构建邻接表 (adj[parent] = [children])
	adj := make([][]int, maxID+1)
	root := 0 // 假设 0 为根节点，或者找到实际根节点
	hasRoot := false

	// 为了处理非连续ID，我们直接使用ID作为索引 (可能会有空洞，但LCA算法能处理)
	for _, c := range categories {
		if c.ParentID > 0 { // 假设 0 是顶层根或无效父ID
			if c.ParentID <= maxID {
				adj[c.ParentID] = append(adj[c.ParentID], c.ID)
			}
		} else {
			if !hasRoot {
				root = c.ID
				hasRoot = true
			}
			// 如果有多个根，这里简化处理只取第一个遇到的
		}
	}

	// 如果没有显式根节点 (ParentID=0)，默认 0 为根
	if !hasRoot && maxID > 0 {
		root = 0 // 虚拟根
		// 将所有顶级节点挂在 0 下?
		// 业务逻辑通常保证有一个根。这里简化。
	}

	// 构造 LCA 查询结构
	a.lca = algorithm.NewTreeLCA(root, adj)

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
