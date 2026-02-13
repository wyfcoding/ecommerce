package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidInstallmentCount = errors.New("invalid installment count")
	ErrInstallmentNotSupported = errors.New("installment not supported for this payment")
	ErrInstallmentAlreadyPaid  = errors.New("installment already paid")
	ErrInstallmentNotDue       = errors.New("installment not due yet")
)

type InstallmentStatus int8

const (
	InstallmentStatusPending InstallmentStatus = iota
	InstallmentStatusPaid
	InstallmentStatusOverdue
	InstallmentStatusCancelled
)

type InstallmentPlan struct {
	ID               uint64              `json:"id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	PaymentID        uint64              `json:"payment_id"`
	PaymentNo        string              `json:"payment_no"`
	OrderID          uint64              `json:"order_id"`
	UserID           uint64             `json:"user_id"`
	TotalAmount      int64               `json:"total_amount"`
	InstallmentCount int                 `json:"installment_count"`
	InterestRate     float64             `json:"interest_rate"`
	InterestAmount   int64               `json:"interest_amount"`
	TotalPayable     int64               `json:"total_payable"`
	PerInstallment   int64               `json:"per_installment"`
	FirstPayAmount   int64               `json:"first_pay_amount"`
	RemainingAmount  int64               `json:"remaining_amount"`
	Status           InstallmentStatus   `json:"status"`
	Installments     []*Installment      `json:"installments"`
	Provider         string              `json:"provider"`
	ProviderPlanID   string              `json:"provider_plan_id"`
}

type Installment struct {
	ID              uint64            `json:"id"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	PlanID          uint64            `json:"plan_id"`
	InstallmentNum  int               `json:"installment_num"`
	PrincipalAmount int64             `json:"principal_amount"`
	InterestAmount  int64             `json:"interest_amount"`
	TotalAmount     int64             `json:"total_amount"`
	DueDate         time.Time         `json:"due_date"`
	PaidDate        *time.Time        `json:"paid_date"`
	Status          InstallmentStatus `json:"status"`
	TransactionID   string            `json:"transaction_id"`
	OverdueDays     int               `json:"overdue_days"`
	OverdueFee      int64             `json:"overdue_fee"`
}

type InstallmentProvider string

const (
	InstallmentProviderHuabei   InstallmentProvider = "huabei"
	InstallmentProviderBaitiao  InstallmentProvider = "baitiao"
	InstallmentProviderAnt      InstallmentProvider = "ant_credit"
	InstallmentProviderInternal InstallmentProvider = "internal"
)

type InstallmentPlanConfig struct {
	MinAmount        int64
	MaxAmount        int64
	MinInstallments  int
	MaxInstallments  int
	InterestRates    map[int]float64
	AllowPrepayment  bool
	PrepaymentFee    float64
	OverdueFeeRate   float64
	GracePeriodDays  int
}

func DefaultInstallmentConfig() *InstallmentPlanConfig {
	return &InstallmentPlanConfig{
		MinAmount:       10000,
		MaxAmount:       100000000,
		MinInstallments: 3,
		MaxInstallments: 24,
		InterestRates: map[int]float64{
			3:  0.025,
			6:  0.045,
			9:  0.065,
			12: 0.088,
			24: 0.15,
		},
		AllowPrepayment: true,
		PrepaymentFee:   0.01,
		OverdueFeeRate:  0.0005,
		GracePeriodDays: 3,
	}
}

