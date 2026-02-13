package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrRiskAssessmentNotFound = errors.New("risk assessment not found")
	ErrBlacklistEntryNotFound = errors.New("blacklist entry not found")
	ErrRiskRuleNotFound       = errors.New("risk rule not found")
	ErrInvalidRiskScore       = errors.New("invalid risk score")
	ErrRuleConditionInvalid   = errors.New("rule condition is invalid")
)

type RiskLevel int

const (
	RiskLevelLow      RiskLevel = 1
	RiskLevelMedium   RiskLevel = 2
	RiskLevelHigh     RiskLevel = 3
	RiskLevelCritical RiskLevel = 4
)

func (r RiskLevel) String() string {
	switch r {
	case RiskLevelLow:
		return "LOW"
	case RiskLevelMedium:
		return "MEDIUM"
	case RiskLevelHigh:
		return "HIGH"
	case RiskLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

type RiskAction string

const (
	RiskActionAllow    RiskAction = "ALLOW"
	RiskActionChallenge RiskAction = "CHALLENGE"
	RiskActionReview   RiskAction = "REVIEW"
	RiskActionBlock    RiskAction = "BLOCK"
)

type BlacklistType string

const (
	BlacklistTypeUser   BlacklistType = "USER"
	BlacklistTypeDevice BlacklistType = "DEVICE"
	BlacklistTypeIP     BlacklistType = "IP"
	BlacklistTypeCard   BlacklistType = "CARD"
	BlacklistTypePhone  BlacklistType = "PHONE"
	BlacklistTypeEmail  BlacklistType = "EMAIL"
)

type RuleType string

const (
	RuleTypeVelocity  RuleType = "VELOCITY"
	RuleTypeAmount    RuleType = "AMOUNT"
	RuleTypeGeography RuleType = "GEOGRAPHY"
	RuleTypeBehavior  RuleType = "BEHAVIOR"
	RuleTypeDevice    RuleType = "DEVICE"
	RuleTypeCustom    RuleType = "CUSTOM"
)

type RiskAssessment struct {
	ID              string        `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	TransactionID   string        `json:"transaction_id"`
	UserID          uint64        `json:"user_id"`
	RiskLevel       RiskLevel     `json:"risk_level"`
	RiskScore       int           `json:"risk_score"`
	RecommendedAction RiskAction  `json:"recommended_action"`
	RiskFactors     []*RiskFactor `json:"risk_factors"`
	TriggeredRules  []string      `json:"triggered_rules"`
	Explanation     string        `json:"explanation"`
}

type RiskFactor struct {
	FactorName       string `json:"factor_name"`
	Description      string `json:"description"`
	Weight           int    `json:"weight"`
	ContributionScore int   `json:"contribution_score"`
}

type BlacklistEntry struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Type      BlacklistType `json:"type"`
	Value     string        `json:"value"`
	Reason    string        `json:"reason"`
	CreatedBy uint64        `json:"created_by"`
	ExpiresAt *time.Time    `json:"expires_at"`
	IsActive  bool          `json:"is_active"`
}

type RiskRule struct {
	ID          string     `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        RuleType   `json:"type"`
	Condition   string     `json:"condition"`
	RiskWeight  int        `json:"risk_weight"`
	Action      RiskAction `json:"action"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
}

type DeviceFingerprint struct {
	Fingerprint            string    `json:"fingerprint"`
	FirstSeenUserID        uint64    `json:"first_seen_user_id"`
	AssociatedAccounts     []uint64  `json:"associated_accounts"`
	AssociatedIPs          []string  `json:"associated_ips"`
	IsSuspicious           bool      `json:"is_suspicious"`
	DeviceType             string    `json:"device_type"`
	OS                     string    `json:"os"`
	Browser                string    `json:"browser"`
	FirstSeenAt            time.Time `json:"first_seen_at"`
	LastSeenAt             time.Time `json:"last_seen_at"`
	AssociatedAccountsCount int       `json:"associated_accounts_count"`
}

type UserRiskProfile struct {
	UserID              uint64    `json:"user_id"`
	RiskScore           int       `json:"risk_score"`
	RiskLevel           RiskLevel `json:"risk_level"`
	TotalTransactions   int       `json:"total_transactions"`
	FlaggedTransactions int       `json:"flagged_transactions"`
	BlockedTransactions int       `json:"blocked_transactions"`
	RiskTags            []string  `json:"risk_tags"`
	FirstTransactionAt  *time.Time `json:"first_transaction_at"`
	LastTransactionAt   *time.Time `json:"last_transaction_at"`
	LastAssessmentAt    *time.Time `json:"last_assessment_at"`
}

type TransactionContext struct {
	TransactionID   string            `json:"transaction_id"`
	UserID          uint64            `json:"user_id"`
	Amount          int64             `json:"amount"`
	Currency        string            `json:"currency"`
	PaymentMethod   string            `json:"payment_method"`
	IPAddress       string            `json:"ip_address"`
	DeviceFingerprint string          `json:"device_fingerprint"`
	UserAgent       string            `json:"user_agent"`
	Country         string            `json:"country"`
	City            string            `json:"city"`
	Latitude        float64           `json:"latitude"`
	Longitude       float64           `json:"longitude"`
	MerchantID      uint64            `json:"merchant_id"`
	CardLastFour    string            `json:"card_last_four"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Metadata        map[string]string `json:"metadata"`
}

type SuspiciousActivity struct {
	ID                string            `json:"id"`
	CreatedAt         time.Time         `json:"created_at"`
	UserID            uint64            `json:"user_id"`
	ActivityType      string            `json:"activity_type"`
	Description       string            `json:"description"`
	IPAddress         string            `json:"ip_address"`
	DeviceFingerprint string            `json:"device_fingerprint"`
	Metadata          map[string]string `json:"metadata"`
}

func NewRiskAssessment(transactionID string, userID uint64) *RiskAssessment {
	return &RiskAssessment{
		TransactionID:  transactionID,
		UserID:         userID,
		RiskLevel:      RiskLevelLow,
		RiskScore:      0,
		RiskFactors:    make([]*RiskFactor, 0),
		TriggeredRules: make([]string, 0),
	}
}

func (r *RiskAssessment) AddRiskFactor(factor *RiskFactor) {
	r.RiskFactors = append(r.RiskFactors, factor)
	r.RiskScore += factor.ContributionScore
	r.updateRiskLevel()
}

func (r *RiskAssessment) AddTriggeredRule(ruleID string) {
	r.TriggeredRules = append(r.TriggeredRules, ruleID)
}

func (r *RiskAssessment) updateRiskLevel() {
	switch {
	case r.RiskScore >= 80:
		r.RiskLevel = RiskLevelCritical
		r.RecommendedAction = RiskActionBlock
	case r.RiskScore >= 60:
		r.RiskLevel = RiskLevelHigh
		r.RecommendedAction = RiskActionReview
	case r.RiskScore >= 40:
		r.RiskLevel = RiskLevelMedium
		r.RecommendedAction = RiskActionChallenge
	default:
		r.RiskLevel = RiskLevelLow
		r.RecommendedAction = RiskActionAllow
	}
}

func NewBlacklistEntry(blacklistType BlacklistType, value, reason string, createdBy uint64, expiresIn *time.Duration) *BlacklistEntry {
	entry := &BlacklistEntry{
		Type:      blacklistType,
		Value:     value,
		Reason:    reason,
		CreatedBy: createdBy,
		IsActive:  true,
	}
	if expiresIn != nil {
		expiresAt := time.Now().Add(*expiresIn)
		entry.ExpiresAt = &expiresAt
	}
	return entry
}

func (b *BlacklistEntry) IsExpired() bool {
	if b.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*b.ExpiresAt)
}

func NewRiskRule(name string, ruleType RuleType, condition string, riskWeight int, action RiskAction) *RiskRule {
	return &RiskRule{
		Name:       name,
		Type:       ruleType,
		Condition:  condition,
		RiskWeight: riskWeight,
		Action:     action,
		Enabled:    true,
		Priority:   100,
	}
}

func NewDeviceFingerprint(fingerprint string, userID uint64, deviceType, os, browser string) *DeviceFingerprint {
	now := time.Now()
	return &DeviceFingerprint{
		Fingerprint:         fingerprint,
		FirstSeenUserID:     userID,
		AssociatedAccounts:  []uint64{userID},
		AssociatedIPs:       []string{},
		IsSuspicious:        false,
		DeviceType:          deviceType,
		OS:                  os,
		Browser:             browser,
		FirstSeenAt:         now,
		LastSeenAt:          now,
		AssociatedAccountsCount: 1,
	}
}

func (d *DeviceFingerprint) AddAssociatedAccount(userID uint64) {
	for _, id := range d.AssociatedAccounts {
		if id == userID {
			return
		}
	}
	d.AssociatedAccounts = append(d.AssociatedAccounts, userID)
	d.AssociatedAccountsCount = len(d.AssociatedAccounts)
	d.LastSeenAt = time.Now()
	
	if d.AssociatedAccountsCount > 5 {
		d.IsSuspicious = true
	}
}

func (d *DeviceFingerprint) AddAssociatedIP(ip string) {
	for _, existingIP := range d.AssociatedIPs {
		if existingIP == ip {
			return
		}
	}
	d.AssociatedIPs = append(d.AssociatedIPs, ip)
	d.LastSeenAt = time.Now()
}

func NewUserRiskProfile(userID uint64) *UserRiskProfile {
	return &UserRiskProfile{
		UserID:            userID,
		RiskScore:         0,
		RiskLevel:         RiskLevelLow,
		TotalTransactions: 0,
		FlaggedTransactions: 0,
		BlockedTransactions: 0,
		RiskTags:          []string{},
	}
}

func (p *UserRiskProfile) RecordTransaction(flagged, blocked bool) {
	p.TotalTransactions++
	if flagged {
		p.FlaggedTransactions++
	}
	if blocked {
		p.BlockedTransactions++
	}
	now := time.Now()
	if p.FirstTransactionAt == nil {
		p.FirstTransactionAt = &now
	}
	p.LastTransactionAt = &now
	p.updateRiskScore()
}

func (p *UserRiskProfile) updateRiskScore() {
	if p.TotalTransactions == 0 {
		p.RiskScore = 0
		return
	}
	
	blockRate := float64(p.BlockedTransactions) / float64(p.TotalTransactions) * 100
	flagRate := float64(p.FlaggedTransactions) / float64(p.TotalTransactions) * 100
	
	p.RiskScore = int(blockRate*50 + flagRate*30)
	
	switch {
	case p.RiskScore >= 70:
		p.RiskLevel = RiskLevelCritical
	case p.RiskScore >= 50:
		p.RiskLevel = RiskLevelHigh
	case p.RiskScore >= 30:
		p.RiskLevel = RiskLevelMedium
	default:
		p.RiskLevel = RiskLevelLow
	}
}

func (p *UserRiskProfile) AddRiskTag(tag string) {
	for _, t := range p.RiskTags {
		if t == tag {
			return
		}
	}
	p.RiskTags = append(p.RiskTags, tag)
}

type RiskAssessmentRepository interface {
	Save(ctx context.Context, assessment *RiskAssessment) error
	FindByID(ctx context.Context, id string) (*RiskAssessment, error)
	FindByTransactionID(ctx context.Context, transactionID string) (*RiskAssessment, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*RiskAssessment, error)
}

type BlacklistRepository interface {
	Save(ctx context.Context, entry *BlacklistEntry) error
	FindByID(ctx context.Context, id string) (*BlacklistEntry, error)
	FindByTypeAndValue(ctx context.Context, blacklistType BlacklistType, value string) (*BlacklistEntry, error)
	FindByType(ctx context.Context, blacklistType BlacklistType, limit, offset int) ([]*BlacklistEntry, int64, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, blacklistType BlacklistType, value string) (bool, error)
}

type RiskRuleRepository interface {
	Save(ctx context.Context, rule *RiskRule) error
	FindByID(ctx context.Context, id string) (*RiskRule, error)
	FindAll(ctx context.Context) ([]*RiskRule, error)
	FindByType(ctx context.Context, ruleType RuleType) ([]*RiskRule, error)
	FindEnabled(ctx context.Context) ([]*RiskRule, error)
	Update(ctx context.Context, rule *RiskRule) error
	Delete(ctx context.Context, id string) error
}

type DeviceFingerprintRepository interface {
	Save(ctx context.Context, fingerprint *DeviceFingerprint) error
	FindByFingerprint(ctx context.Context, fingerprint string) (*DeviceFingerprint, error)
}

type UserRiskProfileRepository interface {
	Save(ctx context.Context, profile *UserRiskProfile) error
	FindByUserID(ctx context.Context, userID uint64) (*UserRiskProfile, error)
}

type SuspiciousActivityRepository interface {
	Save(ctx context.Context, activity *SuspiciousActivity) error
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*SuspiciousActivity, error)
}

type RuleEngine interface {
	Evaluate(ctx context.Context, rule *RiskRule, txCtx *TransactionContext) (bool, int, error)
}

type RiskAnalyzer interface {
	Analyze(ctx context.Context, txCtx *TransactionContext) (*RiskAssessment, error)
}
