// 生成摘要：新增风控读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/risk/domain"
)

// RiskSecurityProjectionService 负责将风控事件投影到读模型。
type RiskSecurityProjectionService struct {
	repo               domain.RiskRepository
	analysisReadRepo   domain.RiskAnalysisReadRepository
	blacklistReadRepo  domain.BlacklistReadRepository
	behaviorReadRepo   domain.UserBehaviorReadRepository
	deviceReadRepo     domain.DeviceFingerprintReadRepository
	analysisSearchRepo domain.RiskAnalysisSearchRepository
	logger             *slog.Logger
}

// NewRiskSecurityProjectionService 创建投影服务。
func NewRiskSecurityProjectionService(
	repo domain.RiskRepository,
	analysisReadRepo domain.RiskAnalysisReadRepository,
	blacklistReadRepo domain.BlacklistReadRepository,
	behaviorReadRepo domain.UserBehaviorReadRepository,
	deviceReadRepo domain.DeviceFingerprintReadRepository,
	analysisSearchRepo domain.RiskAnalysisSearchRepository,
	logger *slog.Logger,
) *RiskSecurityProjectionService {
	return &RiskSecurityProjectionService{
		repo:               repo,
		analysisReadRepo:   analysisReadRepo,
		blacklistReadRepo:  blacklistReadRepo,
		behaviorReadRepo:   behaviorReadRepo,
		deviceReadRepo:     deviceReadRepo,
		analysisSearchRepo: analysisSearchRepo,
		logger:             logger,
	}
}

func (s *RiskSecurityProjectionService) OnRiskAnalysisCreated(ctx context.Context, event *domain.RiskAnalysisCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshRiskAnalysis(ctx, event.ResultID, event.UserID)
}

func (s *RiskSecurityProjectionService) OnBlacklistAdded(ctx context.Context, event *domain.BlacklistAddedEvent) error {
	if event == nil || s.blacklistReadRepo == nil {
		return nil
	}
	entry, err := s.repo.GetBlacklistByID(ctx, event.BlacklistID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load blacklist for projection", "blacklist_id", event.BlacklistID, "error", err)
		return err
	}
	if entry == nil {
		return nil
	}
	if err := s.blacklistReadRepo.Save(ctx, entry); err != nil {
		s.logger.ErrorContext(ctx, "failed to save blacklist cache", "blacklist_id", entry.ID, "error", err)
		return err
	}
	return nil
}

func (s *RiskSecurityProjectionService) OnBlacklistRemoved(ctx context.Context, event *domain.BlacklistRemovedEvent) error {
	if event == nil || s.blacklistReadRepo == nil {
		return nil
	}
	if event.Type != "" && event.Value != "" {
		_ = s.blacklistReadRepo.DeleteByTypeValue(ctx, event.Type, event.Value)
	}
	if event.BlacklistID > 0 {
		_ = s.blacklistReadRepo.DeleteByID(ctx, event.BlacklistID)
	}
	return nil
}

func (s *RiskSecurityProjectionService) OnUserBehaviorUpdated(ctx context.Context, event *domain.UserBehaviorUpdatedEvent) error {
	if event == nil || s.behaviorReadRepo == nil {
		return nil
	}
	behavior, err := s.repo.GetUserBehavior(ctx, event.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load user behavior for projection", "user_id", event.UserID, "error", err)
		return err
	}
	if behavior == nil {
		_ = s.behaviorReadRepo.DeleteByUserID(ctx, event.UserID)
		return nil
	}
	if err := s.behaviorReadRepo.Save(ctx, behavior); err != nil {
		s.logger.ErrorContext(ctx, "failed to save user behavior cache", "user_id", event.UserID, "error", err)
		return err
	}
	return nil
}

func (s *RiskSecurityProjectionService) OnDeviceFingerprintSaved(ctx context.Context, event *domain.DeviceFingerprintSavedEvent) error {
	if event == nil || s.deviceReadRepo == nil {
		return nil
	}
	fp, err := s.repo.GetDeviceFingerprint(ctx, event.DeviceID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load device fingerprint for projection", "device_id", event.DeviceID, "error", err)
		return err
	}
	if fp == nil {
		_ = s.deviceReadRepo.DeleteByDeviceID(ctx, event.DeviceID)
		return nil
	}
	if err := s.deviceReadRepo.Save(ctx, fp); err != nil {
		s.logger.ErrorContext(ctx, "failed to save device fingerprint cache", "device_id", event.DeviceID, "error", err)
		return err
	}
	return nil
}

func (s *RiskSecurityProjectionService) refreshRiskAnalysis(ctx context.Context, resultID uint64, userID uint64) error {
	result, err := s.repo.GetAnalysisResult(ctx, resultID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load risk analysis for projection", "result_id", resultID, "error", err)
		return err
	}
	if result == nil {
		if s.analysisReadRepo != nil && userID > 0 {
			_ = s.analysisReadRepo.DeleteLatestByUser(ctx, userID)
		}
		if s.analysisSearchRepo != nil && resultID > 0 {
			_ = s.analysisSearchRepo.Delete(ctx, resultID)
		}
		return nil
	}
	if s.analysisReadRepo != nil {
		if err := s.analysisReadRepo.SaveLatest(ctx, result.UserID, result); err != nil {
			s.logger.ErrorContext(ctx, "failed to save risk analysis cache", "result_id", resultID, "error", err)
			return err
		}
	}
	if s.analysisSearchRepo != nil {
		if err := s.analysisSearchRepo.Index(ctx, result); err != nil {
			s.logger.ErrorContext(ctx, "failed to index risk analysis", "result_id", resultID, "error", err)
			return err
		}
	}
	return nil
}
