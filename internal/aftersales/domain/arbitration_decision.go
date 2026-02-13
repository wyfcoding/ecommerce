package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDecisionNotFound      = errors.New("decision not found")
	ErrDecisionAlreadyMade   = errors.New("decision already made")
	ErrEvidenceNotVerified   = errors.New("evidence not verified")
	ErrInvalidDecisionType   = errors.New("invalid decision type")
)

type DecisionType int8

const (
	DecisionTypeFullRefund      DecisionType = 1
	DecisionTypePartialRefund   DecisionType = 2
	DecisionTypeReturnRefund    DecisionType = 3
	DecisionTypeExchange        DecisionType = 4
	DecisionTypeRepair          DecisionType = 5
	DecisionTypeCompensation    DecisionType = 6
	DecisionTypeReject          DecisionType = 7
	DecisionTypeNoAction        DecisionType = 8
)

func (t DecisionType) String() string {
	switch t {
	case DecisionTypeFullRefund:
		return "FULL_REFUND"
	case DecisionTypePartialRefund:
		return "PARTIAL_REFUND"
	case DecisionTypeReturnRefund:
		return "RETURN_REFUND"
	case DecisionTypeExchange:
		return "EXCHANGE"
	case DecisionTypeRepair:
		return "REPAIR"
	case DecisionTypeCompensation:
		return "COMPENSATION"
	case DecisionTypeReject:
		return "REJECT"
	case DecisionTypeNoAction:
		return "NO_ACTION"
	default:
		return "UNKNOWN"
	}
}

type DecisionFactor int8

const (
	DecisionFactorEvidence      DecisionFactor = 1
	DecisionFactorHistory       DecisionFactor = 2
	DecisionFactorPolicy        DecisionFactor = 3
	DecisionFactorCommunication DecisionFactor = 4
	DecisionFactorProductStatus DecisionFactor = 5
)

func (f DecisionFactor) String() string {
	switch f {
	case DecisionFactorEvidence:
		return "EVIDENCE"
	case DecisionFactorHistory:
		return "HISTORY"
	case DecisionFactorPolicy:
		return "POLICY"
	case DecisionFactorCommunication:
		return "COMMUNICATION"
	case DecisionFactorProductStatus:
		return "PRODUCT_STATUS"
	default:
		return "UNKNOWN"
	}
}

type ArbitrationDecision struct {
	ID              uint64           `json:"id"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DecisionNo      string           `json:"decision_no"`
	ArbitrationID   uint64           `json:"arbitration_id"`
	ArbitrationNo   string           `json:"arbitration_no"`
	DecisionType    DecisionType     `json:"decision_type"`
	Result          ArbitrationResult `json:"result"`
	Reason          string           `json:"reason"`
	Amount          int64            `json:"amount"`
	RefundRatio     float64          `json:"refund_ratio"`
	HandlerID       uint64           `json:"handler_id"`
	HandlerName     string           `json:"handler_name"`
	DecidedAt       *time.Time       `json:"decided_at"`
	EffectiveAt     *time.Time       `json:"effective_at"`
	ExpiresAt       *time.Time       `json:"expires_at"`
	ExecutedAt      *time.Time       `json:"executed_at"`
	ExecutionStatus string           `json:"execution_status"`
	Factors         []*DecisionFactorItem `json:"factors"`
	Attachments     []string         `json:"attachments"`
	AppealDeadline  *time.Time       `json:"appeal_deadline"`
	IsAppealed      bool             `json:"is_appealed"`
	IsFinal         bool             `json:"is_final"`
}

type DecisionFactorItem struct {
	ID          uint64          `json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	DecisionID  uint64          `json:"decision_id"`
	Factor      DecisionFactor  `json:"factor"`
	Score       int             `json:"score"`
	Weight      float64         `json:"weight"`
	Description string          `json:"description"`
	EvidenceID  uint64          `json:"evidence_id"`
}

