// Package counter 提供了基于概率数据结构的频率统计实现。
package counter

import (
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/algorithm"
)

// TrafficMonitor 流量监控器
type TrafficMonitor struct {
	cms    *algorithm.CountMinSketch
	logger *slog.Logger
	limit  uint64 // 判定为攻击的频率阈值
}

// NewTrafficMonitor 创建流量监控器
func NewTrafficMonitor(threshold uint64, logger *slog.Logger) *TrafficMonitor {
	// epsilon=0.001, delta=0.01 (高精度，低内存占用)
	cms, err := algorithm.NewCountMinSketch(0.001, 0.01)
	if err != nil {
		logger.Error("failed to initialize CountMinSketch", "error", err)
		// 即使失败也返回一个空的结构或抛出异常，这里由于是初始化函数且返回类型不含 error，
		// 生产环境下建议修改函数签名，此处暂且记录日志。
	}

	m := &TrafficMonitor{
		cms:    cms,
		logger: logger,
		limit:  threshold,
	}

	// 启动后台衰减协程：实现类似滑动窗口的效果
	go m.startDecayLoop()

	return m
}

func (m *TrafficMonitor) startDecayLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.logger.Debug("performing CMS decay to slide frequency window")
		m.cms.Decay()
	}
}

// RecordAndCheck 记录访问并检查是否触发风险
func (m *TrafficMonitor) RecordAndCheck(key string) bool {
	m.cms.Add([]byte(key), 1)
	count := m.cms.Estimate([]byte(key))

	if count > m.limit {
		m.logger.Warn("risk detected: frequency limit exceeded",
			"key", key,
			"estimated_count", count,
			"threshold", m.limit)
		return true // 触发风险
	}
	return false // 安全
}

// GetEstimation 获取当前的频率估算值
func (m *TrafficMonitor) GetEstimation(key string) uint64 {
	return m.cms.Estimate([]byte(key))
}
