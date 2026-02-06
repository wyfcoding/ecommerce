// 生成摘要：定义报告读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// ReportReadRepository 定义报告读模型的高性能访问接口。
type ReportReadRepository interface {
	// Save 保存或更新读模型。
	Save(ctx context.Context, report *Report) error
	// GetByID 根据报告ID获取读模型。
	GetByID(ctx context.Context, id uint64) (*Report, error)
	// Delete 删除读模型数据。
	Delete(ctx context.Context, id uint64) error
}
