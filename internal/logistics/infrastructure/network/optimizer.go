package network

import (
	algorithm "github.com/wyfcoding/pkg/algorithm/graph"
)

// TransportLink 代表运输路径，包含起点、终点、容量上限和单位运输成本。
type TransportLink struct {
	FromID   int
	ToID     int
	Capacity int64
	UnitCost int64
}

// Optimizer 全链路物流成本优化器。
type Optimizer struct{}

// NewOptimizer 构造函数。
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// OptimizeFlow 使用最小费用最大流算法寻找全链路最优发货方案。
// 返回: (最大流, 最小总成本)。
func (o *Optimizer) OptimizeFlow(nodes int, links []TransportLink, source, sink int) (int64, int64) {
	mcmf := algorithm.NewMinCostMaxFlowGraph(nodes)
	for _, link := range links {
		mcmf.AddEdgeByID(link.FromID, link.ToID, link.Capacity, link.UnitCost)
	}

	// 1<<62 代表无穷大流量上限，我们寻求的是在网络容量限制下的最大流。
	return mcmf.SolveByID(source, sink, 1<<62)
}