func NewInstallmentPlan(paymentID uint64, paymentNo string, orderID, userID uint64, totalAmount int64, installmentCount int, config *InstallmentPlanConfig) (*InstallmentPlan, error) {
	if installmentCount < config.MinInstallments || installmentCount > config.MaxInstallments {
		return nil, ErrInvalidInstallmentCount
	}
	if totalAmount < config.MinAmount || totalAmount > config.MaxAmount {
		return nil, ErrInstallmentNotSupported
	}

	interestRate, ok := config.InterestRates[installmentCount]
	if !ok {
		interestRate = 0.05
	}

	interestAmount := int64(float64(totalAmount) * interestRate)
	totalPayable := totalAmount + interestAmount
	perInstallment := totalPayable / int64(installmentCount)
	firstPayAmount := totalPayable - perInstallment*int64(installmentCount-1)

	plan := &InstallmentPlan{
		PaymentID:        paymentID,
		PaymentNo:        paymentNo,
		OrderID:          orderID,
		UserID:           userID,
		TotalAmount:      totalAmount,
		InstallmentCount: installmentCount,
		InterestRate:     interestRate,
		InterestAmount:   interestAmount,
		TotalPayable:     totalPayable,
		PerInstallment:   perInstallment,
		FirstPayAmount:   firstPayAmount,
		RemainingAmount:  totalPayable,
		Status:           InstallmentStatusPending,
		Installments:     []*Installment{},
	}

	now := time.Now()
	for i := 1; i <= installmentCount; i++ {
		dueDate := now.AddDate(0, 0, i*30)
		amount := perInstallment
		if i == 1 {
			amount = firstPayAmount
		}
		principal := totalAmount / int64(installmentCount)
		interest := interestAmount / int64(installmentCount)
		if i == installmentCount {
			principal = totalAmount - principal*int64(installmentCount-1)
			interest = interestAmount - interest*int64(installmentCount-1)
		}

		installment := &Installment{
			PlanID:          plan.ID,
			InstallmentNum:  i,
			PrincipalAmount: principal,
			InterestAmount:  interest,
			TotalAmount:     amount,
			DueDate:         dueDate,
			Status:          InstallmentStatusPending,
		}
		plan.Installments = append(plan.Installments, installment)
	}

	return plan, nil
}

func (p *InstallmentPlan) GetCurrentInstallment() *Installment {
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusPending {
			return inst
		}
	}
	return nil
}

func (p *InstallmentPlan) GetNextDueDate() *time.Time {
	inst := p.GetCurrentInstallment()
	if inst != nil {
		return &inst.DueDate
	}
	return nil
}

func (p *InstallmentPlan) PayInstallment(installmentNum int, transactionID string) error {
	var installment *Installment
	for _, inst := range p.Installments {
		if inst.InstallmentNum == installmentNum {
			installment = inst
			break
		}
	}

	if installment == nil {
		return fmt.Errorf("installment %d not found", installmentNum)
	}

	if installment.Status == InstallmentStatusPaid {
		return ErrInstallmentAlreadyPaid
	}

	now := time.Now()
	installment.Status = InstallmentStatusPaid
	installment.PaidDate = &now
	installment.TransactionID = transactionID

	p.RemainingAmount -= installment.TotalAmount

	allPaid := true
	for _, inst := range p.Installments {
		if inst.Status != InstallmentStatusPaid {
			allPaid = false
			break
		}
	}
	if allPaid {
		p.Status = InstallmentStatusPaid
	}

	return nil
}

func (p *InstallmentPlan) MarkOverdue(config *InstallmentPlanConfig) {
	now := time.Now()
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusPending && now.After(inst.DueDate.AddDate(0, 0, config.GracePeriodDays)) {
			inst.Status = InstallmentStatusOverdue
			inst.OverdueDays = int(now.Sub(inst.DueDate).Hours() / 24)
			inst.OverdueFee = int64(float64(inst.TotalAmount) * config.OverdueFeeRate * float64(inst.OverdueDays))
		}
	}
}

func (p *InstallmentPlan) GetTotalOverdueFee() int64 {
	var total int64
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusOverdue {
			total += inst.OverdueFee
		}
	}
	return total
}

func (p *InstallmentPlan) CalculatePrepaymentAmount() int64 {
	var remaining int64
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusPending {
			remaining += inst.TotalAmount
		}
	}
	return remaining
}

func (p *InstallmentPlan) GetPaidCount() int {
	count := 0
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusPaid {
			count++
		}
	}
	return count
}

func (p *InstallmentPlan) GetPendingCount() int {
	count := 0
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusPending {
			count++
		}
	}
	return count
}

func (p *InstallmentPlan) GetOverdueCount() int {
	count := 0
	for _, inst := range p.Installments {
		if inst.Status == InstallmentStatusOverdue {
			count++
		}
	}
	return count
}

type InstallmentPlanRepository interface {
	Save(ctx context.Context, plan *InstallmentPlan) error
	FindByID(ctx context.Context, id uint64) (*InstallmentPlan, error)
	FindByPaymentID(ctx context.Context, paymentID uint64) (*InstallmentPlan, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*InstallmentPlan, error)
	FindOverdue(ctx context.Context) ([]*InstallmentPlan, error)
	Update(ctx context.Context, plan *InstallmentPlan) error
}
