package biz

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrFlashSaleEventNotFound is a specific error for when a flash sale event is not found.
var ErrFlashSaleEventNotFound = errors.New("flash sale event not found")

// ErrFlashSaleProductNotFound is a specific error for when a flash sale product is not found.
var ErrFlashSaleProductNotFound = errors.New("flash sale product not found")

// ErrFlashSaleNotInProgress is a specific error for when trying to participate in a flash sale that is not active.
var ErrFlashSaleNotInProgress = errors.New("flash sale not in progress")

// ErrFlashSaleOutOfStock is a specific error for when a flash sale product is out of stock.
var ErrFlashSaleOutOfStock = errors.New("flash sale product out of stock")

// ErrFlashSaleMaxPerUserExceeded is a specific error for when a user tries to buy more than allowed.
var ErrFlashSaleMaxPerUserExceeded = errors.New("max quantity per user exceeded")

// ErrAcquireLockFailed is a specific error for when acquiring a distributed lock fails.
var ErrAcquireLockFailed = errors.New("failed to acquire distributed lock")

// DistributedLocker defines the interface for a distributed locking mechanism.
type DistributedLocker interface {
	AcquireLock(ctx context.Context, key string, expiry time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) (bool, error)
}

// OrderServiceClient defines the interface for interacting with the Order Service.
type OrderServiceClient interface {
	CreateOrderForFlashSale(ctx context.Context, userID, productID string, quantity int32, price float64) (string, error)
	CompensateCreateOrder(ctx context.Context, orderID string) error
}

