package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTimeLimitNotFound      = errors.New("time limit not found")
	ErrTimeLimitExceeded      = errors.New("time limit exceeded")
	ErrInvalidTimeLimitConfig = errors.New("invalid time limit config")
)

type TimeLimitStatus int8

const (
	TimeLimitStatusActive    TimeLimitStatus = 1
	TimeLimitStatusExpired   TimeLimitStatus = 2
	TimeLimitStatusCompleted TimeLimitStatus = 3
)

type AfterSalesTimeLimit struct {
	ID              uint64           `json:"id"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	OrderID         uint64           `json:"order_id"`
	OrderNo         string           `json:"order_no"`
	UserID          uint64           `json:"user_id"`
	CompletedAt     time.Time        `json:"completed_at"`
	RefundDeadline  *time.Time       `json:"refund_deadline"`
	ReturnDeadline  *time.Time       `json:"return_deadline"`
	ExchangeDeadline *time.Time      `json:"exchange_deadline"`
	RepairDeadline  *time.Time       `json:"repair_deadline"`
	ComplaintDeadline *time.Time     `json:"complaint_deadline"`
	Status          TimeLimitStatus  `json:"status"`
	Items           []*TimeLimitItem `json:"items"`
}

type TimeLimitItem struct {
	ID               uint64          `json:"id"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	TimeLimitID      uint64          `json:"time_limit_id"`
	SkuID            uint64          `json:"sku_id"`
	ProductID        uint64          `json:"product_id"`
	ProductName      string          `json:"product_name"`
	CategoryID       uint64          `json:"category_id"`
	RefundDays       int             `json:"refund_days"`
	ReturnDays       int             `json:"return_days"`
	ExchangeDays     int             `json:"exchange_days"`
	RepairDays       int             `json:"repair_days"`
	RefundDeadline   *time.Time      `json:"refund_deadline"`
	ReturnDeadline   *time.Time      `json:"return_deadline"`
	ExchangeDeadline *time.Time      `json:"exchange_deadline"`
	RepairDeadline   *time.Time      `json:"repair_deadline"`
	UsedRefund       bool            `json:"used_refund"`
	UsedReturn       bool            `json:"used_return"`
	UsedExchange     bool            `json:"used_exchange"`
	UsedRepair       bool            `json:"used_repair"`
}

type TimeLimitConfig struct {
	DefaultRefundDays   int `json:"default_refund_days"`
	DefaultReturnDays   int `json:"default_return_days"`
	DefaultExchangeDays int `json:"default_exchange_days"`
	DefaultRepairDays   int `json:"default_repair_days"`
	ComplaintDays       int `json:"complaint_days"`
	CategoryConfigs     map[uint64]*CategoryTimeLimitConfig `json:"category_configs"`
}

type CategoryTimeLimitConfig struct {
	CategoryID       uint64 `json:"category_id"`
	RefundDays       int    `json:"refund_days"`
	ReturnDays       int    `json:"return_days"`
	ExchangeDays     int    `json:"exchange_days"`
	RepairDays       int    `json:"repair_days"`
	SpecialRules     []string `json:"special_rules"`
}

func DefaultTimeLimitConfig() *TimeLimitConfig {
	return &TimeLimitConfig{
		DefaultRefundDays:   7,
		DefaultReturnDays:   7,
		DefaultExchangeDays: 15,
		DefaultRepairDays:   30,
		ComplaintDays:       15,
		CategoryConfigs:     make(map[uint64]*CategoryTimeLimitConfig),
	}
}

func NewAfterSalesTimeLimit(orderID uint64, orderNo string, userID uint64, completedAt time.Time, config *TimeLimitConfig) *AfterSalesTimeLimit {
	tl := &AfterSalesTimeLimit{
		OrderID:      orderID,
		OrderNo:      orderNo,
		UserID:       userID,
		CompletedAt:  completedAt,
		Status:       TimeLimitStatusActive,
		Items:        []*TimeLimitItem{},
	}

	tl.RefundDeadline = calculateDeadline(completedAt, config.DefaultRefundDays)
	tl.ReturnDeadline = calculateDeadline(completedAt, config.DefaultReturnDays)
	tl.ExchangeDeadline = calculateDeadline(completedAt, config.DefaultExchangeDays)
	tl.RepairDeadline = calculateDeadline(completedAt, config.DefaultRepairDays)
	tl.ComplaintDeadline = calculateDeadline(completedAt, config.ComplaintDays)

	return tl
}

func calculateDeadline(from time.Time, days int) *time.Time {
	if days <= 0 {
		return nil
	}
	deadline := from.AddDate(0, 0, days)
	return &deadline
}

func (tl *AfterSalesTimeLimit) AddItem(skuID, productID, categoryID uint64, productName string, config *TimeLimitConfig) {
	item := &TimeLimitItem{
		TimeLimitID:  tl.ID,
		SkuID:        skuID,
		ProductID:    productID,
		ProductName:  productName,
		CategoryID:   categoryID,
		RefundDays:   config.DefaultRefundDays,
		ReturnDays:   config.DefaultReturnDays,
		ExchangeDays: config.DefaultExchangeDays,
		RepairDays:   config.DefaultRepairDays,
	}

	if catConfig, ok := config.CategoryConfigs[categoryID]; ok {
		item.RefundDays = catConfig.RefundDays
		item.ReturnDays = catConfig.ReturnDays
		item.ExchangeDays = catConfig.ExchangeDays
		item.RepairDays = catConfig.RepairDays
	}

	item.RefundDeadline = calculateDeadline(tl.CompletedAt, item.RefundDays)
	item.ReturnDeadline = calculateDeadline(tl.CompletedAt, item.ReturnDays)
	item.ExchangeDeadline = calculateDeadline(tl.CompletedAt, item.ExchangeDays)
	item.RepairDeadline = calculateDeadline(tl.CompletedAt, item.RepairDays)

	tl.Items = append(tl.Items, item)
}

