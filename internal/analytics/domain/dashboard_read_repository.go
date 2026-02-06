// 生成摘要：定义仪表板读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// DashboardReadRepository 定义仪表板读模型的高性能访问接口。
type DashboardReadRepository interface {
	// Save 保存或更新读模型。
	Save(ctx context.Context, dashboard *Dashboard) error
	// GetByID 根据仪表板ID获取读模型。
	GetByID(ctx context.Context, id uint64) (*Dashboard, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, id uint64) error
}
