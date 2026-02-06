// 生成摘要：定义审计策略读模型仓储接口（Redis）。
package domain

import "context"

// AuditPolicyReadRepository 定义审计策略读模型的高性能访问接口。
type AuditPolicyReadRepository interface {
	Save(ctx context.Context, policy *AuditPolicy) error
	GetByID(ctx context.Context, id uint64) (*AuditPolicy, error)
	Delete(ctx context.Context, id uint64) error
}
