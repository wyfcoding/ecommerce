package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InstallmentStatus 分期状态
type InstallmentStatus int8

const (
	InstallmentStatusActive    InstallmentStatus = 1 // 正常进行中
	InstallmentStatusCompleted InstallmentStatus = 2 // 已完成
	InstallmentStatusOverdue   InstallmentStatus = 3 // 逾期
	InstallmentStatusCancelled InstallmentStatus = 4 // 已取消
)

// RepaymentMethod 还款方式
type RepaymentMethod int8

const (
	MethodEqualPrincipalAndInterest RepaymentMethod = 1 // 等额本息
	MethodEqualPrincipal            RepaymentMethod = 2 // 等额本金
)

// InstallmentPlan 分期还款计划
type InstallmentPlan struct {
	gorm.Model
	PlanID        string            `gorm:"column:plan_id;type:varchar(32);unique_index;not null"`
	InvoiceNo     string            `gorm:"column:invoice_no;type:varchar(32);index;not null"`
	UserID        uint64            `gorm:"column:user_id;index;not null"`
	TotalAmount   decimal.Decimal   `gorm:"column:total_amount;type:decimal(20,2);not null"`
	TotalInterest decimal.Decimal   `gorm:"column:total_interest;type:decimal(20,2);not null;default:0"`
	PeriodCount   int               `gorm:"column:period_count;not null"`           // 分期期数
	Apr           decimal.Decimal   `gorm:"column:apr;type:decimal(10,4);not null"` // 年化利率 %
	Method        RepaymentMethod   `gorm:"column:method;type:tinyint;not null"`
	Status        InstallmentStatus `gorm:"column:status;type:tinyint;not null;default:1"`
	StartDate     time.Time         `gorm:"column:start_date;not null"`

	Details []InstallmentDetail `gorm:"foreignKey:PlanID;references:PlanID"`
}

// InstallmentDetail 分期明细
type InstallmentDetail struct {
	gorm.Model
	PlanID       string          `gorm:"column:plan_id;type:varchar(32);index;not null"`
	PeriodNo     int             `gorm:"column:period_no;not null"` // 第几期
	DueDate      time.Time       `gorm:"column:due_date;not null"`
	Principal    decimal.Decimal `gorm:"column:principal;type:decimal(20,2);not null"`     // 本金
	Interest     decimal.Decimal `gorm:"column:interest;type:decimal(20,2);not null"`      // 利息
	TotalPayment decimal.Decimal `gorm:"column:total_payment;type:decimal(20,2);not null"` // 本期应还总额
	PaidAmount   decimal.Decimal `gorm:"column:paid_amount;type:decimal(20,2);not null;default:0"`
	Status       int8            `gorm:"column:status;not null;default:0"` // 0-待还, 1-已还, 2-逾期
	PaidAt       *time.Time      `gorm:"column:paid_at"`
}

// NewInstallmentPlan 创建分期计划（自动计算明细）
func NewInstallmentPlan(id, invoiceNo string, userID uint64, amount, apr decimal.Decimal, periods int, method RepaymentMethod, start time.Time) (*InstallmentPlan, error) {
	if periods <= 0 {
		return nil, errors.New("periods must be positive")
	}

	plan := &InstallmentPlan{
		PlanID:      id,
		InvoiceNo:   invoiceNo,
		UserID:      userID,
		TotalAmount: amount,
		PeriodCount: periods,
		Apr:         apr,
		Method:      method,
		Status:      InstallmentStatusActive,
		StartDate:   start,
		Details:     make([]InstallmentDetail, 0, periods),
	}

	if err := plan.calculateDetails(); err != nil {
		return nil, err
	}

	return plan, nil
}

// calculateDetails 计算每期还款
func (p *InstallmentPlan) calculateDetails() error {
	monthlyRate := p.Apr.Div(decimal.NewFromInt(1200)) // 年化% -> 月利率

	var totalInterest decimal.Decimal

	if p.Method == MethodEqualPrincipalAndInterest {
		// 等额本息: A = P * [i * (1+i)^n] / [(1+i)^n - 1]
		// pow := (1+i)^n
		onePlusRate := monthlyRate.Add(decimal.NewFromInt(1))
		pow := onePlusRate.Pow(decimal.NewFromInt(int64(p.PeriodCount)))

		numerator := p.TotalAmount.Mul(monthlyRate).Mul(pow)
		denominator := pow.Sub(decimal.NewFromInt(1))

		monthlyPayment := numerator.Div(denominator)

		remainingPrincipal := p.TotalAmount

		for i := 1; i <= p.PeriodCount; i++ {
			interest := remainingPrincipal.Mul(monthlyRate)
			principal := monthlyPayment.Sub(interest)

			// 最后一期调整误差
			if i == p.PeriodCount {
				principal = remainingPrincipal
				monthlyPayment = principal.Add(interest)
			}

			detail := InstallmentDetail{
				PlanID:       p.PlanID,
				PeriodNo:     i,
				DueDate:      p.StartDate.AddDate(0, i, 0),
				Principal:    principal.Round(2),
				Interest:     interest.Round(2),
				TotalPayment: monthlyPayment.Round(2),
			}
			p.Details = append(p.Details, detail)

			remainingPrincipal = remainingPrincipal.Sub(principal)
			totalInterest = totalInterest.Add(interest)
		}
	} else if p.Method == MethodEqualPrincipal {
		// 等额本金: 每月本金 = 总本金 / n, 利息 = 剩余本金 * 月利率
		monthlyPrincipal := p.TotalAmount.Div(decimal.NewFromInt(int64(p.PeriodCount)))
		remainingPrincipal := p.TotalAmount

		for i := 1; i <= p.PeriodCount; i++ {
			interest := remainingPrincipal.Mul(monthlyRate)
			principal := monthlyPrincipal

			if i == p.PeriodCount {
				principal = remainingPrincipal
			}

			detail := InstallmentDetail{
				PlanID:       p.PlanID,
				PeriodNo:     i,
				DueDate:      p.StartDate.AddDate(0, i, 0),
				Principal:    principal.Round(2),
				Interest:     interest.Round(2),
				TotalPayment: principal.Add(interest).Round(2),
			}
			p.Details = append(p.Details, detail)

			remainingPrincipal = remainingPrincipal.Sub(principal)
			totalInterest = totalInterest.Add(interest)
		}
	} else {
		return errors.New("unsupported repayment method")
	}

	p.TotalInterest = totalInterest.Round(2)
	return nil
}

// TableName 表名
func (InstallmentPlan) TableName() string {
	return "installment_plans"
}

func (InstallmentDetail) TableName() string {
	return "installment_details"
}
