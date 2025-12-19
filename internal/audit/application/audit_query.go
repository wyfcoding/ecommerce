package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// AuditQuery 处理审计模块的查询操作。
type AuditQuery struct {
	repo domain.AuditRepository
}

// NewAuditQuery 创建并返回一个新的 AuditQuery 实例。
func NewAuditQuery(repo domain.AuditRepository) *AuditQuery {
	return &AuditQuery{repo: repo}
}

// GetLog 根据ID获取审计日志。
func (q *AuditQuery) GetLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	return q.repo.GetLog(ctx, id)
}

// ListLogs 获取审计日志列表。
func (q *AuditQuery) ListLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	return q.repo.ListLogs(ctx, query)
}

// GetPolicy 根据ID获取审计策略。
func (q *AuditQuery) GetPolicy(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	return q.repo.GetPolicy(ctx, id)
}

// ListPolicies 获取审计策略列表。
func (q *AuditQuery) ListPolicies(ctx context.Context, offset, limit int) ([]*domain.AuditPolicy, int64, error) {
	return q.repo.ListPolicies(ctx, offset, limit)
}

// GetReport 根据ID获取审计报告。
func (q *AuditQuery) GetReport(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	return q.repo.GetReport(ctx, id)
}

// ListReports 获取审计报告列表。
func (q *AuditQuery) ListReports(ctx context.Context, offset, limit int) ([]*domain.AuditReport, int64, error) {
	return q.repo.ListReports(ctx, offset, limit)
}
