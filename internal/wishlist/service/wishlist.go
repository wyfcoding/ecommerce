package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/wishlist/model"
	"ecommerce/internal/wishlist/repository"
)

// ErrWishlistItemNotFound is a specific error for when a wishlist item is not found.
var ErrWishlistItemNotFound = errors.New("wishlist item not found")

// WishlistService is the use case for wishlist-related operations.
// It orchestrates the business logic.
type WishlistService struct {
	repo repository.WishlistRepo
	// You can also inject other dependencies like a logger
}

// NewWishlistService creates a new WishlistService.
func NewWishlistService(repo repository.WishlistRepo) *WishlistService {
	return &WishlistService{repo: repo}
}

// AddItem adds an item to a user's wishlist.
func (s *WishlistService) AddItem(ctx context.Context, userID, productID string) (*model.WishlistItem, error) {
	// Here you can add business logic before adding, e.g., check if item already exists, product validity, etc.
	return s.repo.AddItem(ctx, userID, productID)
}

// RemoveItem removes an item from a user's wishlist.
func (s *WishlistService) RemoveItem(ctx context.Context, userID, productID string) error {
	// Here you can add business logic before removing.
	return s.repo.RemoveItem(ctx, userID, productID)
}

// ListItems lists all items in a user's wishlist.
func (s *WishlistService) ListItems(ctx context.Context, userID string, pageSize, pageToken int32) ([]*model.WishlistItem, int32, error) {
	return s.repo.ListItems(ctx, userID, pageSize, pageToken)
}
