package repository

import (
	"context"

	"ecommerce/internal/wishlist/model"
)

// WishlistRepo defines the data storage interface for wishlist items.
// The business layer depends on this interface, not on a concrete data implementation.
type WishlistRepo interface {
	AddItem(ctx context.Context, userID, productID string) (*model.WishlistItem, error)
	RemoveItem(ctx context.Context, userID, productID string) error
	ListItems(ctx context.Context, userID string, pageSize, pageToken int32) ([]*model.WishlistItem, int32, error)
}