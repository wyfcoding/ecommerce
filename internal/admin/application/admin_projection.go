// 生成摘要：新增管理后台读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AdminProjectionService 负责将管理后台事件投影到读模型。
type AdminProjectionService struct {
	userRepo        domain.AdminRepository
	userReadRepo    domain.AdminUserReadRepository
	settingRepo     domain.SettingRepository
	settingReadRepo domain.SettingReadRepository
	auditRepo       domain.AuditRepository
	auditSearchRepo domain.AuditLogSearchRepository
	logger          *slog.Logger
}

// NewAdminProjectionService 创建投影服务。
func NewAdminProjectionService(
	userRepo domain.AdminRepository,
	userReadRepo domain.AdminUserReadRepository,
	settingRepo domain.SettingRepository,
	settingReadRepo domain.SettingReadRepository,
	auditRepo domain.AuditRepository,
	auditSearchRepo domain.AuditLogSearchRepository,
	logger *slog.Logger,
) *AdminProjectionService {
	return &AdminProjectionService{
		userRepo:        userRepo,
		userReadRepo:    userReadRepo,
		settingRepo:     settingRepo,
		settingReadRepo: settingReadRepo,
		auditRepo:       auditRepo,
		auditSearchRepo: auditSearchRepo,
		logger:          logger,
	}
}

func (s *AdminProjectionService) OnAdminUserChanged(ctx context.Context, userID uint) error {
	if userID == 0 || s.userReadRepo == nil {
		return nil
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load admin user for projection", "user_id", userID, "error", err)
		return err
	}
	if user == nil {
		_ = s.userReadRepo.Delete(ctx, userID)
		return nil
	}
	if err := s.userReadRepo.Save(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to save admin user read model", "user_id", userID, "error", err)
		return err
	}
	return nil
}

func (s *AdminProjectionService) OnSystemSettingUpdated(ctx context.Context, key string) error {
	if key == "" || s.settingReadRepo == nil {
		return nil
	}
	setting, err := s.settingRepo.GetByKey(ctx, key)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load setting for projection", "key", key, "error", err)
		return err
	}
	if setting == nil {
		_ = s.settingReadRepo.Delete(ctx, key)
		return nil
	}
	if err := s.settingReadRepo.Save(ctx, setting); err != nil {
		s.logger.ErrorContext(ctx, "failed to save setting read model", "key", key, "error", err)
		return err
	}
	return nil
}

func (s *AdminProjectionService) OnAuditLogCreated(ctx context.Context, logID uint) error {
	if logID == 0 || s.auditSearchRepo == nil {
		return nil
	}
	logEntry, err := s.auditRepo.GetByID(ctx, logID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load audit log for projection", "log_id", logID, "error", err)
		return err
	}
	if logEntry == nil {
		_ = s.auditSearchRepo.Delete(ctx, logID)
		return nil
	}
	if err := s.auditSearchRepo.Index(ctx, logEntry); err != nil {
		s.logger.ErrorContext(ctx, "failed to index audit log", "log_id", logID, "error", err)
		return err
	}
	return nil
}