func (tl *AfterSalesTimeLimit) CanApply(afterSalesType AfterSalesType, skuID uint64) (bool, *time.Time) {
	now := time.Now()

	switch afterSalesType {
	case AfterSalesTypeRefund:
		if tl.RefundDeadline == nil || now.After(*tl.RefundDeadline) {
			return false, tl.RefundDeadline
		}
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				if item.UsedRefund {
					return false, nil
				}
				if item.RefundDeadline != nil && now.After(*item.RefundDeadline) {
					return false, item.RefundDeadline
				}
				return true, item.RefundDeadline
			}
		}
	case AfterSalesTypeReturnGoods:
		if tl.ReturnDeadline == nil || now.After(*tl.ReturnDeadline) {
			return false, tl.ReturnDeadline
		}
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				if item.UsedReturn {
					return false, nil
				}
				if item.ReturnDeadline != nil && now.After(*item.ReturnDeadline) {
					return false, item.ReturnDeadline
				}
				return true, item.ReturnDeadline
			}
		}
	case AfterSalesTypeExchange:
		if tl.ExchangeDeadline == nil || now.After(*tl.ExchangeDeadline) {
			return false, tl.ExchangeDeadline
		}
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				if item.UsedExchange {
					return false, nil
				}
				if item.ExchangeDeadline != nil && now.After(*item.ExchangeDeadline) {
					return false, item.ExchangeDeadline
				}
				return true, item.ExchangeDeadline
			}
		}
	case AfterSalesTypeRepair:
		if tl.RepairDeadline == nil || now.After(*tl.RepairDeadline) {
			return false, tl.RepairDeadline
		}
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				if item.UsedRepair {
					return false, nil
				}
				if item.RepairDeadline != nil && now.After(*item.RepairDeadline) {
					return false, item.RepairDeadline
				}
				return true, item.RepairDeadline
			}
		}
	case AfterSalesTypeComplaint:
		if tl.ComplaintDeadline == nil || now.After(*tl.ComplaintDeadline) {
			return false, tl.ComplaintDeadline
		}
		return true, tl.ComplaintDeadline
	}

	return false, nil
}

func (tl *AfterSalesTimeLimit) MarkUsed(afterSalesType AfterSalesType, skuID uint64) {
	for _, item := range tl.Items {
		if item.SkuID == skuID {
			switch afterSalesType {
			case AfterSalesTypeRefund:
				item.UsedRefund = true
			case AfterSalesTypeReturnGoods:
				item.UsedReturn = true
			case AfterSalesTypeExchange:
				item.UsedExchange = true
			case AfterSalesTypeRepair:
				item.UsedRepair = true
			}
			break
		}
	}
}

func (tl *AfterSalesTimeLimit) GetRemainingDays(afterSalesType AfterSalesType, skuID uint64) int {
	var deadline *time.Time
	switch afterSalesType {
	case AfterSalesTypeRefund:
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				deadline = item.RefundDeadline
				break
			}
		}
		if deadline == nil {
			deadline = tl.RefundDeadline
		}
	case AfterSalesTypeReturnGoods:
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				deadline = item.ReturnDeadline
				break
			}
		}
		if deadline == nil {
			deadline = tl.ReturnDeadline
		}
	case AfterSalesTypeExchange:
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				deadline = item.ExchangeDeadline
				break
			}
		}
		if deadline == nil {
			deadline = tl.ExchangeDeadline
		}
	case AfterSalesTypeRepair:
		for _, item := range tl.Items {
			if item.SkuID == skuID {
				deadline = item.RepairDeadline
				break
			}
		}
		if deadline == nil {
			deadline = tl.RepairDeadline
		}
	}

	if deadline == nil {
		return 0
	}

	remaining := time.Until(*deadline)
	days := int(remaining.Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func (tl *AfterSalesTimeLimit) IsExpired() bool {
	now := time.Now()
	if tl.RefundDeadline != nil && now.Before(*tl.RefundDeadline) {
		return false
	}
	if tl.ReturnDeadline != nil && now.Before(*tl.ReturnDeadline) {
		return false
	}
	if tl.ExchangeDeadline != nil && now.Before(*tl.ExchangeDeadline) {
		return false
	}
	if tl.RepairDeadline != nil && now.Before(*tl.RepairDeadline) {
		return false
	}
	if tl.ComplaintDeadline != nil && now.Before(*tl.ComplaintDeadline) {
		return false
	}
	return true
}

func (tl *AfterSalesTimeLimit) MarkCompleted() {
	tl.Status = TimeLimitStatusCompleted
}

func (s TimeLimitStatus) String() string {
	switch s {
	case TimeLimitStatusActive:
		return "ACTIVE"
	case TimeLimitStatusExpired:
		return "EXPIRED"
	case TimeLimitStatusCompleted:
		return "COMPLETED"
	default:
		return "UNKNOWN"
	}
}

type AfterSalesTimeLimitRepository interface {
	Save(ctx context.Context, limit *AfterSalesTimeLimit) error
	FindByID(ctx context.Context, id uint64) (*AfterSalesTimeLimit, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*AfterSalesTimeLimit, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*AfterSalesTimeLimit, error)
	FindExpiringSoon(ctx context.Context, days int) ([]*AfterSalesTimeLimit, error)
	Update(ctx context.Context, limit *AfterSalesTimeLimit) error
}
