package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrFinanceApplicationNotFound = errors.New("finance application not found")
	ErrCreditLineNotFound         = errors.New("credit line not found")
	ErrInvoiceFinancingNotFound   = errors.New("invoice financing not found")
	ErrInsufficientCreditLimit    = errors.New("insufficient credit limit")
	ErrCollateralInsufficient     = errors.New("collateral insufficient")
	ErrRepaymentFailed            = errors.New("repayment failed")
	ErrFactoringNotFound          = errors.New("factoring not found")
	ErrSupplierNotFound           = errors.New("supplier not found")
	ErrBuyerNotFound              = errors.New("buyer not found")
	ErrRiskAssessmentFailed       = errors.New("risk assessment failed")
	ErrGuaranteeNotFound          = errors.New("guarantee not found")
	ErrInvalidFinanceAmount       = errors.New("invalid finance amount")
	ErrFinanceAlreadyApproved     = errors.New("finance already approved")
	ErrFinanceAlreadyRejected     = errors.New("finance already rejected")
)

type FinanceType string

const (
	FinanceTypeInvoice        FinanceType = "INVOICE"
	FinanceTypePurchaseOrder  FinanceType = "PURCHASE_ORDER"
	FinanceTypeInventory      FinanceType = "INVENTORY"
	FinanceTypeSupplier       FinanceType = "SUPPLIER"
	FinanceTypeDistributor    FinanceType = "DISTRIBUTOR"
	FinanceTypeFactoring      FinanceType = "FACTORING"
	FinanceTypeReverseFactoring FinanceType = "REVERSE_FACTORING"
	FinanceTypeForfaiting     FinanceType = "FORFAITING"
	FinanceTypeWorkingCapital FinanceType = "WORKING_CAPITAL"
)

type FinanceStatus string

const (
	FinanceStatusDraft      FinanceStatus = "DRAFT"
	FinanceStatusPending    FinanceStatus = "PENDING"
	FinanceStatusUnderReview FinanceStatus = "UNDER_REVIEW"
	FinanceStatusApproved   FinanceStatus = "APPROVED"
	FinanceStatusRejected   FinanceStatus = "REJECTED"
	FinanceStatusDisbursed  FinanceStatus = "DISBURSED"
	FinanceStatusActive     FinanceStatus = "ACTIVE"
	FinanceStatusRepaying   FinanceStatus = "REPAYING"
	FinanceStatusCompleted  FinanceStatus = "COMPLETED"
	FinanceStatusDefaulted  FinanceStatus = "DEFAULTED"
	FinanceStatusCancelled  FinanceStatus = "CANCELLED"
	FinanceStatusOverdue    FinanceStatus = "OVERDUE"
)

type CreditLineStatus string

