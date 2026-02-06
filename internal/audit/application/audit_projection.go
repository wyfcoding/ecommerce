// 生成摘要：新增审计读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// AuditProjectionService 负责将审计事件投影到读模型。
type AuditProjectionService struct {
	repo           domain.AuditRepository
	logReadRepo    domain.AuditLogReadRepository
	policyReadRepo domain.AuditPolicyReadRepository
	reportReadRepo domain.AuditReportReadRepository
	logSearchRepo  domain.AuditLogSearchRepository
	logger         *slog.Logger
}

// NewAuditProjectionService 创建投影服务。
func NewAuditProjectionService(
	repo domain.AuditRepository,
	logReadRepo domain.AuditLogReadRepository,
	policyReadRepo domain.AuditPolicyReadRepository,
	reportReadRepo domain.AuditReportReadRepository,
	logSearchRepo domain.AuditLogSearchRepository,
	logger *slog.Logger,
) *AuditProjectionService {
	return &AuditProjectionService{
		repo:           repo,
		logReadRepo:    logReadRepo,
		policyReadRepo: policyReadRepo,
		reportReadRepo: reportReadRepo,
		logSearchRepo:  logSearchRepo,
		logger:         logger,
	}
}

func (s *AuditProjectionService) OnLogCreated(ctx context.Context, event *domain.AuditLogCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshLog(ctx, event.AuditID)
}

func (s *AuditProjectionService) OnLogDeleted(ctx context.Context, event *domain.AuditLogDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.logReadRepo != nil {
		_ = s.logReadRepo.Delete(ctx, event.AuditID)
	}
	if s.logSearchRepo != nil {
		_ = s.logSearchRepo.Delete(ctx, event.AuditID)
	}
	return nil
}

func (s *AuditProjectionService) OnPolicyCreated(ctx context.Context, event *domain.AuditPolicyCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshPolicy(ctx, event.PolicyID)
}

func (s *AuditProjectionService) OnPolicyUpdated(ctx context.Context, event *domain.AuditPolicyUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshPolicy(ctx, event.PolicyID)
}

func (s *AuditProjectionService) OnPolicyDeleted(ctx context.Context, event *domain.AuditPolicyDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.policyReadRepo != nil {
		_ = s.policyReadRepo.Delete(ctx, event.PolicyID)
	}
	return nil
}

func (s *AuditProjectionService) OnReportCreated(ctx context.Context, event *domain.AuditReportCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AuditProjectionService) OnReportUpdated(ctx context.Context, event *domain.AuditReportUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AuditProjectionService) OnReportGenerated(ctx context.Context, event *domain.AuditReportGeneratedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AuditProjectionService) OnReportPublished(ctx context.Context, event *domain.AuditReportPublishedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshReport(ctx, event.ReportID)
}

func (s *AuditProjectionService) OnReportDeleted(ctx context.Context, event *domain.AuditReportDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.reportReadRepo != nil {
		_ = s.reportReadRepo.Delete(ctx, event.ReportID)
	}
	return nil
}

func (s *AuditProjectionService) refreshLog(ctx context.Context, logID uint64) error {
	if s.logReadRepo == nil && s.logSearchRepo == nil {
		return nil
	}
	logItem, err := s.repo.GetLog(ctx, logID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load audit log for projection", "log_id", logID, "error", err)
		return err
	}
	if logItem == nil {
		if s.logReadRepo != nil {
			_ = s.logReadRepo.Delete(ctx, logID)
		}
		if s.logSearchRepo != nil {
			_ = s.logSearchRepo.Delete(ctx, logID)
		}
		return nil
	}
	if s.logReadRepo != nil {
		if err := s.logReadRepo.Save(ctx, logItem); err != nil {
			s.logger.ErrorContext(ctx, "failed to save audit log cache", "log_id", logID, "error", err)
			return err
		}
	}
	if s.logSearchRepo != nil {
		if err := s.logSearchRepo.Index(ctx, logItem); err != nil {
			s.logger.ErrorContext(ctx, "failed to index audit log", "log_id", logID, "error", err)
			return err
		}
	}
	return nil
}

func (s *AuditProjectionService) refreshPolicy(ctx context.Context, policyID uint64) error {
	if s.policyReadRepo == nil {
		return nil
	}
	policy, err := s.repo.GetPolicy(ctx, policyID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load audit policy for projection", "policy_id", policyID, "error", err)
		return err
	}
	if policy == nil {
		_ = s.policyReadRepo.Delete(ctx, policyID)
		return nil
	}
	if err := s.policyReadRepo.Save(ctx, policy); err != nil {
		s.logger.ErrorContext(ctx, "failed to save audit policy cache", "policy_id", policyID, "error", err)
		return err
	}
	return nil
}

func (s *AuditProjectionService) refreshReport(ctx context.Context, reportID uint64) error {
	if s.reportReadRepo == nil {
		return nil
	}
	report, err := s.repo.GetReport(ctx, reportID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load audit report for projection", "report_id", reportID, "error", err)
		return err
	}
	if report == nil {
		_ = s.reportReadRepo.Delete(ctx, reportID)
		return nil
	}
	if err := s.reportReadRepo.Save(ctx, report); err != nil {
		s.logger.ErrorContext(ctx, "failed to save audit report cache", "report_id", reportID, "error", err)
		return err
	}
	return nil
}
