package application

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/audit/domain/entity"
	"ecommerce/internal/audit/domain/repository"
	"ecommerce/pkg/idgen"

	"log/slog"
)

type AuditService struct {
	repo        repository.AuditRepository
	idGenerator idgen.Generator
	logger      *slog.Logger
}

func NewAuditService(repo repository.AuditRepository, idGenerator idgen.Generator, logger *slog.Logger) *AuditService {
	return &AuditService{
		repo:        repo,
		idGenerator: idGenerator,
		logger:      logger,
	}
}

// LogEvent 记录审计事件
func (s *AuditService) LogEvent(ctx context.Context, userID uint64, username string, eventType entity.AuditEventType, module, action string, opts ...LogOption) error {
	auditNo := fmt.Sprintf("AUD%d", s.idGenerator.Generate())
	log := entity.NewAuditLog(auditNo, userID, username, eventType, module, action)

	for _, opt := range opts {
		opt(log)
	}

	if err := s.repo.CreateLog(ctx, log); err != nil {
		s.logger.Error("failed to create audit log", "error", err)
		return err
	}
	return nil
}

type LogOption func(*entity.AuditLog)

func WithError(errMsg string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetError(errMsg)
	}
}

func WithResource(resourceType, resourceID string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetResource(resourceType, resourceID)
	}
}

func WithChange(oldValue, newValue string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetChange(oldValue, newValue)
	}
}

func WithClientInfo(ip, userAgent string) LogOption {
	return func(l *entity.AuditLog) {
		l.SetClientInfo(ip, userAgent)
	}
}

func WithDuration(duration int64) LogOption {
	return func(l *entity.AuditLog) {
		l.SetDuration(duration)
	}
}

// QueryLogs 查询审计日志
func (s *AuditService) QueryLogs(ctx context.Context, query *repository.AuditLogQuery) ([]*entity.AuditLog, int64, error) {
	return s.repo.ListLogs(ctx, query)
}

// CreatePolicy 创建审计策略
func (s *AuditService) CreatePolicy(ctx context.Context, name, description string) (*entity.AuditPolicy, error) {
	policy := entity.NewAuditPolicy(name, description)
	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		s.logger.Error("failed to create audit policy", "error", err)
		return nil, err
	}
	return policy, nil
}

// UpdatePolicy 更新审计策略
func (s *AuditService) UpdatePolicy(ctx context.Context, id uint64, eventTypes, modules []string, enabled bool) error {
	policy, err := s.repo.GetPolicy(ctx, id)
	if err != nil {
		return err
	}

	policy.EventTypes = eventTypes
	policy.Modules = modules
	policy.Enabled = enabled
	policy.UpdatedAt = time.Now()

	return s.repo.UpdatePolicy(ctx, policy)
}

// ListPolicies 获取审计策略列表
func (s *AuditService) ListPolicies(ctx context.Context, page, pageSize int) ([]*entity.AuditPolicy, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPolicies(ctx, offset, pageSize)
}

// CreateReport 创建审计报告
func (s *AuditService) CreateReport(ctx context.Context, title, description string) (*entity.AuditReport, error) {
	reportNo := fmt.Sprintf("AUDRPT%d", s.idGenerator.Generate())
	report := entity.NewAuditReport(reportNo, title, description)

	if err := s.repo.CreateReport(ctx, report); err != nil {
		s.logger.Error("failed to create audit report", "error", err)
		return nil, err
	}
	return report, nil
}

// GenerateReport 生成报告内容
func (s *AuditService) GenerateReport(ctx context.Context, id uint64) error {
	report, err := s.repo.GetReport(ctx, id)
	if err != nil {
		return err
	}

	// Mock report generation
	content := fmt.Sprintf("Audit Report for %s generated at %s", report.Title, time.Now().Format(time.RFC3339))
	report.Generate(content)

	return s.repo.UpdateReport(ctx, report)
}

// ListReports 获取审计报告列表
func (s *AuditService) ListReports(ctx context.Context, page, pageSize int) ([]*entity.AuditReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReports(ctx, offset, pageSize)
}
