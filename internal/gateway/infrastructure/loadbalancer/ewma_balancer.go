package loadbalancer

import (
	"sync"

	"github.com/fynnwu/all/pkg/algorithm"
	"github.com/fynnwu/all/ecommerce/internal/gateway/domain/service"
)

// EWMABalancer 基于 EWMA 的响应耗时敏感负载均衡器
type EWMABalancer struct {
	nodes    []*nodeWrapper
	alpha    float64
	mu       sync.RWMutex
}

type nodeWrapper struct {
	backend *service.Backend
	ewma    *algorithm.EWMA
}

func NewEWMABalancer(alpha float64) *EWMABalancer {
	return &EWMABalancer{
		nodes: make([]*nodeWrapper, 0),
		alpha: alpha,
	}
}

// AddNode 添加节点
func (b *EWMABalancer) AddNode(backend *service.Backend) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nodes = append(b.nodes, &nodeWrapper{
		backend: backend,
		ewma:    algorithm.NewEWMA(b.alpha),
	})
}

// Pick 选择当前延迟最小的节点 (Latency-Aware)
func (b *EWMABalancer) Pick() *service.Backend {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.nodes) == 0 {
		return nil
	}

	var bestNode *nodeWrapper
	minLatency := -1.0

	for _, n := range b.nodes {
		latency := n.ewma.Value()
		if minLatency == -1.0 || latency < minLatency {
			minLatency = latency
			bestNode = n
		}
	}

	return bestNode.backend
}

// ReportLatency 上报延迟
func (b *EWMABalancer) ReportLatency(id string, durationMs float64) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, n := range b.nodes {
		if n.backend.ID == id {
			n.ewma.Update(durationMs)
			break
		}
	}
}

var _ service.LoadBalancer = (*EWMABalancer)(nil)
