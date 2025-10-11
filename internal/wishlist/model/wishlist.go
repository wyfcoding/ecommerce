package biz

import (
	"context"
	"errors"
	"time"
)

// ErrWishlistItemNotFound is a specific error for when a wishlist item is not found.
var ErrWishlistItemNotFound = errors.New("wishlist item not found")

// WishlistItem represents an item in a user's wishlist in the business layer.
type WishlistItem struct {
	ID        uint
	UserID    string
	ProductID string
	AddedAt   time.Time
}

// WishlistRepo defines the data storage interface for wishlist items.
// The business layer depends on this interface, not on a concrete data implementation.
type WishlistRepo interface {
	AddItem(ctx context.Context, userID, productID string) (*WishlistItem, error)
	RemoveItem(ctx context.Context, userID, productID string) error
	ListItems(ctx context.Context, userID string, pageSize, pageToken int32) ([]*WishlistItem, int32, error)
}

// WishlistUsecase is the use case for wishlist-related operations.
// It orchestrates the business logic.
type WishlistUsecase struct {
	repo WishlistRepo
	// You can also inject other dependencies like a logger
}

// NewWishlistUsecase creates a new WishlistUsecase.
func NewWishlistUsecase(repo WishlistRepo) *WishlistUsecase {
	return &WishlistUsecase{repo: repo}
}

// AddItem adds an item to a user's wishlist.
func (uc *WishlistUsecase) AddItem(ctx context.Context, userID, productID string) (*WishlistItem, error) {
	// Here you can add business logic before adding, e.g., check if item already exists, product validity, etc.
	return uc.repo.AddItem(ctx, userID, productID)
}

// RemoveItem removes an item from a user's wishlist.
func (uc *WishlistUsecase) RemoveItem(ctx context.Context, userID, productID string) error {
	// Here you can add business logic before removing.
	return uc.repo.RemoveItem(ctx, userID, productID)
}

// ListItems lists all items in a user's wishlist.
func (uc *WishlistUsecase) ListItems(ctx context.Context, userID string, pageSize, pageToken int32) ([]*WishlistItem, int32, error) {
	return uc.repo.ListItems(ctx, userID, pageSize, pageToken)
}
