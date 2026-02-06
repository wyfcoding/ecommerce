// 生成摘要：定义审计日志搜索仓储接口（Elasticsearch）。
package domain

import "context"

// AuditLogSearchRepository 定义审计日志搜索仓储接口。
type AuditLogSearchRepository interface {
	Index(ctx context.Context, log *AuditLog) error
	Delete(ctx context.Context, logID uint64) error
	Search(ctx context.Context, query *AuditLogQuery, offset, limit int) ([]*AuditLog, int64, error)
}
