package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrOrderNotCompleted      = errors.New("order not completed, cannot review")
	ErrAlreadyReviewed        = errors.New("order already reviewed")
	ErrReviewWindowExpired    = errors.New("review window has expired")
	ErrInvalidRating          = errors.New("rating must be between 1 and 5")
)

type ReviewStatus int8

const (
	ReviewStatusPending   ReviewStatus = 0
	ReviewStatusPublished ReviewStatus = 1
	ReviewStatusHidden    ReviewStatus = 2
	ReviewStatusDeleted   ReviewStatus = 3
)

type OrderReview struct {
	ID           uint64        `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	OrderID      uint64        `json:"order_id"`
	OrderNo      string        `json:"order_no"`
	UserID       uint64        `json:"user_id"`
	ProductID    uint64        `json:"product_id"`
	SkuID        uint64        `json:"sku_id"`
	Rating       int8          `json:"rating"`
	Content      string        `json:"content"`
	Images       []string      `json:"images"`
	VideoURL     string        `json:"video_url"`
	IsAnonymous  bool          `json:"is_anonymous"`
	Status       ReviewStatus  `json:"status"`
	Likes        int32         `json:"likes"`
	MerchantReply string       `json:"merchant_reply"`
	RepliedAt    *time.Time    `json:"replied_at"`
	Tags         []string      `json:"tags"`
}

type OrderReviewSummary struct {
	OrderID        uint64 `json:"order_id"`
	OrderNo        string `json:"order_no"`
	UserID         uint64 `json:"user_id"`
	TotalItems     int    `json:"total_items"`
	ReviewedItems  int    `json:"reviewed_items"`
	IsFullyReviewed bool  `json:"is_fully_reviewed"`
	ReviewDeadline *time.Time `json:"review_deadline"`
}

type ReviewConfig struct {
	ReviewWindowDays int
	MaxImages        int
	MaxContentLength int
	MinContentLength int
}

func DefaultReviewConfig() *ReviewConfig {
	return &ReviewConfig{
		ReviewWindowDays: 30,
		MaxImages:        9,
		MaxContentLength: 500,
		MinContentLength: 10,
	}
}

func NewOrderReview(orderID uint64, orderNo string, userID, productID, skuID uint64, rating int8, content string, images []string, isAnonymous bool) (*OrderReview, error) {
	if rating < 1 || rating > 5 {
		return nil, ErrInvalidRating
	}

	return &OrderReview{
		OrderID:     orderID,
		OrderNo:     orderNo,
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		Rating:      rating,
		Content:     content,
		Images:      images,
		IsAnonymous: isAnonymous,
		Status:      ReviewStatusPending,
		Tags:        []string{},
	}, nil
}

func (r *OrderReview) Publish() {
	r.Status = ReviewStatusPublished
}

func (r *OrderReview) Hide() {
	r.Status = ReviewStatusHidden
}

func (r *OrderReview) Delete() {
	r.Status = ReviewStatusDeleted
}

func (r *OrderReview) MerchantReplyTo(reply string) {
	r.MerchantReply = reply
	now := time.Now()
	r.RepliedAt = &now
}

func (r *OrderReview) AddLike() {
	r.Likes++
}

func (r *OrderReview) IsPublished() bool {
	return r.Status == ReviewStatusPublished
}

type OrderReviewAssociation struct {
	OrderID         uint64            `json:"order_id"`
	OrderNo         string            `json:"order_no"`
	UserID          uint64            `json:"user_id"`
	CompletedAt     time.Time         `json:"completed_at"`
	ReviewDeadline  time.Time         `json:"review_deadline"`
	ItemReviews     []*ItemReviewInfo `json:"item_reviews"`
	AllItemsReviewed bool              `json:"all_items_reviewed"`
}

type ItemReviewInfo struct {
	ProductID   uint64  `json:"product_id"`
	SkuID       uint64  `json:"sku_id"`
	ProductName string  `json:"product_name"`
	SkuName     string  `json:"sku_name"`
	Quantity    int32   `json:"quantity"`
	Reviewed    bool    `json:"reviewed"`
	ReviewID    uint64  `json:"review_id"`
	Rating      int8    `json:"rating"`
}

func NewOrderReviewAssociation(orderID uint64, orderNo string, userID uint64, completedAt time.Time, reviewWindowDays int) *OrderReviewAssociation {
	return &OrderReviewAssociation{
		OrderID:        orderID,
		OrderNo:        orderNo,
		UserID:         userID,
		CompletedAt:    completedAt,
		ReviewDeadline: completedAt.AddDate(0, 0, reviewWindowDays),
		ItemReviews:    []*ItemReviewInfo{},
	}
}

func (a *OrderReviewAssociation) AddItem(productID, skuID uint64, productName, skuName string, quantity int32) {
	a.ItemReviews = append(a.ItemReviews, &ItemReviewInfo{
		ProductID:   productID,
		SkuID:       skuID,
		ProductName: productName,
		SkuName:     skuName,
		Quantity:    quantity,
		Reviewed:    false,
	})
}

func (a *OrderReviewAssociation) MarkItemReviewed(skuID uint64, reviewID uint64, rating int8) {
	for _, item := range a.ItemReviews {
		if item.SkuID == skuID {
			item.Reviewed = true
			item.ReviewID = reviewID
			item.Rating = rating
			break
		}
	}
	a.checkAllReviewed()
}

func (a *OrderReviewAssociation) checkAllReviewed() {
	a.AllItemsReviewed = true
	for _, item := range a.ItemReviews {
		if !item.Reviewed {
			a.AllItemsReviewed = false
			break
		}
	}
}

func (a *OrderReviewAssociation) CanReview() bool {
	return time.Now().Before(a.ReviewDeadline)
}

func (a *OrderReviewAssociation) GetUnreviewedItems() []*ItemReviewInfo {
	var items []*ItemReviewInfo
	for _, item := range a.ItemReviews {
		if !item.Reviewed {
			items = append(items, item)
		}
	}
	return items
}

func (a *OrderReviewAssociation) GetReviewedItems() []*ItemReviewInfo {
	var items []*ItemReviewInfo
	for _, item := range a.ItemReviews {
		if item.Reviewed {
			items = append(items, item)
		}
	}
	return items
}

type OrderReviewRepository interface {
	Save(ctx context.Context, review *OrderReview) error
	FindByID(ctx context.Context, id uint64) (*OrderReview, error)
	FindByOrderID(ctx context.Context, orderID uint64) ([]*OrderReview, error)
	FindByOrderIDAndSkuID(ctx context.Context, orderID, skuID uint64) (*OrderReview, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*OrderReview, error)
	FindByProductID(ctx context.Context, productID uint64, limit, offset int) ([]*OrderReview, error)
	Update(ctx context.Context, review *OrderReview) error
	Delete(ctx context.Context, id uint64) error
}

type OrderReviewAssociationRepository interface {
	Save(ctx context.Context, association *OrderReviewAssociation) error
	FindByOrderID(ctx context.Context, orderID uint64) (*OrderReviewAssociation, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*OrderReviewAssociation, error)
	Update(ctx context.Context, association *OrderReviewAssociation) error
}