// FlashSaleEvent represents a flash sale event in the business layer.
type FlashSaleEvent struct {
	ID          uint
	Name        string
	Description string
	StartTime   time.Time
	EndTime     time.Time
	Status      string // e.g., UPCOMING, ACTIVE, ENDED
	Products    []*FlashSaleProduct
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FlashSaleProduct represents a product within a flash sale event in the business layer.
type FlashSaleProduct struct {
	ID              uint
	EventID         uint
	ProductID       string
	FlashPrice      float64
	TotalStock      int32
	RemainingStock  int32
	MaxPerUser      int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// FlashSaleRepo defines the data storage interface for flash sale data.
// The business layer depends on this interface, not on a concrete data implementation.
type FlashSaleRepo interface {
	CreateFlashSaleEvent(ctx context.Context, event *FlashSaleEvent) (*FlashSaleEvent, error)
	GetFlashSaleEvent(ctx context.Context, id uint) (*FlashSaleEvent, error)
	ListActiveFlashSaleEvents(ctx context.Context) ([]*FlashSaleEvent, int32, error)
	GetFlashSaleProduct(ctx context.Context, eventID, productID string) (*FlashSaleProduct, error)
	UpdateFlashSaleProductStock(ctx context.Context, product *FlashSaleProduct, quantity int32) error
	CompensateFlashSaleProductStock(ctx context.Context, productID uint, quantity int32) error // For Saga compensation
	// TODO: Add methods for user purchase history in flash sale to check max_per_user
}

// FlashSaleUsecase is the use case for flash sale operations.
// It orchestrates the business logic.
type FlashSaleUsecase struct {
	repo        FlashSaleRepo
	locker      DistributedLocker
	orderClient OrderServiceClient
	// You can also inject other dependencies like a logger, payment service client
}

// NewFlashSaleUsecase creates a new FlashSaleUsecase.
func NewFlashSaleUsecase(repo FlashSaleRepo, locker DistributedLocker, orderClient OrderServiceClient) *FlashSaleUsecase {
	return &FlashSaleUsecase{repo: repo, locker: locker, orderClient: orderClient}
}

// CreateFlashSaleEvent creates a new flash sale event.
func (uc *FlashSaleUsecase) CreateFlashSaleEvent(ctx context.Context, name, description string, startTime, endTime time.Time, products []*FlashSaleProduct) (*FlashSaleEvent, error) {
	event := &FlashSaleEvent{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      "UPCOMING", // Initial status
		Products:    products,
	}
	return uc.repo.CreateFlashSaleEvent(ctx, event)
}

// GetFlashSaleEvent retrieves a flash sale event by ID.
func (uc *FlashSaleUsecase) GetFlashSaleEvent(ctx context.Context, id uint) (*FlashSaleEvent, error) {
	return uc.repo.GetFlashSaleEvent(ctx, id)
}

// ListActiveFlashSaleEvents lists all active flash sale events.
func (uc *FlashSaleUsecase) ListActiveFlashSaleEvents(ctx context.Context) ([]*FlashSaleEvent, int32, error) {
	return uc.repo.ListActiveFlashSaleEvents(ctx)
}

// ParticipateInFlashSale handles a user's attempt to purchase a flash sale product.
func (uc *FlashSaleUsecase) ParticipateInFlashSale(ctx context.Context, eventID uint, productID, userID string, quantity int32) (string, string, error) {
	lockKey := fmt.Sprintf("flashsale:lock:%d:%s", eventID, productID)
	// Acquire distributed lock
	locked, err := uc.locker.AcquireLock(ctx, lockKey, 5*time.Second) // 5 seconds expiry
	if err != nil {
		return "", "", fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !locked {
		return "", "", ErrAcquireLockFailed // Could not acquire lock, likely high contention
	}
	defer func() {
		// Release distributed lock
		_, releaseErr := uc.locker.ReleaseLock(ctx, lockKey)
		if releaseErr != nil {
			// Log the error, but don't fail the main transaction
			fmt.Printf("释放分布式锁失败: %v\n", releaseErr)
		}
	}()

	event, err := uc.repo.GetFlashSaleEvent(ctx, eventID)
	if err != nil {
		return "", "", err
	}

	// Check if flash sale is active
	now := time.Now()
	if now.Before(event.StartTime) || now.After(event.EndTime) {
		return "", "", ErrFlashSaleNotInProgress
	}

	// Find the product in the event
	var fsProduct *FlashSaleProduct
	for _, p := range event.Products {
		if p.ProductID == productID {
			fsProduct = p
			break
		}
	}

	if fsProduct == nil {
		return "", "", ErrFlashSaleProductNotFound
	}

	// Check stock
	if fsProduct.RemainingStock < quantity {
		return "", "", ErrFlashSaleOutOfStock
	}

	// Check max per user (requires tracking user purchases for this event/product)
	// For simplicity, this is skipped for now but would be crucial.
	if fsProduct.MaxPerUser > 0 && quantity > fsProduct.MaxPerUser {
		return "", "", ErrFlashSaleMaxPerUserExceeded
	}

	// --- Saga Step 1: Deduct stock (Flash Sale Service - Local Transaction) ---
	if err := uc.repo.UpdateFlashSaleProductStock(ctx, fsProduct, quantity); err != nil {
		return "", "", fmt.Errorf("扣减秒杀库存失败: %w", err)
	}

	// --- Saga Step 2: Create Order (Order Service - Remote Call) ---
	orderID, err := uc.orderClient.CreateOrderForFlashSale(ctx, userID, productID, quantity, fsProduct.FlashPrice)
	if err != nil {
		// Compensation for Step 1
		_ = uc.repo.CompensateFlashSaleProductStock(ctx, fsProduct.ID, quantity) // Log error if compensation fails
		return "", "", fmt.Errorf("创建订单失败: %w", err)
	}

	// TODO: Saga Step 3: Deduct user points/balance (Payment Service - Remote Call)
	// If this step fails, need to compensate Step 1 and Step 2.

	return orderID, "SUCCESS", nil
}

// GetFlashSaleProductDetails retrieves details for a specific product within a flash sale event.
func (uc *FlashSaleUsecase) GetFlashSaleProductDetails(ctx context.Context, eventID uint, productID string) (*FlashSaleProduct, error) {
	return uc.repo.GetFlashSaleProduct(ctx, eventID, productID)
}
