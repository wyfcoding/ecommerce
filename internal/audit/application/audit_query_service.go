package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// AuditQueryService 处理审计模块的查询操作。
type AuditQueryService struct {
	repo domain.AuditRepository
}

// NewAuditQueryService 创建并返回一个新的 AuditQueryService 实例。
func NewAuditQueryService(repo domain.AuditRepository) *AuditQueryService {
	return &AuditQueryService{repo: repo}
}

// GetLog 根据ID获取审计日志。
func (q *AuditQueryService) GetLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	return q.repo.GetLog(ctx, id)
}

// ListLogs 获取审计日志列表。
func (q *AuditQueryService) ListLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	return q.repo.ListLogs(ctx, query)
}

// GetPolicy 根据ID获取审计策略。
func (q *AuditQueryService) GetPolicy(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	return q.repo.GetPolicy(ctx, id)
}

// ListPolicies 获取审计策略列表。
func (q *AuditQueryService) ListPolicies(ctx context.Context, offset, limit int) ([]*domain.AuditPolicy, int64, error) {
	return q.repo.ListPolicies(ctx, offset, limit)
}

// GetReport 根据ID获取审计报告。
func (q *AuditQueryService) GetReport(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	return q.repo.GetReport(ctx, id)
}

// ListReports 获取审计报告列表。
func (q *AuditQueryService) ListReports(ctx context.Context, offset, limit int) ([]*domain.AuditReport, int64, error) {
	return q.repo.ListReports(ctx, offset, limit)
}
