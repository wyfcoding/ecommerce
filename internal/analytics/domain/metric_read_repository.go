// 生成摘要：定义指标读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// MetricReadRepository 定义指标读模型的高性能访问接口。
type MetricReadRepository interface {
	// Save 保存或更新读模型。
	Save(ctx context.Context, metric *Metric) error
	// GetByID 根据指标ID获取读模型。
	GetByID(ctx context.Context, id uint64) (*Metric, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, id uint64) error
}