type EvidenceAssessment struct {
	ID              uint64           `json:"id"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	ArbitrationID   uint64           `json:"arbitration_id"`
	EvidenceID      uint64           `json:"evidence_id"`
	AssessorID      uint64           `json:"assessor_id"`
	AssessorName    string           `json:"assessor_name"`
	Authenticity    int              `json:"authenticity"`
	Relevance       int              `json:"relevance"`
	Credibility     int              `json:"credibility"`
	Weight          float64          `json:"weight"`
	Comment         string           `json:"comment"`
	AssessedAt      *time.Time       `json:"assessed_at"`
	FavorParty      ArbitrationParty `json:"favor_party"`
}

type ArbitrationPolicy struct {
	ID              uint64      `json:"id"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	CategoryID      uint64      `json:"category_id"`
	MinAmount       int64       `json:"min_amount"`
	MaxAmount       int64       `json:"max_amount"`
	DecisionRules   []*DecisionRule `json:"decision_rules"`
	Priority        int         `json:"priority"`
	Enabled         bool        `json:"enabled"`
}

type DecisionRule struct {
	ID          uint          `json:"id"`
	CreatedAt   time.Time     `json:"created_at"`
	PolicyID    uint          `json:"policy_id"`
	Condition   string        `json:"condition"`
	Decision    DecisionType  `json:"decision"`
	RefundRatio float64       `json:"refund_ratio"`
	Priority    int           `json:"priority"`
	Enabled     bool          `json:"enabled"`
}

func NewArbitrationDecision(decisionNo string, arbitrationID uint64, arbitrationNo string) *ArbitrationDecision {
	return &ArbitrationDecision{
		DecisionNo:    decisionNo,
		ArbitrationID: arbitrationID,
		ArbitrationNo: arbitrationNo,
		Factors:       make([]*DecisionFactorItem, 0),
		Attachments:   make([]string, 0),
		IsAppealed:    false,
		IsFinal:       false,
	}
}

func (d *ArbitrationDecision) SetDecision(decisionType DecisionType, result ArbitrationResult, reason string, amount int64, refundRatio float64, handlerID uint64, handlerName string) error {
	if d.DecidedAt != nil {
		return ErrDecisionAlreadyMade
	}

	now := time.Now()
	d.DecisionType = decisionType
	d.Result = result
	d.Reason = reason
	d.Amount = amount
	d.RefundRatio = refundRatio
	d.HandlerID = handlerID
	d.HandlerName = handlerName
	d.DecidedAt = &now

	effectiveAt := now.Add(time.Hour * 24)
	d.EffectiveAt = &effectiveAt

	return nil
}

func (d *ArbitrationDecision) AddFactor(factor DecisionFactor, score int, weight float64, description string, evidenceID uint64) {
	item := &DecisionFactorItem{
		DecisionID:  d.ID,
		Factor:      factor,
		Score:       score,
		Weight:      weight,
		Description: description,
		EvidenceID:  evidenceID,
	}
	d.Factors = append(d.Factors, item)
}

func (d *ArbitrationDecision) SetAppealDeadline(duration time.Duration) {
	deadline := time.Now().Add(duration)
	d.AppealDeadline = &deadline
}

func (d *ArbitrationDecision) CanAppeal() bool {
	if d.IsFinal {
		return false
	}
	if d.AppealDeadline == nil {
		return false
	}
	return time.Now().Before(*d.AppealDeadline) && !d.IsAppealed
}

func (d *ArbitrationDecision) MarkAppealed() {
	d.IsAppealed = true
}

func (d *ArbitrationDecision) MarkFinal() {
	d.IsFinal = true
}

func (d *ArbitrationDecision) MarkExecuted() {
	now := time.Now()
	d.ExecutedAt = &now
	d.ExecutionStatus = "EXECUTED"
}

func (d *ArbitrationDecision) IsEffective() bool {
	if d.EffectiveAt == nil {
		return false
	}
	return time.Now().After(*d.EffectiveAt)
}

func (d *ArbitrationDecision) IsExpired() bool {
	if d.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*d.ExpiresAt)
}

func (d *ArbitrationDecision) CalculateWeightedScore() float64 {
	var totalScore float64
	var totalWeight float64

	for _, factor := range d.Factors {
		totalScore += float64(factor.Score) * factor.Weight
		totalWeight += factor.Weight
	}

	if totalWeight == 0 {
		return 0
	}

	return totalScore / totalWeight
}

func NewEvidenceAssessment(arbitrationID, evidenceID, assessorID uint64, assessorName string) *EvidenceAssessment {
	return &EvidenceAssessment{
		ArbitrationID: arbitrationID,
		EvidenceID:    evidenceID,
		AssessorID:    assessorID,
		AssessorName:  assessorName,
		Weight:        1.0,
	}
}

