package domain

import "time"

// TimeoutScheduler 定义超时调度接口
type TimeoutScheduler interface {
	// ScheduleTimeout 调度一个超时任务
	ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error
	// Start 启动调度器
	Start()
	// Stop 停止调度器
	Stop()
}
