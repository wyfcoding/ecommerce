// 生成摘要：定义审计日志读模型仓储接口（Redis），用于高频查询。
package domain

import "context"

// AuditLogReadRepository 定义审计日志读模型的高性能访问接口。
type AuditLogReadRepository interface {
	Save(ctx context.Context, log *AuditLog) error
	GetByID(ctx context.Context, id uint64) (*AuditLog, error)
	Delete(ctx context.Context, id uint64) error
}
