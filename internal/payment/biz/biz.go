package biz

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrInvalidAmount = errors.New("invalid payment amount")
	ErrPaymentFailed = errors.New("payment failed")
)

// PaymentTransaction represents a payment transaction in the business logic layer.
type PaymentTransaction struct {
	ID            uint
	PaymentID     string // Unique ID from payment system
	OrderID       uint64
	UserID        uint64
	Amount        uint64
	Currency      string
	PaymentMethod string
	Status        string // PENDING, SUCCESS, FAILED
	TransactionNo string // From payment gateway
	CallbackData  string
	PaidAt        *time.Time
}

// PaymentRepo defines the interface for payment data access.
type PaymentRepo interface {
	CreatePaymentTransaction(ctx context.Context, tx *PaymentTransaction) (*PaymentTransaction, error)
	GetPaymentTransactionByPaymentID(ctx context.Context, paymentID string) (*PaymentTransaction, error)
	UpdatePaymentTransactionStatus(ctx context.Context, paymentID string, newStatus string, transactionNo string, callbackData string, paidAt *time.Time) error
}

// OrderInfo represents basic order info from Order Service.
type OrderInfo struct {
	ID          uint64
	UserID      uint64
	TotalAmount uint64
	Status      int8
}

// OrderClient defines the interface to interact with the Order Service.
type OrderClient interface {
	GetOrder(ctx context.Context, orderID uint64) (*OrderInfo, error)
	UpdateOrderStatus(ctx context.Context, orderID uint64, newStatus int8) error
}

// PaymentUsecase is the business logic for payment processing.
type PaymentUsecase struct {
	repo        PaymentRepo
	orderClient OrderClient
	// TODO: Add clients for external payment gateways (e.g., Alipay, WeChat Pay)
}

// NewPaymentUsecase creates a new PaymentUsecase.
func NewPaymentUsecase(repo PaymentRepo, orderClient OrderClient) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, orderClient: orderClient}
}

// CreatePayment initiates a payment process.
func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID, userID, amount uint64, currency, paymentMethod, returnURL string) (*PaymentTransaction, string, string, error) {
	// 1. Get order details
	order, err := uc.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		return nil, "", "", err
	}
	if order == nil {
		return nil, "", "", ErrOrderNotFound
	}
	if order.TotalAmount != amount {
		return nil, "", "", ErrInvalidAmount // Amount mismatch
	}
	// TODO: Check order status (should be pending payment)

	// 2. Create a pending payment transaction record
	paymentID := fmt.Sprintf("PAY_%d_%d", orderID, time.Now().UnixNano()) // Generate a unique payment ID
	tx := &PaymentTransaction{
		PaymentID:     paymentID,
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		Status:        "PENDING",
	}
	createdTx, err := uc.repo.CreatePaymentTransaction(ctx, tx)
	if err != nil {
		return nil, "", "", err
	}

	// 3. Call external payment gateway (simulated)
	// In a real system, this would involve calling Alipay/WeChat Pay SDKs
	redirectURL := fmt.Sprintf("http://mock-payment-gateway.com/pay?id=%s&amount=%d", paymentID, amount)
	qrCodeURL := fmt.Sprintf("http://mock-payment-gateway.com/qrcode?id=%s", paymentID)

	return createdTx, redirectURL, qrCodeURL, nil
}

// HandlePaymentCallback processes payment gateway callbacks.
func (uc *PaymentUsecase) HandlePaymentCallback(ctx context.Context, paymentMethod string, callbackData map[string]string) (*PaymentTransaction, error) {
	// 1. Validate callback data (e.g., signature verification)
	// This is highly dependent on the payment gateway.
	// For simplicity, assume data is valid and extract paymentID and status.
	paymentID := callbackData["payment_id"]
	statusFromGateway := callbackData["status"]
	transactionNo := callbackData["transaction_no"]

	// 2. Get the payment transaction record
	tx, err := uc.repo.GetPaymentTransactionByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, fmt.Errorf("payment transaction %s not found", paymentID)
	}
	if tx.Status == "SUCCESS" {
		return tx, nil // Already processed
	}

	// 3. Update payment transaction status
	newStatus := "FAILED"
	if statusFromGateway == "SUCCESS" {
		newStatus = "SUCCESS"
	}

	paidAt := time.Now()
	err = uc.repo.UpdatePaymentTransactionStatus(ctx, paymentID, newStatus, transactionNo, fmt.Sprintf("%v", callbackData), &paidAt)
	if err != nil {
		return nil, err
	}
	tx.Status = newStatus
	tx.TransactionNo = transactionNo
	tx.CallbackData = fmt.Sprintf("%v", callbackData)
	tx.PaidAt = &paidAt

	// 4. Update order status (if payment successful)
	if newStatus == "SUCCESS" {
		err = uc.orderClient.UpdateOrderStatus(ctx, tx.OrderID, 2) // Assuming 2 is OrderStatusPaid
		if err != nil {
			// Log error, but payment transaction is already updated.
			// This might require a separate retry mechanism or manual intervention.
			fmt.Printf("failed to update order status for order %d: %v\n", tx.OrderID, err)
		}
	}

	return tx, nil
}
