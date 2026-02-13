package domain

import (
	"errors"
)

var (
	ErrInvalidStatus      = errors.New("invalid settlement status for this operation")
	ErrSettlementNotFound = errors.New("settlement not found")
	ErrMerchantNotFound   = errors.New("merchant not found")
	ErrBankAccountNotFound = errors.New("bank account not found")
	ErrConfigNotFound     = errors.New("settlement config not found")
	ErrAmountTooSmall     = errors.New("settlement amount is below minimum")
	ErrDuplicateSettlement = errors.New("settlement already exists for this period")
)
