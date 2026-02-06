// 生成摘要：定义审计报告读模型仓储接口（Redis）。
package domain

import "context"

// AuditReportReadRepository 定义审计报告读模型的高性能访问接口。
type AuditReportReadRepository interface {
	Save(ctx context.Context, report *AuditReport) error
	GetByID(ctx context.Context, id uint64) (*AuditReport, error)
	Delete(ctx context.Context, id uint64) error
}
