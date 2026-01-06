// Package scheduler 提供了基于高性能算法的超时调度器实现。
package scheduler

import (
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// WheelScheduler 实现了 domain.TimeoutScheduler 接口，利用分层时间轮（Timing Wheel）算法高效处理海量延时任务。
// 相对于标准库的 time.After，时间轮在任务规模极大时具有显著的内存和 CPU 优势（O(1) 插入与删除）。
type WheelScheduler struct {
	wheel *algorithm.TimingWheel // 底层核心时间轮算法实例
}

// NewWheelScheduler 初始化并返回一个新的时间轮调度器。
// 参数说明：
//   - tick: 指针移动的时间间隔（解析精度，如 1s）
//   - buckets: 每一层时间轮的槽位数量（如 3600 代表一小时一圈）
func NewWheelScheduler(tick time.Duration, buckets int) (*WheelScheduler, error) {
	tw, err := algorithm.NewTimingWheel(tick, buckets)
	if err != nil {
		slog.Error("failed to initialize timing_wheel", "error", err)
		return nil, err
	}
	slog.Info("wheel_scheduler initialized", "tick", tick, "buckets", buckets)
	return &WheelScheduler{
		wheel: tw,
	}, nil
}

// ScheduleTimeout 注册一个新的超时任务。
// 当达到指定的 timeout 时间后，将自动触发 callback。
func (s *WheelScheduler) ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error {
	// 封装业务回调，注入安全 Panic 恢复机制与追踪日志
	task := func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in timeout task", "order_id", orderID, "recover", r)
			}
		}()
		slog.Debug("timeout task triggered", "order_id", orderID)
		callback(orderID)
	}

	slog.Debug("scheduling timeout task", "order_id", orderID, "timeout", timeout)
	return s.wheel.AddTask(timeout, task)
}

// Start 正式开启时间轮的指针推进轮询。
func (s *WheelScheduler) Start() {
	s.wheel.Start()
}

// Stop 安全停止时间轮运行并释放相关资源。
func (s *WheelScheduler) Stop() {
	s.wheel.Stop()
}

var _ domain.TimeoutScheduler = (*WheelScheduler)(nil)
