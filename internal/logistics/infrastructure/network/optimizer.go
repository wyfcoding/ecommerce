package network

import (
	"github.com/wyfcoding/pkg/algorithm"
)

// TransportLink 代表运输路径
type TransportLink struct {
	FromID   int
	ToID     int
	Capacity int
	UnitCost int
}

// NetworkOptimizer 全链路物流成本优化器
type NetworkOptimizer struct{}

// OptimizeFlow 寻找全链路最低成本的发货方案
func (o *NetworkOptimizer) OptimizeFlow(nodes int, links []TransportLink, source, sink int) (int, int) {
	mcmf := algorithm.NewMCMF(nodes)
	for _, link := range links {
		mcmf.AddEdge(link.FromID, link.ToID, link.Capacity, link.UnitCost)
	}

	// 返回 (最大发货量, 最小总成本)
	return mcmf.Solve(source, sink)
}
