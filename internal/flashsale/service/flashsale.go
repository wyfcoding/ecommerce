package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/flashsale/client"
	"ecommerce/internal/flashsale/model"
	"ecommerce/internal/flashsale/repository"
)

var (
	ErrFlashSaleEventNotFound     = errors.New("flash sale event not found")
	ErrFlashSaleProductNotFound   = errors.New("flash sale product not found")
	ErrFlashSaleNotInProgress     = errors.New("flash sale not in progress")
	ErrFlashSaleOutOfStock        = errors.New("flash sale product out of stock")
	ErrFlashSaleMaxPerUserExceeded = errors.New("max quantity per user exceeded")
	ErrAcquireLockFailed          = errors.New("failed to acquire distributed lock")
)

// FlashSaleService is the use case for flash sale operations.
// It orchestrates the business logic.
type FlashSaleService struct {
	repo        repository.FlashSaleRepo
	locker      repository.DistributedLocker
	orderClient client.OrderServiceClient
	// You can also inject other dependencies like a logger, payment service client
}

// NewFlashSaleService creates a new FlashSaleService.
func NewFlashSaleService(repo repository.FlashSaleRepo, locker repository.DistributedLocker, orderClient client.OrderServiceClient) *FlashSaleService {
	return &FlashSaleService{repo: repo, locker: locker, orderClient: orderClient}
}

// CreateFlashSaleEvent creates a new flash sale event.
func (s *FlashSaleService) CreateFlashSaleEvent(ctx context.Context, name, description string, startTime, endTime time.Time, products []*model.FlashSaleProduct) (*model.FlashSaleEvent, error) {
	event := &model.FlashSaleEvent{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      "UPCOMING", // Initial status
		Products:    products,
	}
	return s.repo.CreateFlashSaleEvent(ctx, event)
}

// GetFlashSaleEvent retrieves a flash sale event by ID.
func (s *FlashSaleService) GetFlashSaleEvent(ctx context.Context, id uint) (*model.FlashSaleEvent, error) {
	return s.repo.GetFlashSaleEvent(ctx, id)
}

// ListActiveFlashSaleEvents lists all active flash sale events.
func (s *FlashSaleService) ListActiveFlashSaleEvents(ctx context.Context) ([]*model.FlashSaleEvent, int32, error) {
	return s.repo.ListActiveFlashSaleEvents(ctx)
}

// ParticipateInFlashSale handles a user's attempt to purchase a flash sale product.
func (s *FlashSaleService) ParticipateInFlashSale(ctx context.Context, eventID uint, productID, userID string, quantity int32) (string, string, error) {
	lockKey := fmt.Sprintf("flashsale:lock:%d:%s", eventID, productID)
	// Acquire distributed lock
	locked, err := s.locker.AcquireLock(ctx, lockKey, 5*time.Second) // 5 seconds expiry
	if err != nil {
		return "", "", fmt.Errorf("获取分布式锁失败: %w", err)
	}
	if !locked {
		return "", "", ErrAcquireLockFailed // Could not acquire lock, likely high contention
	}
	defer func() {
		// Release distributed lock
		_, releaseErr := s.locker.ReleaseLock(ctx, lockKey)
		if releaseErr != nil {
			// Log the error, but don't fail the main transaction
			fmt.Printf("释放分布式锁失败: %v\n", releaseErr)
		}
	}()

	event, err := s.repo.GetFlashSaleEvent(ctx, eventID)
	if err != nil {
		return "", "", err
	}

	// Check if flash sale is active
	now := time.Now()
	if now.Before(event.StartTime) || now.After(event.EndTime) {
		return "", "", ErrFlashSaleNotInProgress
	}

	// Find the product in the event
	var fsProduct *model.FlashSaleProduct
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
	if err := s.repo.UpdateFlashSaleProductStock(ctx, fsProduct, quantity); err != nil {
		return "", "", fmt.Errorf("扣减秒杀库存失败: %w", err)
	}

	// --- Saga Step 2: Create Order (Order Service - Remote Call) ---
	orderID, err := s.orderClient.CreateOrderForFlashSale(ctx, userID, productID, quantity, fsProduct.FlashPrice)
	if err != nil {
		// Compensation for Step 1
		_ = s.repo.CompensateFlashSaleProductStock(ctx, fsProduct.ID, quantity) // Log error if compensation fails
		return "", "", fmt.Errorf("创建订单失败: %w", err)
	}

	// TODO: Saga Step 3: Deduct user points/balance (Payment Service - Remote Call)
	// If this step fails, need to compensate Step 1 and Step 2.

	return orderID, "SUCCESS", nil
}

// GetFlashSaleProductDetails retrieves details for a specific product within a flash sale event.
func (s *FlashSaleService) GetFlashSaleProductDetails(ctx context.Context, eventID uint, productID string) (*model.FlashSaleProduct, error) {
	return s.repo.GetFlashSaleProduct(ctx, eventID, productID)
}
