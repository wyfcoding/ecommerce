package application

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/insurance/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

type InsuranceService struct {
	repo     domain.Repository
	eventPub messagequeue.EventPublisher
	logger   *slog.Logger
}

func NewInsuranceService(repo domain.Repository, eventPub messagequeue.EventPublisher, logger *slog.Logger) *InsuranceService {
	return &InsuranceService{
		repo:     repo,
		eventPub: eventPub,
		logger:   logger,
	}
}

// CreatePolicyRequest 请求参数
type CreatePolicyRequest struct {
	OrderID        string
	UserID         string
	Type           domain.PolicyType
	Premium        float64
	CoverageAmount float64
	DurationDays   int
}

func (s *InsuranceService) CreatePolicy(ctx context.Context, req CreatePolicyRequest) (*domain.InsurancePolicy, error) {
	now := time.Now()
	policy := &domain.InsurancePolicy{
		PolicyID:       fmt.Sprintf("POL-%s-%d", req.OrderID, now.UnixNano()),
		OrderID:        req.OrderID,
		UserID:         req.UserID,
		Type:           req.Type,
		Premium:        decimal.NewFromFloat(req.Premium),
		CoverageAmount: decimal.NewFromFloat(req.CoverageAmount),
		Status:         domain.PolicyStatusActive,
		StartTime:      now,
		EndTime:        now.AddDate(0, 0, req.DurationDays),
	}

	if err := s.repo.SavePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to save policy: %w", err)
	}

	event := &domain.PolicyCreatedEvent{
		PolicyID:  policy.PolicyID,
		OrderID:   policy.OrderID,
		UserID:    policy.UserID,
		Timestamp: now,
	}
	if err := s.eventPub.Publish(ctx, event.EventName(), policy.PolicyID, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish policy created event", "error", err)
	}

	s.logger.InfoContext(ctx, "policy created", "policy_id", policy.PolicyID)
	return policy, nil
}

func (s *InsuranceService) GetPolicy(ctx context.Context, policyID string) (*domain.InsurancePolicy, error) {
	return s.repo.GetPolicy(ctx, policyID)
}

// FileClaimRequest 理赔请求参数
type FileClaimRequest struct {
	PolicyID     string
	UserID       string
	Reason       string
	Amount       float64
	EvidenceURLs []string
}

func (s *InsuranceService) FileClaim(ctx context.Context, req FileClaimRequest) (*domain.InsuranceClaim, error) {
	// 1. Check policy
	policy, err := s.repo.GetPolicy(ctx, req.PolicyID)
	if err != nil {
		return nil, err
	}

	if policy.UserID != req.UserID {
		return nil, fmt.Errorf("user %s is not authorized for policy %s", req.UserID, req.PolicyID)
	}

	if policy.Status != domain.PolicyStatusActive {
		return nil, fmt.Errorf("policy %s is not active", req.PolicyID)
	}

	if time.Now().After(policy.EndTime) {
		return nil, fmt.Errorf("policy %s has expired", req.PolicyID)
	}

	// 2. Create claim
	now := time.Now()
	claim := &domain.InsuranceClaim{
		ClaimID:         fmt.Sprintf("CLM-%s-%d", req.PolicyID, now.UnixNano()),
		PolicyID:        req.PolicyID,
		UserID:          req.UserID,
		Reason:          req.Reason,
		AmountRequested: decimal.NewFromFloat(req.Amount),
		Status:          domain.ClaimStatusSubmitted,
		// EvidenceURLs could be joined or stored as JSON, simplified here to string
		EvidenceURLs: fmt.Sprintf("%v", req.EvidenceURLs),
	}

	if err := s.repo.SaveClaim(ctx, claim); err != nil {
		return nil, fmt.Errorf("failed to save claim: %w", err)
	}

	// 3. Publish event
	event := &domain.ClaimFiledEvent{
		ClaimID:   claim.ClaimID,
		PolicyID:  claim.PolicyID,
		UserID:    claim.UserID,
		Timestamp: now,
	}
	if err := s.eventPub.Publish(ctx, event.EventName(), claim.ClaimID, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish claim filed event", "error", err)
	}

	s.logger.InfoContext(ctx, "claim filed", "claim_id", claim.ClaimID)
	return claim, nil
}

func (s *InsuranceService) GetClaim(ctx context.Context, claimID string) (*domain.InsuranceClaim, error) {
	return s.repo.GetClaim(ctx, claimID)
}

func (s *InsuranceService) ListPolicies(ctx context.Context, userID, orderID string, pageSize int, pageToken string) ([]*domain.InsurancePolicy, string, error) {
	offset := 0
	if pageToken != "" {
		if parsed, err := strconv.Atoi(pageToken); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	limit := pageSize
	if limit <= 0 {
		limit = 10
	}

	policies, total, err := s.repo.ListPolicies(ctx, userID, "", offset, limit)
	if err != nil {
		return nil, "", err
	}
	if orderID != "" {
		filtered := make([]*domain.InsurancePolicy, 0, len(policies))
		for _, policy := range policies {
			if policy.OrderID == orderID {
				filtered = append(filtered, policy)
			}
		}
		policies = filtered
	}

	nextPageToken := ""
	nextOffset := offset + len(policies)
	if int64(nextOffset) < total {
		nextPageToken = strconv.Itoa(nextOffset)
	}

	return policies, nextPageToken, nil
}
