package service

// Backend 节点信息
type Backend struct {
	ID   string
	Addr string
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// Pick 选择一个后端节点
	Pick() *Backend
	// ReportLatency 上报某个节点的响应延迟
	ReportLatency(id string, durationMs float64)
}
