package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/riskanalyzer/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

type RiskCommandService struct {
	repo        domain.RiskRepository
	creditRepo  domain.CreditProfileRepository
	fraudEngine domain.FraudEngine
	creditEval  domain.CreditEvaluator
	riskBridge  domain.RiskSynergyService
	publisher   messagequeue.EventPublisher
	topic       string
	logger      *slog.Logger
}

func NewRiskCommandService(
	repo domain.RiskRepository,
	creditRepo domain.CreditProfileRepository,
	fraudEngine domain.FraudEngine,
	creditEval domain.CreditEvaluator,
	riskBridge domain.RiskSynergyService,
	publisher messagequeue.EventPublisher,
	topic string,
	logger *slog.Logger,
) *RiskCommandService {
	return &RiskCommandService{
		repo:        repo,
		creditRepo:  creditRepo,
		fraudEngine: fraudEngine,
		creditEval:  creditEval,
		riskBridge:  riskBridge,
		publisher:   publisher,
		topic:       topic,
		logger:      logger,
	}
}

type AnalyzeRiskCommand struct {
	TargetID     string
	TargetType   string
	UserID       uint64
	DeviceFinger string
	IPAddress    string
	UserAgent    string
}

func (s *RiskCommandService) AnalyzeRisk(ctx context.Context, cmd AnalyzeRiskCommand) (*domain.RiskAssessment, error) {
	s.logger.InfoContext(ctx, "analyzing risk", "target_id", cmd.TargetID, "user_id", cmd.UserID)

	assessment := &domain.RiskAssessment{
		ID:           uint64(idgen.GenID()),
		CreatedAt:    time.Now(),
		TargetID:     cmd.TargetID,
		TargetType:   cmd.TargetType,
		UserID:       cmd.UserID,
		DeviceFinger: cmd.DeviceFinger,
		IPAddress:    cmd.IPAddress,
	}

	inBlacklist, err := s.repo.IsInBlacklist(ctx, domain.BlacklistUser, fmt.Sprintf("%d", cmd.UserID))
	if err != nil {
		s.logger.WarnContext(ctx, "failed to check blacklist", "error", err)
	}
	if inBlacklist {
		assessment.TotalScore = 100
		assessment.Level = domain.RiskCritical
		assessment.Decision = "REJECT"
		assessment.MatchedRules = append(assessment.MatchedRules, &domain.MatchedRule{
			RuleID:   0,
			RuleName: "BLACKLIST_USER",
			Score:    100,
			Reason:   "User is in blacklist",
		})
	} else {
		if s.fraudEngine != nil {
			assessment, err = s.fraudEngine.Analyze(ctx, assessment)
			if err != nil {
				s.logger.ErrorContext(ctx, "fraud engine analysis failed", "error", err)
				assessment.TotalScore = 50
				assessment.Level = domain.RiskMedium
				assessment.Decision = "PASS"
			}
		} else {
			assessment.TotalScore = 10
			assessment.Level = domain.RiskLow
			assessment.Decision = "PASS"
		}
	}

	if err := s.repo.SaveAssessment(ctx, assessment); err != nil {
		return nil, fmt.Errorf("failed to save assessment: %w", err)
	}

	if s.riskBridge != nil && assessment.Level >= domain.RiskHigh {
		signal := &domain.SharedRiskSignal{
			SignalID:   fmt.Sprintf("RISK-%d", assessment.ID),
			UserID:     fmt.Sprintf("%d", cmd.UserID),
			Source:     domain.ScopeEcommerce,
			RiskType:   "FRAUD_DETECTED",
			Level:      assessment.Level,
			Score:      float64(assessment.TotalScore),
			Details:    fmt.Sprintf("Risk assessment triggered for %s", cmd.TargetType),
			DetectedAt: time.Now(),
		}
		if err := s.riskBridge.ReportSignal(ctx, signal); err != nil {
			s.logger.WarnContext(ctx, "failed to report risk signal", "error", err)
		}
	}

	s.logger.InfoContext(ctx, "risk analysis completed", "assessment_id", assessment.ID, "score", assessment.TotalScore, "decision", assessment.Decision)
	return assessment, nil
}

type AddToBlacklistCommand struct {
	Type   domain.BlacklistType
	Value  string
	Reason string
	Source string
}

func (s *RiskCommandService) AddToBlacklist(ctx context.Context, cmd AddToBlacklistCommand) error {
	s.logger.InfoContext(ctx, "adding to blacklist", "type", cmd.Type, "value", cmd.Value)

	item := &domain.Blacklist{
		ID:        uint64(idgen.GenID()),
		Type:      cmd.Type,
		Value:     cmd.Value,
		Reason:    cmd.Reason,
		Source:    cmd.Source,
		CreatedAt: time.Now(),
	}

	return s.repo.AddToBlacklist(ctx, item)
}

type RemoveFromBlacklistCommand struct {
	Type  domain.BlacklistType
	Value string
}

func (s *RiskCommandService) RemoveFromBlacklist(ctx context.Context, cmd RemoveFromBlacklistCommand) error {
	s.logger.InfoContext(ctx, "removing from blacklist", "type", cmd.Type, "value", cmd.Value)
	return s.repo.RemoveFromBlacklist(ctx, cmd.Type, cmd.Value)
}

