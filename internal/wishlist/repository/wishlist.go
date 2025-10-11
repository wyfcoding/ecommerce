package data

import (
	"context"
	"ecommerce/internal/wishlist/biz"
	"time"

	"gorm.io/gorm"
)

// wishlistRepo is the data layer implementation for WishlistRepo.
type wishlistRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.WishlistItem model to a biz.WishlistItem entity.
func (w *WishlistItem) toBiz() *biz.WishlistItem {
	if w == nil {
		return nil
	}
	return &biz.WishlistItem{
		ID:        w.ID,
		UserID:    w.UserID,
		ProductID: w.ProductID,
		AddedAt:   w.AddedAt,
	}
}

// fromBiz converts a biz.WishlistItem entity to a data.WishlistItem model.
func fromBiz(b *biz.WishlistItem) *WishlistItem {
	if b == nil {
		return nil
	}
	return &WishlistItem{
		UserID:    b.UserID,
		ProductID: b.ProductID,
		AddedAt:   b.AddedAt,
	}
}

// AddItem adds an item to a user's wishlist.
func (r *wishlistRepo) AddItem(ctx context.Context, userID, productID string) (*biz.WishlistItem, error) {
	item := &WishlistItem{
		UserID:    userID,
		ProductID: productID,
		AddedAt:   time.Now(),
	}

	// Check if item already exists in wishlist
	var existingItem WishlistItem
	if err := r.data.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&existingItem).Error; err == nil {
		// Item already exists, return it
		return existingItem.toBiz(), nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Item does not exist, create it
	if err := r.data.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}

	return item.toBiz(), nil
}

// RemoveItem removes an item from a user's wishlist.
func (r *wishlistRepo) RemoveItem(ctx context.Context, userID, productID string) error {
	result := r.data.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).Delete(&WishlistItem{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrWishlistItemNotFound
	}
	return nil
}

// ListItems lists all items in a user's wishlist.
func (r *wishlistRepo) ListItems(ctx context.Context, userID string, pageSize, pageToken int32) ([]*biz.WishlistItem, int32, error) {
	var items []WishlistItem
	var totalCount int32

	query := r.data.db.WithContext(ctx).Where("user_id = ?", userID)

	// Get total count
	query.Model(&WishlistItem{}).Count(int64(&totalCount))

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(int(pageSize)).Offset(int(pageToken * pageSize))
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, 0, err
	}

	bizItems := make([]*biz.WishlistItem, len(items))
	for i, item := range items {
		bizItems[i] = item.toBiz()
	}

	return bizItems, totalCount, nil
}
