package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/frauddetection/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type FraudDetectionService struct {
	assessmentRepo   domain.RiskAssessmentRepository
	blacklistRepo    domain.BlacklistRepository
	ruleRepo         domain.RiskRuleRepository
	deviceRepo       domain.DeviceFingerprintRepository
	profileRepo      domain.UserRiskProfileRepository
	activityRepo     domain.SuspiciousActivityRepository
	ruleEngine       domain.RuleEngine
	idGenerator      idgen.Generator
	logger           *slog.Logger
}

func NewFraudDetectionService(
	assessmentRepo domain.RiskAssessmentRepository,
	blacklistRepo domain.BlacklistRepository,
	ruleRepo domain.RiskRuleRepository,
	deviceRepo domain.DeviceFingerprintRepository,
	profileRepo domain.UserRiskProfileRepository,
	activityRepo domain.SuspiciousActivityRepository,
	ruleEngine domain.RuleEngine,
	idGenerator idgen.Generator,
	logger *slog.Logger,
) *FraudDetectionService {
	return &FraudDetectionService{
		assessmentRepo:   assessmentRepo,
		blacklistRepo:    blacklistRepo,
		ruleRepo:         ruleRepo,
		deviceRepo:       deviceRepo,
		profileRepo:      profileRepo,
		activityRepo:     activityRepo,
		ruleEngine:       ruleEngine,
		idGenerator:      idGenerator,
		logger:           logger,
	}
}

type AnalyzeTransactionCommand struct {
	TransactionID     string
	UserID            uint64
	Amount            int64
	Currency          string
	PaymentMethod     string
	IPAddress         string
	DeviceFingerprint string
	UserAgent         string
	Country           string
	City              string
	Latitude          float64
	Longitude         float64
	MerchantID        uint64
	CardLastFour      string
	Email             string
	Phone             string
	Metadata          map[string]string
}

func (s *FraudDetectionService) AnalyzeTransaction(ctx context.Context, cmd *AnalyzeTransactionCommand) (*domain.RiskAssessment, error) {
	s.logger.InfoContext(ctx, "analyzing transaction for fraud", "transaction_id", cmd.TransactionID, "user_id", cmd.UserID)

	txCtx := &domain.TransactionContext{
		TransactionID:     cmd.TransactionID,
		UserID:            cmd.UserID,
		Amount:            cmd.Amount,
		Currency:          cmd.Currency,
		PaymentMethod:     cmd.PaymentMethod,
		IPAddress:         cmd.IPAddress,
		DeviceFingerprint: cmd.DeviceFingerprint,
		UserAgent:         cmd.UserAgent,
		Country:           cmd.Country,
		City:              cmd.City,
		Latitude:          cmd.Latitude,
		Longitude:         cmd.Longitude,
		MerchantID:        cmd.MerchantID,
		CardLastFour:      cmd.CardLastFour,
		Email:             cmd.Email,
		Phone:             cmd.Phone,
		Metadata:          cmd.Metadata,
	}

	assessment := domain.NewRiskAssessment(cmd.TransactionID, cmd.UserID)
	assessment.ID = fmt.Sprintf("RA%d", s.idGenerator.Generate())

	if isBlacklisted, _ := s.checkBlacklist(ctx, txCtx); isBlacklisted {
		assessment.AddRiskFactor(&domain.RiskFactor{
			FactorName:       "BLACKLISTED",
			Description:      "User or device is in blacklist",
			Weight:           100,
			ContributionScore: 100,
		})
	}

	s.checkDeviceFingerprint(ctx, txCtx, assessment)
	s.checkUserRiskProfile(ctx, txCtx, assessment)

	rules, err := s.ruleRepo.FindEnabled(ctx)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to load risk rules", "error", err)
	} else {
		for _, rule := range rules {
			triggered, score, err := s.ruleEngine.Evaluate(ctx, rule, txCtx)
			if err != nil {
				s.logger.WarnContext(ctx, "rule evaluation failed", "rule_id", rule.ID, "error", err)
				continue
			}
			if triggered {
				assessment.AddTriggeredRule(rule.ID)
				assessment.AddRiskFactor(&domain.RiskFactor{
					FactorName:       rule.Name,
					Description:      rule.Description,
					Weight:           rule.RiskWeight,
					ContributionScore: score,
				})
			}
		}
	}

	assessment.Explanation = s.generateExplanation(assessment)

	if err := s.assessmentRepo.Save(ctx, assessment); err != nil {
		return nil, fmt.Errorf("failed to save assessment: %w", err)
	}

	s.logger.InfoContext(ctx, "transaction analysis completed",
		"transaction_id", cmd.TransactionID,
		"risk_level", assessment.RiskLevel.String(),
		"risk_score", assessment.RiskScore,
		"action", assessment.RecommendedAction)

	return assessment, nil
}

