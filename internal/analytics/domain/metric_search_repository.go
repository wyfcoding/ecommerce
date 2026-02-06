// 生成摘要：定义指标搜索仓储接口（Elasticsearch），用于分页与过滤查询。
package domain

import "context"

// MetricSearchRepository 定义指标搜索仓储接口。
type MetricSearchRepository interface {
	// Index 将指标写入搜索索引。
	Index(ctx context.Context, metric *Metric) error
	// Delete 从索引中删除指标。
	Delete(ctx context.Context, metricID uint64) error
	// Search 按条件检索指标（支持类型/粒度/维度/时间范围、分页）。
	Search(ctx context.Context, query *MetricQuery, offset, limit int) ([]*Metric, int64, error)
}