func (a *EvidenceAssessment) SetScores(authenticity, relevance, credibility int) {
	a.Authenticity = authenticity
	a.Relevance = relevance
	a.Credibility = credibility
}

func (a *EvidenceAssessment) CalculateOverallScore() float64 {
	return float64(a.Authenticity+a.Relevance+a.Credibility) / 3.0
}

func (a *EvidenceAssessment) SetFavorParty(party ArbitrationParty) {
	a.FavorParty = party
}

func (a *EvidenceAssessment) Complete() {
	now := time.Now()
	a.AssessedAt = &now
}

func (a *EvidenceAssessment) IsFavorUser() bool {
	return a.FavorParty == ArbitrationPartyUser
}

func (a *EvidenceAssessment) IsFavorMerchant() bool {
	return a.FavorParty == ArbitrationPartyMerchant
}

func (a *Arbitration) CreateDecision(decisionNo string) *ArbitrationDecision {
	decision := NewArbitrationDecision(decisionNo, a.ID, a.ArbitrationNo)
	return decision
}

func (a *Arbitration) AssessEvidence(evidenceID, assessorID uint64, assessorName string, authenticity, relevance, credibility int, comment string, favorParty ArbitrationParty) (*EvidenceAssessment, error) {
	evidence := a.findEvidence(evidenceID)
	if evidence == nil {
		return nil, ErrArbitrationNotFound
	}

	assessment := NewEvidenceAssessment(a.ID, evidenceID, assessorID, assessorName)
	assessment.SetScores(authenticity, relevance, credibility)
	assessment.Comment = comment
	assessment.SetFavorParty(favorParty)
	assessment.Complete()

	evidence.Verified = true
	now := time.Now()
	evidence.VerifiedAt = &now

	return assessment, nil
}

func (a *Arbitration) findEvidence(evidenceID uint64) *ArbitrationEvidence {
	for _, e := range a.Evidence {
		if e.ID == evidenceID {
			return e
		}
	}
	return nil
}

func (a *Arbitration) CalculateEvidenceScore() (userScore, merchantScore float64) {
	for _, evidence := range a.Evidence {
		if !evidence.Verified {
			continue
		}
		switch evidence.Party {
		case ArbitrationPartyUser:
			userScore++
		case ArbitrationPartyMerchant:
			merchantScore++
		}
	}
	return
}

func (a *Arbitration) SuggestDecision() DecisionType {
	userScore, merchantScore := a.CalculateEvidenceScore()

	if userScore > merchantScore*1.5 {
		return DecisionTypeFullRefund
	} else if merchantScore > userScore*1.5 {
		return DecisionTypeReject
	} else if userScore > 0 || merchantScore > 0 {
		return DecisionTypePartialRefund
	}

	return DecisionTypeNoAction
}

type ArbitrationDecisionRepository interface {
	Save(ctx context.Context, decision *ArbitrationDecision) error
	FindByID(ctx context.Context, id uint64) (*ArbitrationDecision, error)
	FindByDecisionNo(ctx context.Context, decisionNo string) (*ArbitrationDecision, error)
	FindByArbitrationID(ctx context.Context, arbitrationID uint64) (*ArbitrationDecision, error)
	FindPendingExecution(ctx context.Context, limit int) ([]*ArbitrationDecision, error)
	Update(ctx context.Context, decision *ArbitrationDecision) error
}

type EvidenceAssessmentRepository interface {
	Save(ctx context.Context, assessment *EvidenceAssessment) error
	FindByID(ctx context.Context, id uint64) (*EvidenceAssessment, error)
	FindByEvidenceID(ctx context.Context, evidenceID uint64) (*EvidenceAssessment, error)
	FindByArbitrationID(ctx context.Context, arbitrationID uint64) ([]*EvidenceAssessment, error)
	Update(ctx context.Context, assessment *EvidenceAssessment) error
}

type ArbitrationPolicyRepository interface {
	FindByID(ctx context.Context, id uint64) (*ArbitrationPolicy, error)
	FindByCategoryID(ctx context.Context, categoryID uint64) ([]*ArbitrationPolicy, error)
	FindEnabled(ctx context.Context) ([]*ArbitrationPolicy, error)
	FindMatchingPolicy(ctx context.Context, categoryID uint64, amount int64) (*ArbitrationPolicy, error)
	Save(ctx context.Context, policy *ArbitrationPolicy) error
}
