package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

var (
	ErrPriceAdjustmentNotAllowed = errors.New("price adjustment not allowed in current status")
	ErrInvalidPriceAdjustment    = errors.New("invalid price adjustment amount")
	ErrPriceAdjustmentTooLarge   = errors.New("price adjustment exceeds maximum allowed percentage")
)

type PriceAdjustmentType string

const (
	PriceAdjustmentTypeDiscount PriceAdjustmentType = "DISCOUNT"
	PriceAdjustmentTypeIncrease PriceAdjustmentType = "INCREASE"
	PriceAdjustmentTypeOverride PriceAdjustmentType = "OVERRIDE"
)

type PriceAdjustment struct {
	ID              uint                `json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	OrderID         uint64              `json:"order_id"`
	OrderNo         string              `json:"order_no"`
	OperatorID      uint64              `json:"operator_id"`
	OperatorName    string              `json:"operator_name"`
	AdjustmentType  PriceAdjustmentType `json:"adjustment_type"`
	OriginalAmount  int64               `json:"original_amount"`
	AdjustmentValue int64               `json:"adjustment_value"`
	FinalAmount     int64               `json:"final_amount"`
	Reason          string              `json:"reason"`
	ApprovalStatus  ApprovalStatus      `json:"approval_status"`
	ApprovedBy      uint64              `json:"approved_by"`
	ApprovedAt      *time.Time          `json:"approved_at"`
	RejectedAt      *time.Time          `json:"rejected_at"`
	RejectionReason string              `json:"rejection_reason"`
}

type ApprovalStatus int8

const (
	ApprovalStatusPending ApprovalStatus = 0
	ApprovalStatusApproved ApprovalStatus = 1
	ApprovalStatusRejected ApprovalStatus = 2
)

type PriceAdjustmentConfig struct {
	MaxDiscountPercent float64
	MaxIncreasePercent float64
	RequireApproval    bool
	ApprovalThreshold  int64
}

func DefaultPriceAdjustmentConfig() *PriceAdjustmentConfig {
	return &PriceAdjustmentConfig{
		MaxDiscountPercent: 30.0,
		MaxIncreasePercent: 10.0,
		RequireApproval:    true,
		ApprovalThreshold:  10000,
	}
}

func NewPriceAdjustment(orderID uint64, orderNo string, operatorID uint64, operatorName string,
	adjustmentType PriceAdjustmentType, originalAmount, adjustmentValue int64, reason string) *PriceAdjustment {
	var finalAmount int64
	switch adjustmentType {
	case PriceAdjustmentTypeDiscount:
		finalAmount = originalAmount - adjustmentValue
	case PriceAdjustmentTypeIncrease:
		finalAmount = originalAmount + adjustmentValue
	case PriceAdjustmentTypeOverride:
		finalAmount = adjustmentValue
	}

	return &PriceAdjustment{
		OrderID:         orderID,
		OrderNo:         orderNo,
		OperatorID:      operatorID,
		OperatorName:    operatorName,
		AdjustmentType:  adjustmentType,
		OriginalAmount:  originalAmount,
		AdjustmentValue: adjustmentValue,
		FinalAmount:     finalAmount,
		Reason:          reason,
		ApprovalStatus:  ApprovalStatusPending,
	}
}

func (o *Order) CanAdjustPrice() bool {
	return o.Status == pb.OrderStatus_PENDING_PAYMENT || o.Status == pb.OrderStatus_ALLOCATING
}

func (o *Order) AdjustPrice(ctx context.Context, adjustment *PriceAdjustment, config *PriceAdjustmentConfig) error {
	if !o.CanAdjustPrice() {
		return ErrPriceAdjustmentNotAllowed
	}

	if adjustment.FinalAmount <= 0 {
		return ErrInvalidPriceAdjustment
	}

	switch adjustment.AdjustmentType {
	case PriceAdjustmentTypeDiscount:
		maxDiscount := int64(float64(o.TotalAmount) * config.MaxDiscountPercent / 100)
		if adjustment.AdjustmentValue > maxDiscount {
			return fmt.Errorf("%w: max allowed is %d", ErrPriceAdjustmentTooLarge, maxDiscount)
		}
	case PriceAdjustmentTypeIncrease:
		maxIncrease := int64(float64(o.TotalAmount) * config.MaxIncreasePercent / 100)
		if adjustment.AdjustmentValue > maxIncrease {
			return fmt.Errorf("%w: max allowed is %d", ErrPriceAdjustmentTooLarge, maxIncrease)
		}
	}

	o.ActualAmount = adjustment.FinalAmount
	o.DiscountAmount = o.TotalAmount - adjustment.FinalAmount

	o.AddLog(adjustment.OperatorName, "PriceAdjusted", o.Status.String(), o.Status.String(),
		fmt.Sprintf("Type: %s, Value: %d, Reason: %s", adjustment.AdjustmentType, adjustment.AdjustmentValue, adjustment.Reason))

	return nil
}

func (a *PriceAdjustment) Approve(approverID uint64) {
	a.ApprovalStatus = ApprovalStatusApproved
	a.ApprovedBy = approverID
	now := time.Now()
	a.ApprovedAt = &now
}

func (a *PriceAdjustment) Reject(approverID uint64, reason string) {
	a.ApprovalStatus = ApprovalStatusRejected
	a.RejectionReason = reason
	now := time.Now()
	a.RejectedAt = &now
}

func (a *PriceAdjustment) IsApproved() bool {
	return a.ApprovalStatus == ApprovalStatusApproved
}

func (a *PriceAdjustment) IsPending() bool {
	return a.ApprovalStatus == ApprovalStatusPending
}

type PriceAdjustmentRepository interface {
	Save(ctx context.Context, adjustment *PriceAdjustment) error
	FindByID(ctx context.Context, id uint64) (*PriceAdjustment, error)
	FindByOrderID(ctx context.Context, orderID uint64) ([]*PriceAdjustment, error)
	FindPendingApprovals(ctx context.Context, limit, offset int) ([]*PriceAdjustment, error)
	Update(ctx context.Context, adjustment *PriceAdjustment) error
}