func (s *FraudDetectionService) checkBlacklist(ctx context.Context, txCtx *domain.TransactionContext) (bool, error) {
	checks := []struct {
		blacklistType domain.BlacklistType
		value         string
	}{
		{domain.BlacklistTypeUser, fmt.Sprintf("%d", txCtx.UserID)},
		{domain.BlacklistTypeIP, txCtx.IPAddress},
		{domain.BlacklistTypeDevice, txCtx.DeviceFingerprint},
		{domain.BlacklistTypeEmail, txCtx.Email},
		{domain.BlacklistTypePhone, txCtx.Phone},
	}

	for _, check := range checks {
		if check.value == "" {
			continue
		}
		exists, err := s.blacklistRepo.Exists(ctx, check.blacklistType, check.value)
		if err != nil {
			s.logger.WarnContext(ctx, "blacklist check failed", "type", check.blacklistType, "error", err)
			continue
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

func (s *FraudDetectionService) checkDeviceFingerprint(ctx context.Context, txCtx *domain.TransactionContext, assessment *domain.RiskAssessment) {
	if txCtx.DeviceFingerprint == "" {
		return
	}

	device, err := s.deviceRepo.FindByFingerprint(ctx, txCtx.DeviceFingerprint)
	if err != nil || device == nil {
		return
	}

	if device.IsSuspicious {
		assessment.AddRiskFactor(&domain.RiskFactor{
			FactorName:       "SUSPICIOUS_DEVICE",
			Description:      fmt.Sprintf("Device associated with %d accounts", device.AssociatedAccountsCount),
			Weight:           30,
			ContributionScore: 30,
		})
	}

	if device.AssociatedAccountsCount > 3 {
		assessment.AddRiskFactor(&domain.RiskFactor{
			FactorName:       "MULTI_ACCOUNT_DEVICE",
			Description:      fmt.Sprintf("Device used by %d different accounts", device.AssociatedAccountsCount),
			Weight:           20,
			ContributionScore: 20,
		})
	}
}

func (s *FraudDetectionService) checkUserRiskProfile(ctx context.Context, txCtx *domain.TransactionContext, assessment *domain.RiskAssessment) {
	profile, err := s.profileRepo.FindByUserID(ctx, txCtx.UserID)
	if err != nil || profile == nil {
		return
	}

	if profile.RiskLevel >= domain.RiskLevelHigh {
		assessment.AddRiskFactor(&domain.RiskFactor{
			FactorName:       "HIGH_RISK_USER",
			Description:      fmt.Sprintf("User has risk level: %s", profile.RiskLevel.String()),
			Weight:           int(profile.RiskLevel) * 15,
			ContributionScore: int(profile.RiskLevel) * 15,
		})
	}

	if len(profile.RiskTags) > 0 {
		assessment.AddRiskFactor(&domain.RiskFactor{
			FactorName:       "RISK_TAGS",
			Description:      fmt.Sprintf("User has risk tags: %v", profile.RiskTags),
			Weight:           len(profile.RiskTags) * 5,
			ContributionScore: len(profile.RiskTags) * 5,
		})
	}
}

func (s *FraudDetectionService) generateExplanation(assessment *domain.RiskAssessment) string {
	if len(assessment.RiskFactors) == 0 {
		return "No significant risk factors detected"
	}

	explanation := fmt.Sprintf("Risk assessment completed with score %d. ", assessment.RiskScore)
	explanation += fmt.Sprintf("Risk level: %s. ", assessment.RiskLevel.String())
	explanation += fmt.Sprintf("Recommended action: %s. ", assessment.RecommendedAction)

	if len(assessment.TriggeredRules) > 0 {
		explanation += fmt.Sprintf("Triggered %d risk rules. ", len(assessment.TriggeredRules))
	}

	return explanation
}

func (s *FraudDetectionService) GetRiskAssessment(ctx context.Context, assessmentID string) (*domain.RiskAssessment, error) {
	return s.assessmentRepo.FindByID(ctx, assessmentID)
}

type AddToBlacklistCommand struct {
	Type        domain.BlacklistType
	Value       string
	Reason      string
	CreatedBy   uint64
	ExpiresIn   *time.Duration
}

func (s *FraudDetectionService) AddToBlacklist(ctx context.Context, cmd *AddToBlacklistCommand) (*domain.BlacklistEntry, error) {
	entry := domain.NewBlacklistEntry(cmd.Type, cmd.Value, cmd.Reason, cmd.CreatedBy, cmd.ExpiresIn)
	entry.ID = fmt.Sprintf("BL%d", s.idGenerator.Generate())

	if err := s.blacklistRepo.Save(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to add to blacklist: %w", err)
	}

	s.logger.InfoContext(ctx, "added to blacklist", "id", entry.ID, "type", cmd.Type, "value", cmd.Value)
	return entry, nil
}

func (s *FraudDetectionService) RemoveFromBlacklist(ctx context.Context, id string) error {
	if err := s.blacklistRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to remove from blacklist: %w", err)
	}
	s.logger.InfoContext(ctx, "removed from blacklist", "id", id)
	return nil
}

func (s *FraudDetectionService) GetBlacklist(ctx context.Context, blacklistType domain.BlacklistType, page, pageSize int) ([]*domain.BlacklistEntry, int64, error) {
	offset := (page - 1) * pageSize
	return s.blacklistRepo.FindByType(ctx, blacklistType, pageSize, offset)
}

type CreateRiskRuleCommand struct {
	Name       string
	Type       domain.RuleType
	Condition  string
	RiskWeight int
	Action     domain.RiskAction
	Priority   int
}

func (s *FraudDetectionService) CreateRiskRule(ctx context.Context, cmd *CreateRiskRuleCommand) (*domain.RiskRule, error) {
	rule := domain.NewRiskRule(cmd.Name, cmd.Type, cmd.Condition, cmd.RiskWeight, cmd.Action)
	rule.ID = fmt.Sprintf("RR%d", s.idGenerator.Generate())
	if cmd.Priority > 0 {
		rule.Priority = cmd.Priority
	}

	if err := s.ruleRepo.Save(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create risk rule: %w", err)
	}

	s.logger.InfoContext(ctx, "created risk rule", "id", rule.ID, "name", cmd.Name)
	return rule, nil
}

func (s *FraudDetectionService) GetRiskRules(ctx context.Context, ruleType domain.RuleType, enabledOnly bool) ([]*domain.RiskRule, error) {
	if enabledOnly {
		return s.ruleRepo.FindEnabled(ctx)
	}
	if ruleType != "" {
		return s.ruleRepo.FindByType(ctx, ruleType)
	}
	return s.ruleRepo.FindAll(ctx)
}

func (s *FraudDetectionService) GetDeviceFingerprint(ctx context.Context, fingerprint string) (*domain.DeviceFingerprint, error) {
	return s.deviceRepo.FindByFingerprint(ctx, fingerprint)
}

func (s *FraudDetectionService) GetUserRiskProfile(ctx context.Context, userID uint64) (*domain.UserRiskProfile, error) {
	return s.profileRepo.FindByUserID(ctx, userID)
}

type RecordSuspiciousActivityCommand struct {
	UserID            uint64
	ActivityType      string
	Description       string
	IPAddress         string
	DeviceFingerprint string
	Metadata          map[string]string
}

func (s *FraudDetectionService) RecordSuspiciousActivity(ctx context.Context, cmd *RecordSuspiciousActivityCommand) error {
	activity := &domain.SuspiciousActivity{
		ID:                fmt.Sprintf("SA%d", s.idGenerator.Generate()),
		UserID:            cmd.UserID,
		ActivityType:      cmd.ActivityType,
		Description:       cmd.Description,
		IPAddress:         cmd.IPAddress,
		DeviceFingerprint: cmd.DeviceFingerprint,
		Metadata:          cmd.Metadata,
	}

	if err := s.activityRepo.Save(ctx, activity); err != nil {
		return fmt.Errorf("failed to record suspicious activity: %w", err)
	}

	s.logger.WarnContext(ctx, "recorded suspicious activity",
		"user_id", cmd.UserID,
		"type", cmd.ActivityType,
		"description", cmd.Description)

	return nil
}
