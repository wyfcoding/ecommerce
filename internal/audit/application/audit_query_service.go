package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// AuditQueryService 处理审计模块的查询操作。
type AuditQueryService struct {
	repo           domain.AuditRepository
	logReadRepo    domain.AuditLogReadRepository
	policyReadRepo domain.AuditPolicyReadRepository
	reportReadRepo domain.AuditReportReadRepository
	logSearchRepo  domain.AuditLogSearchRepository
	logger         *slog.Logger
}

// NewAuditQueryService 创建并返回一个新的 AuditQueryService 实例。
func NewAuditQueryService(
	repo domain.AuditRepository,
	logReadRepo domain.AuditLogReadRepository,
	policyReadRepo domain.AuditPolicyReadRepository,
	reportReadRepo domain.AuditReportReadRepository,
	logSearchRepo domain.AuditLogSearchRepository,
	logger *slog.Logger,
) *AuditQueryService {
	return &AuditQueryService{
		repo:           repo,
		logReadRepo:    logReadRepo,
		policyReadRepo: policyReadRepo,
		reportReadRepo: reportReadRepo,
		logSearchRepo:  logSearchRepo,
		logger:         logger,
	}
}

// GetLog 根据ID获取审计日志。
func (q *AuditQueryService) GetLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	if q.logReadRepo != nil {
		if cached, err := q.logReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	log, err := q.repo.GetLog(ctx, id)
	if err != nil {
		return nil, err
	}
	if log != nil && q.logReadRepo != nil {
		_ = q.logReadRepo.Save(ctx, log)
	}
	return log, nil
}

// ListLogs 获取审计日志列表。
func (q *AuditQueryService) ListLogs(ctx context.Context, query *domain.AuditLogQuery) ([]*domain.AuditLog, int64, error) {
	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.logSearchRepo != nil {
		list, total, err := q.logSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "audit log search fallback to mysql", "error", err)
	}
	return q.repo.ListLogs(ctx, query)
}

// GetPolicy 根据ID获取审计策略。
func (q *AuditQueryService) GetPolicy(ctx context.Context, id uint64) (*domain.AuditPolicy, error) {
	if q.policyReadRepo != nil {
		if cached, err := q.policyReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	policy, err := q.repo.GetPolicy(ctx, id)
	if err != nil {
		return nil, err
	}
	if policy != nil && q.policyReadRepo != nil {
		_ = q.policyReadRepo.Save(ctx, policy)
	}
	return policy, nil
}

// ListPolicies 获取审计策略列表。
func (q *AuditQueryService) ListPolicies(ctx context.Context, offset, limit int) ([]*domain.AuditPolicy, int64, error) {
	return q.repo.ListPolicies(ctx, offset, limit)
}

// GetReport 根据ID获取审计报告。
func (q *AuditQueryService) GetReport(ctx context.Context, id uint64) (*domain.AuditReport, error) {
	if q.reportReadRepo != nil {
		if cached, err := q.reportReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	report, err := q.repo.GetReport(ctx, id)
	if err != nil {
		return nil, err
	}
	if report != nil && q.reportReadRepo != nil {
		_ = q.reportReadRepo.Save(ctx, report)
	}
	return report, nil
}

// ListReports 获取审计报告列表。
func (q *AuditQueryService) ListReports(ctx context.Context, offset, limit int) ([]*domain.AuditReport, int64, error) {
	return q.repo.ListReports(ctx, offset, limit)
}