const (
	CreditLineStatusActive     CreditLineStatus = "ACTIVE"
	CreditLineStatusFrozen     CreditLineStatus = "FROZEN"
	CreditLineStatusClosed     CreditLineStatus = "CLOSED"
	CreditLineStatusSuspended  CreditLineStatus = "SUSPENDED"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

type RepaymentFrequency string

const (
	RepaymentFrequencyOneTime     RepaymentFrequency = "ONE_TIME"
	RepaymentFrequencyWeekly      RepaymentFrequency = "WEEKLY"
	RepaymentFrequencyBiWeekly    RepaymentFrequency = "BI_WEEKLY"
	RepaymentFrequencyMonthly     RepaymentFrequency = "MONTHLY"
	RepaymentFrequencyQuarterly   RepaymentFrequency = "QUARTERLY"
	RepaymentFrequencySemiAnnually RepaymentFrequency = "SEMI_ANNUALLY"
)

type CollateralType string

const (
	CollateralTypeInventory   CollateralType = "INVENTORY"
	CollateralTypeReceivables CollateralType = "RECEIVABLES"
	CollateralTypeEquipment   CollateralType = "EQUIPMENT"
	CollateralTypeRealEstate  CollateralType = "REAL_ESTATE"
	CollateralTypeSecurities  CollateralType = "SECURITIES"
	CollateralTypeCash        CollateralType = "CASH"
	CollateralTypeGuarantee   CollateralType = "GUARANTEE"
)

type GuaranteeType string

const (
	GuaranteeTypeBank       GuaranteeType = "BANK"
	GuaranteeTypeCorporate  GuaranteeType = "CORPORATE"
	GuaranteeTypePersonal   GuaranteeType = "PERSONAL"
	GuaranteeTypeGovernment GuaranteeType = "GOVERNMENT"
	GuaranteeTypeInsurance  GuaranteeType = "INSURANCE"
)

type FinanceApplication struct {
	ID              string          `json:"id"`
	ApplicantID     string          `json:"applicant_id"`
	ApplicantName   string          `json:"applicant_name"`
	ApplicantType   string          `json:"applicant_type"`
	FinanceType     FinanceType     `json:"finance_type"`
	RequestedAmount decimal.Decimal `json:"requested_amount"`
	ApprovedAmount  decimal.Decimal `json:"approved_amount"`
	Currency        string          `json:"currency"`
	Purpose         string          `json:"purpose"`
	TermDays        int             `json:"term_days"`
	Status          FinanceStatus   `json:"status"`
	CreditLineID    string          `json:"credit_line_id"`
	CollateralIDs   []string        `json:"collateral_ids"`
	GuaranteeIDs    []string        `json:"guarantee_ids"`
	RiskLevel       RiskLevel       `json:"risk_level"`
	RiskScore       decimal.Decimal `json:"risk_score"`
	InterestRate    decimal.Decimal `json:"interest_rate"`
	FeeAmount       decimal.Decimal `json:"fee_amount"`
	Documents       map[string]string `json:"documents"`
	ApprovalHistory []string        `json:"approval_history"`
	SubmittedAt     time.Time       `json:"submitted_at"`
	ApprovedAt      *time.Time      `json:"approved_at"`
	RejectedAt      *time.Time      `json:"rejected_at"`
	RejectionReason string          `json:"rejection_reason"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func NewFinanceApplication(applicantID, applicantName string, financeType FinanceType, amount decimal.Decimal, currency string, termDays int, purpose string) *FinanceApplication {
	return &FinanceApplication{
		ApplicantID:     applicantID,
		ApplicantName:   applicantName,
		FinanceType:     financeType,
		RequestedAmount: amount,
		Currency:        currency,
		TermDays:        termDays,
		Purpose:         purpose,
		Status:          FinanceStatusDraft,
		CollateralIDs:   []string{},
		GuaranteeIDs:    []string{},
		Documents:       make(map[string]string),
		ApprovalHistory: []string{},
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (fa *FinanceApplication) Submit() {
	fa.Status = FinanceStatusPending
	fa.SubmittedAt = time.Now()
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceApplication) Approve(amount, interestRate, feeAmount decimal.Decimal, approvedBy string) {
	now := time.Now()
	fa.Status = FinanceStatusApproved
	fa.ApprovedAmount = amount
	fa.InterestRate = interestRate
	fa.FeeAmount = feeAmount
	fa.ApprovedAt = &now
	fa.ApprovalHistory = append(fa.ApprovalHistory, "Approved by "+approvedBy+" at "+now.Format(time.RFC3339))
	fa.UpdatedAt = now
}

func (fa *FinanceApplication) Reject(reason, rejectedBy string) {
	now := time.Now()
	fa.Status = FinanceStatusRejected
	fa.RejectionReason = reason
	fa.RejectedAt = &now
	fa.ApprovalHistory = append(fa.ApprovalHistory, "Rejected by "+rejectedBy+" at "+now.Format(time.RFC3339)+": "+reason)
	fa.UpdatedAt = now
}

func (fa *FinanceApplication) Cancel(reason string) {
	fa.Status = FinanceStatusCancelled
	fa.RejectionReason = reason
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceApplication) AddCollateral(collateralID string) {
	for _, c := range fa.CollateralIDs {
		if c == collateralID {
			return
		}
	}
	fa.CollateralIDs = append(fa.CollateralIDs, collateralID)
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceApplication) AddGuarantee(guaranteeID string) {
	for _, g := range fa.GuaranteeIDs {
		if g == guaranteeID {
			return
		}
	}
	fa.GuaranteeIDs = append(fa.GuaranteeIDs, guaranteeID)
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceApplication) AddDocument(docType, docID string) {
	fa.Documents[docType] = docID
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceApplication) SetRiskAssessment(level RiskLevel, score decimal.Decimal) {
	fa.RiskLevel = level
	fa.RiskScore = score
	fa.UpdatedAt = time.Now()
}

type CreditLine struct {
	ID              string           `json:"id"`
	OwnerID         string           `json:"owner_id"`
	OwnerName       string           `json:"owner_name"`
	OwnerType       string           `json:"owner_type"`
	TotalLimit      decimal.Decimal  `json:"total_limit"`
	UsedAmount      decimal.Decimal  `json:"used_amount"`
	AvailableAmount decimal.Decimal  `json:"available_amount"`
	Currency        string           `json:"currency"`
	Status          CreditLineStatus `json:"status"`
	InterestRate    decimal.Decimal  `json:"interest_rate"`
	AnnualFee       decimal.Decimal  `json:"annual_fee"`
	EffectiveFrom   time.Time        `json:"effective_from"`
	EffectiveTo     time.Time        `json:"effective_to"`
	ReviewFrequency string           `json:"review_frequency"`
	LastReview      *time.Time       `json:"last_review"`
	NextReview      *time.Time       `json:"next_review"`
	FinancingIDs    []string         `json:"financing_ids"`
	Terms           map[string]string `json:"terms"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func NewCreditLine(ownerID, ownerName, ownerType string, limit decimal.Decimal, currency string) *CreditLine {
	return &CreditLine{
		OwnerID:         ownerID,
		OwnerName:       ownerName,
		OwnerType:       ownerType,
		TotalLimit:      limit,
		UsedAmount:      decimal.Zero,
		AvailableAmount: limit,
		Currency:        currency,
		Status:          CreditLineStatusActive,
		InterestRate:    decimal.NewFromFloat(0.08),
		EffectiveFrom:   time.Now(),
		EffectiveTo:     time.Now().AddDate(1, 0, 0),
		ReviewFrequency: "QUARTERLY",
		FinancingIDs:    []string{},
		Terms:           make(map[string]string),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

func (cl *CreditLine) UseCredit(amount decimal.Decimal) error {
	if cl.AvailableAmount.LessThan(amount) {
		return ErrInsufficientCreditLimit
	}
	cl.UsedAmount = cl.UsedAmount.Add(amount)
	cl.AvailableAmount = cl.AvailableAmount.Sub(amount)
	cl.UpdatedAt = time.Now()
	return nil
}

func (cl *CreditLine) ReleaseCredit(amount decimal.Decimal) {
	cl.UsedAmount = cl.UsedAmount.Sub(amount)
	cl.AvailableAmount = cl.AvailableAmount.Add(amount)
	if cl.UsedAmount.LessThan(decimal.Zero) {
		cl.UsedAmount = decimal.Zero
	}
	if cl.AvailableAmount.GreaterThan(cl.TotalLimit) {
		cl.AvailableAmount = cl.TotalLimit
	}
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) IncreaseLimit(amount decimal.Decimal) {
	cl.TotalLimit = cl.TotalLimit.Add(amount)
	cl.AvailableAmount = cl.AvailableAmount.Add(amount)
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) DecreaseLimit(amount decimal.Decimal) error {
	if cl.TotalLimit.Sub(amount).LessThan(cl.UsedAmount) {
		return ErrInsufficientCreditLimit
	}
	cl.TotalLimit = cl.TotalLimit.Sub(amount)
	cl.AvailableAmount = cl.TotalLimit.Sub(cl.UsedAmount)
	cl.UpdatedAt = time.Now()
	return nil
}

func (cl *CreditLine) Freeze() {
	cl.Status = CreditLineStatusFrozen
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) Unfreeze() {
	cl.Status = CreditLineStatusActive
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) Suspend() {
	cl.Status = CreditLineStatusSuspended
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) Close() {
	cl.Status = CreditLineStatusClosed
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) AddFinancing(financingID string) {
	cl.FinancingIDs = append(cl.FinancingIDs, financingID)
	cl.UpdatedAt = time.Now()
}

func (cl *CreditLine) MarkReviewed() {
	now := time.Now()
	cl.LastReview = &now
	cl.UpdatedAt = now
}

func (cl *CreditLine) UtilizationRate() decimal.Decimal {
	if cl.TotalLimit.IsZero() {
		return decimal.Zero
	}
	return cl.UsedAmount.Div(cl.TotalLimit).Mul(decimal.NewFromInt(100))
}

type InvoiceFinancing struct {
	ID               string          `json:"id"`
	ApplicationID    string          `json:"application_id"`
	BorrowerID       string          `json:"borrower_id"`
	BorrowerName     string          `json:"borrower_name"`
	InvoiceID        string          `json:"invoice_id"`
	InvoiceNumber    string          `json:"invoice_number"`
	InvoiceAmount    decimal.Decimal `json:"invoice_amount"`
	FinancingAmount  decimal.Decimal `json:"financing_amount"`
	AdvanceRate      decimal.Decimal `json:"advance_rate"`
	Currency         string          `json:"currency"`
	InterestRate     decimal.Decimal `json:"interest_rate"`
	InterestAmount   decimal.Decimal `json:"interest_amount"`
	FeeAmount        decimal.Decimal `json:"fee_amount"`
	DisbursementAmount decimal.Decimal `json:"disbursement_amount"`
	OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
	RepaidAmount     decimal.Decimal `json:"repaid_amount"`
	Status           FinanceStatus   `json:"status"`
	InvoiceDate      time.Time       `json:"invoice_date"`
	InvoiceDueDate   time.Time       `json:"invoice_due_date"`
	DisbursementDate *time.Time      `json:"disbursement_date"`
	MaturityDate     time.Time       `json:"maturity_date"`
	BuyerID          string          `json:"buyer_id"`
	BuyerName        string          `json:"buyer_name"`
	CreditLineID     string          `json:"credit_line_id"`
	CollateralID     string          `json:"collateral_id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func NewInvoiceFinancing(applicationID, borrowerID, invoiceID string, invoiceAmount, advanceRate decimal.Decimal, currency string) *InvoiceFinancing {
	financingAmount := invoiceAmount.Mul(advanceRate)
	return &InvoiceFinancing{
		ApplicationID:    applicationID,
		BorrowerID:       borrowerID,
		InvoiceID:        invoiceID,
		InvoiceAmount:    invoiceAmount,
		FinancingAmount:  financingAmount,
		AdvanceRate:      advanceRate,
		Currency:         currency,
		InterestRate:     decimal.NewFromFloat(0.06),
		InterestAmount:   decimal.Zero,
		FeeAmount:        decimal.Zero,
		DisbursementAmount: financingAmount,
		OutstandingAmount: financingAmount,
		RepaidAmount:     decimal.Zero,
		Status:           FinanceStatusPending,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (ifc *InvoiceFinancing) Approve(interestRate, feeAmount decimal.Decimal) {
	now := time.Now()
	ifc.Status = FinanceStatusApproved
	ifc.InterestRate = interestRate
	ifc.FeeAmount = feeAmount
	ifc.DisbursementAmount = ifc.FinancingAmount.Sub(feeAmount)
	ifc.UpdatedAt = now
}

func (ifc *InvoiceFinancing) Disburse() {
	now := time.Now()
	ifc.Status = FinanceStatusDisbursed
	ifc.DisbursementDate = &now
	ifc.UpdatedAt = now
}

func (ifc *InvoiceFinancing) AccrueInterest(days int) {
	dailyRate := ifc.InterestRate.Div(decimal.NewFromInt(365))
	interest := ifc.OutstandingAmount.Mul(dailyRate).Mul(decimal.NewFromInt(int64(days)))
	ifc.InterestAmount = ifc.InterestAmount.Add(interest)
	ifc.UpdatedAt = time.Now()
}

func (ifc *InvoiceFinancing) Repay(amount decimal.Decimal) {
	if ifc.OutstandingAmount.LessThanOrEqual(amount) {
		ifc.OutstandingAmount = decimal.Zero
		ifc.RepaidAmount = ifc.FinancingAmount
		ifc.Status = FinanceStatusCompleted
	} else {
		ifc.OutstandingAmount = ifc.OutstandingAmount.Sub(amount)
		ifc.RepaidAmount = ifc.RepaidAmount.Add(amount)
		ifc.Status = FinanceStatusRepaying
	}
	ifc.UpdatedAt = time.Now()
}

func (ifc *InvoiceFinancing) MarkOverdue() {
	ifc.Status = FinanceStatusOverdue
	ifc.UpdatedAt = time.Now()
}

func (ifc *InvoiceFinancing) MarkDefaulted() {
	ifc.Status = FinanceStatusDefaulted
	ifc.UpdatedAt = time.Now()
}

type SupplierRiskProfile struct {
	ID                    string          `json:"id"`
	SupplierID            string          `json:"supplier_id"`
	SupplierName          string          `json:"supplier_name"`
	OverallRiskLevel      RiskLevel       `json:"overall_risk_level"`
	OverallRiskScore      decimal.Decimal `json:"overall_risk_score"`
	CreditRating          string          `json:"credit_rating"`
	FinancialRiskScore    decimal.Decimal `json:"financial_risk_score"`
	OperationalRiskScore  decimal.Decimal `json:"operational_risk_score"`
	MarketRiskScore       decimal.Decimal `json:"market_risk_score"`
	ComplianceRiskScore   decimal.Decimal `json:"compliance_risk_score"`
	PaymentHistoryScore   decimal.Decimal `json:"payment_history_score"`
	DeliveryPerfScore     decimal.Decimal `json:"delivery_perf_score"`
	QualityScore          decimal.Decimal `json:"quality_score"`
	TotalTransactions     int64           `json:"total_transactions"`
	TotalTransactionValue decimal.Decimal `json:"total_transaction_value"`
	OnTimePayments        int64           `json:"on_time_payments"`
	LatePayments          int64           `json:"late_payments"`
	Defaults              int64           `json:"defaults"`
	DefaultRate           decimal.Decimal `json:"default_rate"`
	Alerts                []string        `json:"alerts"`
	LastAssessment        *time.Time      `json:"last_assessment"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

func NewSupplierRiskProfile(supplierID, supplierName string) *SupplierRiskProfile {
	return &SupplierRiskProfile{
		SupplierID:       supplierID,
		SupplierName:     supplierName,
		OverallRiskLevel: RiskLevelMedium,
		OverallRiskScore: decimal.NewFromInt(50),
		Alerts:           []string{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (srp *SupplierRiskProfile) CalculateOverallScore() {
	srp.OverallRiskScore = srp.FinancialRiskScore.Mul(decimal.NewFromFloat(0.25)).
		Add(srp.OperationalRiskScore.Mul(decimal.NewFromFloat(0.2))).
		Add(srp.MarketRiskScore.Mul(decimal.NewFromFloat(0.15))).
		Add(srp.ComplianceRiskScore.Mul(decimal.NewFromFloat(0.15))).
		Add(srp.PaymentHistoryScore.Mul(decimal.NewFromFloat(0.15))).
		Add(srp.DeliveryPerfScore.Mul(decimal.NewFromFloat(0.05))).
		Add(srp.QualityScore.Mul(decimal.NewFromFloat(0.05)))
	
	srp.determineRiskLevel()
	srp.UpdatedAt = time.Now()
}

func (srp *SupplierRiskProfile) determineRiskLevel() {
	score := srp.OverallRiskScore
	switch {
	case score.GreaterThanOrEqual(decimal.NewFromInt(80)):
		srp.OverallRiskLevel = RiskLevelLow
	case score.GreaterThanOrEqual(decimal.NewFromInt(60)):
		srp.OverallRiskLevel = RiskLevelMedium
	case score.GreaterThanOrEqual(decimal.NewFromInt(40)):
		srp.OverallRiskLevel = RiskLevelHigh
	default:
		srp.OverallRiskLevel = RiskLevelCritical
	}
}

func (srp *SupplierRiskProfile) RecordPayment(onTime bool) {
	srp.TotalTransactions++
	if onTime {
		srp.OnTimePayments++
	} else {
		srp.LatePayments++
	}
	srp.calculateDefaultRate()
	srp.UpdatedAt = time.Now()
}

func (srp *SupplierRiskProfile) RecordDefault() {
	srp.Defaults++
	srp.calculateDefaultRate()
	srp.UpdatedAt = time.Now()
}

func (srp *SupplierRiskProfile) calculateDefaultRate() {
	if srp.TotalTransactions > 0 {
		srp.DefaultRate = decimal.NewFromInt(srp.Defaults).Div(decimal.NewFromInt(srp.TotalTransactions)).Mul(decimal.NewFromInt(100))
	}
}

func (srp *SupplierRiskProfile) AddAlert(alert string) {
	srp.Alerts = append(srp.Alerts, alert)
	srp.UpdatedAt = time.Now()
}

func (srp *SupplierRiskProfile) MarkAssessed() {
	now := time.Now()
	srp.LastAssessment = &now
	srp.UpdatedAt = now
}

type BuyerRiskProfile struct {
	ID                    string          `json:"id"`
	BuyerID               string          `json:"buyer_id"`
	BuyerName             string          `json:"buyer_name"`
	OverallRiskLevel      RiskLevel       `json:"overall_risk_level"`
	OverallRiskScore      decimal.Decimal `json:"overall_risk_score"`
	CreditRating          string          `json:"credit_rating"`
	PaymentBehaviorScore  decimal.Decimal `json:"payment_behavior_score"`
	FinancialStabilityScore decimal.Decimal `json:"financial_stability_score"`
	TotalOrders           int64           `json:"total_orders"`
	TotalOrderValue       decimal.Decimal `json:"total_order_value"`
	OnTimePayments        int64           `json:"on_time_payments"`
	LatePayments          int64           `json:"late_payments"`
	Defaults              int64           `json:"defaults"`
	DefaultRate           decimal.Decimal `json:"default_rate"`
	Alerts                []string        `json:"alerts"`
	LastAssessment        *time.Time      `json:"last_assessment"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

func NewBuyerRiskProfile(buyerID, buyerName string) *BuyerRiskProfile {
	return &BuyerRiskProfile{
		BuyerID:          buyerID,
		BuyerName:        buyerName,
		OverallRiskLevel: RiskLevelMedium,
		OverallRiskScore: decimal.NewFromInt(50),
		Alerts:           []string{},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (brp *BuyerRiskProfile) CalculateOverallScore() {
	brp.OverallRiskScore = brp.PaymentBehaviorScore.Mul(decimal.NewFromFloat(0.6)).
		Add(brp.FinancialStabilityScore.Mul(decimal.NewFromFloat(0.4)))
	
	brp.determineRiskLevel()
	brp.UpdatedAt = time.Now()
}

func (brp *BuyerRiskProfile) determineRiskLevel() {
	score := brp.OverallRiskScore
	switch {
	case score.GreaterThanOrEqual(decimal.NewFromInt(80)):
		brp.OverallRiskLevel = RiskLevelLow
	case score.GreaterThanOrEqual(decimal.NewFromInt(60)):
		brp.OverallRiskLevel = RiskLevelMedium
	case score.GreaterThanOrEqual(decimal.NewFromInt(40)):
		brp.OverallRiskLevel = RiskLevelHigh
	default:
		brp.OverallRiskLevel = RiskLevelCritical
	}
}

func (brp *BuyerRiskProfile) RecordPayment(onTime bool, orderValue decimal.Decimal) {
	brp.TotalOrders++
	brp.TotalOrderValue = brp.TotalOrderValue.Add(orderValue)
	if onTime {
		brp.OnTimePayments++
	} else {
		brp.LatePayments++
	}
	brp.calculateDefaultRate()
	brp.UpdatedAt = time.Now()
}

func (brp *BuyerRiskProfile) RecordDefault() {
	brp.Defaults++
	brp.calculateDefaultRate()
	brp.UpdatedAt = time.Now()
}

func (brp *BuyerRiskProfile) calculateDefaultRate() {
	if brp.TotalOrders > 0 {
		brp.DefaultRate = decimal.NewFromInt(brp.Defaults).Div(decimal.NewFromInt(brp.TotalOrders)).Mul(decimal.NewFromInt(100))
	}
}

func (brp *BuyerRiskProfile) AddAlert(alert string) {
	brp.Alerts = append(brp.Alerts, alert)
	brp.UpdatedAt = time.Now()
}

func (brp *BuyerRiskProfile) MarkAssessed() {
	now := time.Now()
	brp.LastAssessment = &now
	brp.UpdatedAt = now
}

type Collateral struct {
	ID              string          `json:"id"`
	OwnerID         string          `json:"owner_id"`
	OwnerName       string          `json:"owner_name"`
	CollateralType  CollateralType  `json:"collateral_type"`
	Description     string          `json:"description"`
	OriginalValue   decimal.Decimal `json:"original_value"`
	CurrentValue    decimal.Decimal `json:"current_value"`
	Haircut         decimal.Decimal `json:"haircut"`
	EligibleValue   decimal.Decimal `json:"eligible_value"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	Location        string          `json:"location"`
	Custodian       string          `json:"custodian"`
	FinancingIDs    []string        `json:"financing_ids"`
	LastValuation   *time.Time      `json:"last_valuation"`
	NextValuation   *time.Time      `json:"next_valuation"`
	Details         map[string]string `json:"details"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func NewCollateral(ownerID, ownerName string, collateralType CollateralType, value decimal.Decimal, currency string) *Collateral {
	return &Collateral{
		OwnerID:        ownerID,
		OwnerName:      ownerName,
		CollateralType: collateralType,
		OriginalValue:  value,
		CurrentValue:   value,
		Haircut:        decimal.NewFromFloat(0.2),
		EligibleValue:  value.Mul(decimal.NewFromFloat(0.8)),
		Currency:       currency,
		Status:         "ACTIVE",
		FinancingIDs:   []string{},
		Details:        make(map[string]string),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (c *Collateral) UpdateValue(newValue decimal.Decimal) {
	c.CurrentValue = newValue
	c.EligibleValue = newValue.Mul(decimal.NewFromInt(1).Sub(c.Haircut))
	now := time.Now()
	c.LastValuation = &now
	c.UpdatedAt = now
}

func (c *Collateral) SetHaircut(haircut decimal.Decimal) {
	c.Haircut = haircut
	c.EligibleValue = c.CurrentValue.Mul(decimal.NewFromInt(1).Sub(haircut))
	c.UpdatedAt = time.Now()
}

func (c *Collateral) Lock() {
	c.Status = "LOCKED"
	c.UpdatedAt = time.Now()
}

func (c *Collateral) Release() {
	c.Status = "RELEASED"
	c.UpdatedAt = time.Now()
}

func (c *Collateral) AddFinancing(financingID string) {
	c.FinancingIDs = append(c.FinancingIDs, financingID)
	c.UpdatedAt = time.Now()
}

type Guarantee struct {
	ID             string        `json:"id"`
	GuarantorID    string        `json:"guarantor_id"`
	GuarantorName  string        `json:"guarantor_name"`
	BeneficiaryID  string        `json:"beneficiary_id"`
	BeneficiaryName string       `json:"beneficiary_name"`
	GuaranteeType  GuaranteeType `json:"guarantee_type"`
	GuaranteeAmount decimal.Decimal `json:"guarantee_amount"`
	Currency       string        `json:"currency"`
	Status         string        `json:"status"`
	EffectiveFrom  time.Time     `json:"effective_from"`
	EffectiveTo    time.Time     `json:"effective_to"`
	FinancingID    string        `json:"financing_id"`
	ClaimHistory   []string      `json:"claim_history"`
	Terms          string        `json:"terms"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func NewGuarantee(guarantorID, guarantorName, beneficiaryID, beneficiaryName string, guaranteeType GuaranteeType, amount decimal.Decimal, currency string) *Guarantee {
	return &Guarantee{
		GuarantorID:    guarantorID,
		GuarantorName:  guarantorName,
		BeneficiaryID:  beneficiaryID,
		BeneficiaryName: beneficiaryName,
		GuaranteeType:  guaranteeType,
		GuaranteeAmount: amount,
		Currency:       currency,
		Status:         "PENDING",
		EffectiveFrom:  time.Now(),
		EffectiveTo:    time.Now().AddDate(1, 0, 0),
		ClaimHistory:   []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (g *Guarantee) Activate() {
	g.Status = "ACTIVE"
	g.UpdatedAt = time.Now()
}

func (g *Guarantee) Claim(amount decimal.Decimal, reason string) {
	claim := time.Now().Format(time.RFC3339) + ": Claimed " + amount.String() + " - " + reason
	g.ClaimHistory = append(g.ClaimHistory, claim)
	g.UpdatedAt = time.Now()
}

func (g *Guarantee) Release() {
	g.Status = "RELEASED"
	g.UpdatedAt = time.Now()
}

func (g *Guarantee) IsExpired() bool {
	return time.Now().After(g.EffectiveTo)
}

type RepaymentPlan struct {
	ID                 string             `json:"id"`
	FinancingID        string             `json:"financing_id"`
	BorrowerID         string             `json:"borrower_id"`
	TotalAmount        decimal.Decimal    `json:"total_amount"`
	PaidAmount         decimal.Decimal    `json:"paid_amount"`
	RemainingAmount    decimal.Decimal    `json:"remaining_amount"`
	RepaymentFrequency RepaymentFrequency `json:"repayment_frequency"`
	TotalInstallments  int                `json:"total_installments"`
	PaidInstallments   int                `json:"paid_installments"`
	RemainingInstallments int             `json:"remaining_installments"`
	NextPaymentAmount  decimal.Decimal    `json:"next_payment_amount"`
	NextPaymentDate    time.Time          `json:"next_payment_date"`
	Status             string             `json:"status"`
	Installments       []RepaymentInstallment `json:"installments"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type RepaymentInstallment struct {
	ID                string          `json:"id"`
	PlanID            string          `json:"plan_id"`
	InstallmentNumber int             `json:"installment_number"`
	DueDate           time.Time       `json:"due_date"`
	PrincipalAmount   decimal.Decimal `json:"principal_amount"`
	InterestAmount    decimal.Decimal `json:"interest_amount"`
	FeeAmount         decimal.Decimal `json:"fee_amount"`
	TotalAmount       decimal.Decimal `json:"total_amount"`
	RemainingPrincipal decimal.Decimal `json:"remaining_principal"`
	Status            string          `json:"status"`
	PaidDate          *time.Time      `json:"paid_date"`
	PaidAmount        decimal.Decimal `json:"paid_amount"`
}

func NewRepaymentPlan(financingID, borrowerID string, totalAmount decimal.Decimal, frequency RepaymentFrequency, installments int) *RepaymentPlan {
	return &RepaymentPlan{
		FinancingID:        financingID,
		BorrowerID:         borrowerID,
		TotalAmount:        totalAmount,
		PaidAmount:         decimal.Zero,
		RemainingAmount:    totalAmount,
		RepaymentFrequency: frequency,
		TotalInstallments:  installments,
		PaidInstallments:   0,
		RemainingInstallments: installments,
		Status:             "ACTIVE",
		Installments:       []RepaymentInstallment{},
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func (rp *RepaymentPlan) AddInstallment(installment RepaymentInstallment) {
	rp.Installments = append(rp.Installments, installment)
	rp.UpdatedAt = time.Now()
}

func (rp *RepaymentPlan) RecordPayment(installmentID string, amount decimal.Decimal) {
	for i, inst := range rp.Installments {
		if inst.ID == installmentID {
			now := time.Now()
			rp.Installments[i].Status = "PAID"
			rp.Installments[i].PaidDate = &now
			rp.Installments[i].PaidAmount = amount
			rp.PaidAmount = rp.PaidAmount.Add(amount)
			rp.RemainingAmount = rp.RemainingAmount.Sub(amount)
			rp.PaidInstallments++
			rp.RemainingInstallments--
			break
		}
	}
	
	if rp.RemainingInstallments == 0 {
		rp.Status = "COMPLETED"
	}
	rp.UpdatedAt = time.Now()
}

func (rp *RepaymentPlan) MarkOverdue(installmentID string) {
	for i, inst := range rp.Installments {
		if inst.ID == installmentID {
			rp.Installments[i].Status = "OVERDUE"
			break
		}
	}
	rp.Status = "OVERDUE"
	rp.UpdatedAt = time.Now()
}

type FinanceAccount struct {
	ID                string          `json:"id"`
	OwnerID           string          `json:"owner_id"`
	OwnerName         string          `json:"owner_name"`
	AccountType       string          `json:"account_type"`
	Balance           decimal.Decimal `json:"balance"`
	AvailableBalance  decimal.Decimal `json:"available_balance"`
	Currency          string          `json:"currency"`
	Status            string          `json:"status"`
	CreditLineIDs     []string        `json:"credit_line_ids"`
	FinancingIDs      []string        `json:"financing_ids"`
	OpenedAt          time.Time       `json:"opened_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func NewFinanceAccount(ownerID, ownerName, accountType, currency string) *FinanceAccount {
	return &FinanceAccount{
		OwnerID:          ownerID,
		OwnerName:        ownerName,
		AccountType:      accountType,
		Balance:          decimal.Zero,
		AvailableBalance: decimal.Zero,
		Currency:         currency,
		Status:           "ACTIVE",
		CreditLineIDs:    []string{},
		FinancingIDs:     []string{},
		OpenedAt:         time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (fa *FinanceAccount) Deposit(amount decimal.Decimal) {
	fa.Balance = fa.Balance.Add(amount)
	fa.AvailableBalance = fa.AvailableBalance.Add(amount)
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceAccount) Withdraw(amount decimal.Decimal) error {
	if fa.AvailableBalance.LessThan(amount) {
		return ErrInsufficientCreditLimit
	}
	fa.Balance = fa.Balance.Sub(amount)
	fa.AvailableBalance = fa.AvailableBalance.Sub(amount)
	fa.UpdatedAt = time.Now()
	return nil
}

func (fa *FinanceAccount) AddCreditLine(creditLineID string) {
	fa.CreditLineIDs = append(fa.CreditLineIDs, creditLineID)
	fa.UpdatedAt = time.Now()
}

func (fa *FinanceAccount) AddFinancing(financingID string) {
	fa.FinancingIDs = append(fa.FinancingIDs, financingID)
	fa.UpdatedAt = time.Now()
}

type FinanceApplicationRepository interface {
	Create(ctx context.Context, application *FinanceApplication) error
	Update(ctx context.Context, application *FinanceApplication) error
	FindByID(ctx context.Context, id string) (*FinanceApplication, error)
	FindByApplicantID(ctx context.Context, applicantID string) ([]*FinanceApplication, error)
	FindByStatus(ctx context.Context, status FinanceStatus, limit, offset int) ([]*FinanceApplication, int64, error)
	Delete(ctx context.Context, id string) error
}

type CreditLineRepository interface {
	Create(ctx context.Context, creditLine *CreditLine) error
	Update(ctx context.Context, creditLine *CreditLine) error
	FindByID(ctx context.Context, id string) (*CreditLine, error)
	FindByOwnerID(ctx context.Context, ownerID string) ([]*CreditLine, error)
	FindByStatus(ctx context.Context, status CreditLineStatus, limit, offset int) ([]*CreditLine, int64, error)
	Delete(ctx context.Context, id string) error
}

type InvoiceFinancingRepository interface {
	Create(ctx context.Context, financing *InvoiceFinancing) error
	Update(ctx context.Context, financing *InvoiceFinancing) error
	FindByID(ctx context.Context, id string) (*InvoiceFinancing, error)
	FindByBorrowerID(ctx context.Context, borrowerID string) ([]*InvoiceFinancing, error)
	FindByInvoiceID(ctx context.Context, invoiceID string) (*InvoiceFinancing, error)
	FindByStatus(ctx context.Context, status FinanceStatus, limit, offset int) ([]*InvoiceFinancing, int64, error)
	Delete(ctx context.Context, id string) error
}

type SupplierRiskProfileRepository interface {
	Create(ctx context.Context, profile *SupplierRiskProfile) error
	Update(ctx context.Context, profile *SupplierRiskProfile) error
	FindByID(ctx context.Context, id string) (*SupplierRiskProfile, error)
	FindBySupplierID(ctx context.Context, supplierID string) (*SupplierRiskProfile, error)
	FindByRiskLevel(ctx context.Context, level RiskLevel, limit, offset int) ([]*SupplierRiskProfile, int64, error)
	Delete(ctx context.Context, id string) error
}

type BuyerRiskProfileRepository interface {
	Create(ctx context.Context, profile *BuyerRiskProfile) error
	Update(ctx context.Context, profile *BuyerRiskProfile) error
	FindByID(ctx context.Context, id string) (*BuyerRiskProfile, error)
	FindByBuyerID(ctx context.Context, buyerID string) (*BuyerRiskProfile, error)
	FindByRiskLevel(ctx context.Context, level RiskLevel, limit, offset int) ([]*BuyerRiskProfile, int64, error)
	Delete(ctx context.Context, id string) error
}

type CollateralRepository interface {
	Create(ctx context.Context, collateral *Collateral) error
	Update(ctx context.Context, collateral *Collateral) error
	FindByID(ctx context.Context, id string) (*Collateral, error)
	FindByOwnerID(ctx context.Context, ownerID string) ([]*Collateral, error)
	Delete(ctx context.Context, id string) error
}

type GuaranteeRepository interface {
	Create(ctx context.Context, guarantee *Guarantee) error
	Update(ctx context.Context, guarantee *Guarantee) error
	FindByID(ctx context.Context, id string) (*Guarantee, error)
	FindByGuarantorID(ctx context.Context, guarantorID string) ([]*Guarantee, error)
	FindByBeneficiaryID(ctx context.Context, beneficiaryID string) ([]*Guarantee, error)
	Delete(ctx context.Context, id string) error
}

type RepaymentPlanRepository interface {
	Create(ctx context.Context, plan *RepaymentPlan) error
	Update(ctx context.Context, plan *RepaymentPlan) error
	FindByID(ctx context.Context, id string) (*RepaymentPlan, error)
	FindByFinancingID(ctx context.Context, financingID string) (*RepaymentPlan, error)
	FindByBorrowerID(ctx context.Context, borrowerID string) ([]*RepaymentPlan, error)
	Delete(ctx context.Context, id string) error
}

type FinanceAccountRepository interface {
	Create(ctx context.Context, account *FinanceAccount) error
	Update(ctx context.Context, account *FinanceAccount) error
	FindByID(ctx context.Context, id string) (*FinanceAccount, error)
	FindByOwnerID(ctx context.Context, ownerID string) (*FinanceAccount, error)
	Delete(ctx context.Context, id string) error
}
