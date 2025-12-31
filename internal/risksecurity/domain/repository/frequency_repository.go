package repository

import "context"

// FrequencyRepository 定义频率统计的接口
type FrequencyRepository interface {
	// Add 增加计数
	Add(ctx context.Context, key string, delta uint64) error
	// Estimate 获取估计的频率
	Estimate(ctx context.Context, key string) (uint64, error)
	// Reset 重置计数器
	Reset(ctx context.Context) error
}
