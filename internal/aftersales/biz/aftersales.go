package biz

import (
	"context"
	"errors"
	"time"
)

// ErrReturnRequestNotFound is a specific error for when a return request is not found.
var ErrReturnRequestNotFound = errors.New("return request not found")

// ErrRefundRequestNotFound is a specific error for when a refund request is not found.
var ErrRefundRequestNotFound = errors.New("refund request not found")

// ReturnRequest represents a return request in the business layer.
type ReturnRequest struct {
	ID        uint
	OrderID   string
	UserID    string
	ProductID string
	Quantity  int32
	Reason    string
	Status    string // e.g., PENDING, APPROVED, REJECTED, RECEIVED, REFUNDED
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RefundRequest represents a refund request in the business layer.
type RefundRequest struct {
	ID              uint
	ReturnRequestID string
	OrderID         string
	UserID          string
	Amount          float64
	Currency        string
	Status          string // e.g., PENDING, APPROVED, REJECTED, COMPLETED
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AftersalesRepo defines the data storage interface for aftersales data.
// The business layer depends on this interface, not on a concrete data implementation.
type AftersalesRepo interface {
	CreateReturnRequest(ctx context.Context, req *ReturnRequest) (*ReturnRequest, error)
	GetReturnRequest(ctx context.Context, id uint) (*ReturnRequest, error)
	UpdateReturnRequest(ctx context.Context, req *ReturnRequest) (*ReturnRequest, error)

	CreateRefundRequest(ctx context.Context, req *RefundRequest) (*RefundRequest, error)
	GetRefundRequest(ctx context.Context, id uint) (*RefundRequest, error)
	UpdateRefundRequest(ctx context.Context, req *RefundRequest) (*RefundRequest, error)
}

// AftersalesUsecase is the use case for aftersales operations.
// It orchestrates the business logic.
type AftersalesUsecase struct {
	repo AftersalesRepo
	// You can also inject other dependencies like a logger or an order service client
}

// NewAftersalesUsecase creates a new AftersalesUsecase.
func NewAftersalesUsecase(repo AftersalesRepo) *AftersalesUsecase {
	return &AftersalesUsecase{repo: repo}
}

// CreateReturnRequest creates a new return request.
func (uc *AftersalesUsecase) CreateReturnRequest(ctx context.Context, orderID, userID, productID string, quantity int32, reason string) (*ReturnRequest, error) {
	// TODO: Add business logic, e.g., validate order, check product eligibility for return
	req := &ReturnRequest{
		OrderID:   orderID,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Reason:    reason,
		Status:    "PENDING", // Initial status
	}
	return uc.repo.CreateReturnRequest(ctx, req)
}

// GetReturnRequest retrieves a return request by ID.
func (uc *AftersalesUsecase) GetReturnRequest(ctx context.Context, id uint) (*ReturnRequest, error) {
	return uc.repo.GetReturnRequest(ctx, id)
}

// UpdateReturnRequestStatus updates the status of a return request.
func (uc *AftersalesUsecase) UpdateReturnRequestStatus(ctx context.Context, id uint, status string) (*ReturnRequest, error) {
	req, err := uc.repo.GetReturnRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Status = status
	return uc.repo.UpdateReturnRequest(ctx, req)
}

// CreateRefundRequest creates a new refund request.
func (uc *AftersalesUsecase) CreateRefundRequest(ctx context.Context, returnRequestID, orderID, userID string, amount float64, currency string) (*RefundRequest, error) {
	// TODO: Add business logic, e.g., validate return request, check payment status
	req := &RefundRequest{
		ReturnRequestID: returnRequestID,
		OrderID:         orderID,
		UserID:          userID,
		Amount:          amount,
		Currency:        currency,
		Status:          "PENDING", // Initial status
	}
	return uc.repo.CreateRefundRequest(ctx, req)
}

// GetRefundRequest retrieves a refund request by ID.
func (uc *AftersalesUsecase) GetRefundRequest(ctx context.Context, id uint) (*RefundRequest, error) {
	return uc.repo.GetRefundRequest(ctx, id)
}

// UpdateRefundRequestStatus updates the status of a refund request.
func (uc *AftersalesUsecase) UpdateRefundRequestStatus(ctx context.Context, id uint, status string) (*RefundRequest, error) {
	req, err := uc.repo.GetRefundRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Status = status
	return uc.repo.UpdateRefundRequest(ctx, req)
}