type EvaluateCreditCommand struct {
	UserID uint64
}

func (s *RiskCommandService) EvaluateCredit(ctx context.Context, cmd EvaluateCreditCommand) (*domain.CreditProfile, error) {
	s.logger.InfoContext(ctx, "evaluating credit", "user_id", cmd.UserID)

	if s.creditEval == nil {
		return nil, errors.New("credit evaluator not configured")
	}

	profile, err := s.creditEval.Evaluate(ctx, cmd.UserID)
	if err != nil {
		return nil, fmt.Errorf("credit evaluation failed: %w", err)
	}

	profile.DetermineLevel()

	if err := s.creditRepo.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save credit profile: %w", err)
	}

	s.logger.InfoContext(ctx, "credit evaluation completed", "user_id", cmd.UserID, "score", profile.Score, "level", profile.Level)
	return profile, nil
}

type FreezeCreditLimitCommand struct {
	UserID uint64
	Amount int64
}

func (s *RiskCommandService) FreezeCreditLimit(ctx context.Context, cmd FreezeCreditLimitCommand) error {
	s.logger.InfoContext(ctx, "freezing credit limit", "user_id", cmd.UserID, "amount", cmd.Amount)

	profile, err := s.creditRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find credit profile: %w", err)
	}
	if profile == nil {
		return errors.New("credit profile not found")
	}

	if err := profile.FreezeLimit(cmd.Amount); err != nil {
		return fmt.Errorf("failed to freeze limit: %w", err)
	}

	return s.creditRepo.Save(ctx, profile)
}

type ReleaseCreditLimitCommand struct {
	UserID uint64
	Amount int64
}

func (s *RiskCommandService) ReleaseCreditLimit(ctx context.Context, cmd ReleaseCreditLimitCommand) error {
	s.logger.InfoContext(ctx, "releasing credit limit", "user_id", cmd.UserID, "amount", cmd.Amount)

	profile, err := s.creditRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find credit profile: %w", err)
	}
	if profile == nil {
		return errors.New("credit profile not found")
	}

	profile.ReleaseLimit(cmd.Amount)
	return s.creditRepo.Save(ctx, profile)
}

type AddCreditRecordCommand struct {
	UserID       uint64
	RecordType   domain.CreditRecordType
	ScoreChange  int
	LimitChange  int64
	RelatedID    string
	Reason       string
}

func (s *RiskCommandService) AddCreditRecord(ctx context.Context, cmd AddCreditRecordCommand) error {
	s.logger.InfoContext(ctx, "adding credit record", "user_id", cmd.UserID, "type", cmd.RecordType)

	profile, err := s.creditRepo.FindByUserID(ctx, cmd.UserID)
	if err != nil {
		return fmt.Errorf("failed to find credit profile: %w", err)
	}
	if profile == nil {
		return errors.New("credit profile not found")
	}

	record := &domain.CreditRecord{
		ID:          uint64(idgen.GenID()),
		ProfileID:   profile.ID,
		Type:        cmd.RecordType,
		ScoreChange: cmd.ScoreChange,
		LimitChange: cmd.LimitChange,
		RelatedID:   cmd.RelatedID,
		Reason:      cmd.Reason,
		HappenedAt:  time.Now(),
	}

	profile.Score += cmd.ScoreChange
	if profile.Score < 300 {
		profile.Score = 300
	}
	if profile.Score > 850 {
		profile.Score = 850
	}
	profile.DetermineLevel()

	if cmd.LimitChange != 0 {
		profile.TotalLimit += cmd.LimitChange
		profile.AvailableLimit += cmd.LimitChange
	}

	if err := s.creditRepo.Save(ctx, profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return s.creditRepo.AddRecord(ctx, record)
}

type RiskQueryService struct {
	repo       domain.RiskRepository
	creditRepo domain.CreditProfileRepository
	logger     *slog.Logger
}

func NewRiskQueryService(
	repo domain.RiskRepository,
	creditRepo domain.CreditProfileRepository,
	logger *slog.Logger,
) *RiskQueryService {
	return &RiskQueryService{
		repo:       repo,
		creditRepo: creditRepo,
		logger:     logger,
	}
}

func (s *RiskQueryService) GetAssessment(ctx context.Context, targetID string) (*domain.RiskAssessment, error) {
	return s.repo.FindAssessment(ctx, targetID)
}

func (s *RiskQueryService) GetCreditProfile(ctx context.Context, userID uint64) (*domain.CreditProfile, error) {
	return s.creditRepo.FindByUserID(ctx, userID)
}

func (s *RiskQueryService) GetCreditRecords(ctx context.Context, userID uint64, limit int) ([]*domain.CreditRecord, error) {
	profile, err := s.creditRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	return s.creditRepo.FindRecords(ctx, profile.ID, limit)
}

func (s *RiskQueryService) IsInBlacklist(ctx context.Context, blacklistType domain.BlacklistType, value string) (bool, error) {
	return s.repo.IsInBlacklist(ctx, blacklistType, value)
}

func (s *RiskQueryService) GetActiveRules(ctx context.Context, targetType string) ([]*domain.RiskRule, error) {
	return s.repo.FindActiveRules(ctx, targetType)
}
