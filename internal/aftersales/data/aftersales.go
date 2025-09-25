package data

import (
	"context"
	"ecommerce/internal/aftersales/biz"

	"gorm.io/gorm"
)

// aftersalesRepo is the data layer implementation for AftersalesRepo.
type aftersalesRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.ReturnRequest model to a biz.ReturnRequest entity.
func (r *ReturnRequest) toBiz() *biz.ReturnRequest {
	if r == nil {
		return nil
	}
	return &biz.ReturnRequest{
		ID:        r.ID,
		OrderID:   r.OrderID,
		UserID:    r.UserID,
		ProductID: r.ProductID,
		Quantity:  r.Quantity,
		Reason:    r.Reason,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// fromBiz converts a biz.ReturnRequest entity to a data.ReturnRequest model.
func fromBizReturnRequest(b *biz.ReturnRequest) *ReturnRequest {
	if b == nil {
		return nil
	}
	return &ReturnRequest{
		OrderID:   b.OrderID,
		UserID:    b.UserID,
		ProductID: b.ProductID,
		Quantity:  b.Quantity,
		Reason:    b.Reason,
		Status:    b.Status,
	}
}

// toBiz converts a data.RefundRequest model to a biz.RefundRequest entity.
func (r *RefundRequest) toBiz() *biz.RefundRequest {
	if r == nil {
		return nil
	}
	return &biz.RefundRequest{
		ID:              r.ID,
		ReturnRequestID: r.ReturnRequestID,
		OrderID:         r.OrderID,
		UserID:          r.UserID,
		Amount:          r.Amount,
		Currency:        r.Currency,
		Status:          r.Status,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// fromBiz converts a biz.RefundRequest entity to a data.RefundRequest model.
func fromBizRefundRequest(b *biz.RefundRequest) *RefundRequest {
	if b == nil {
		return nil
	}
	return &RefundRequest{
		ReturnRequestID: b.ReturnRequestID,
		OrderID:         b.OrderID,
		UserID:          b.UserID,
		Amount:          b.Amount,
		Currency:        b.Currency,
		Status:          b.Status,
	}
}

// CreateReturnRequest creates a new return request in the database.
func (r *aftersalesRepo) CreateReturnRequest(ctx context.Context, b *biz.ReturnRequest) (*biz.ReturnRequest, error) {
	req := fromBizReturnRequest(b)
	if err := r.data.db.WithContext(ctx).Create(req).Error; err != nil {
		return nil, err
	}
	return req.toBiz(), nil
}

// GetReturnRequest retrieves a return request by ID from the database.
func (r *aftersalesRepo) GetReturnRequest(ctx context.Context, id uint) (*biz.ReturnRequest, error) {
	var req ReturnRequest
	if err := r.data.db.WithContext(ctx).First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrReturnRequestNotFound
		}
		return nil, err
	}
	return req.toBiz(), nil
}

// UpdateReturnRequest updates an existing return request in the database.
func (r *aftersalesRepo) UpdateReturnRequest(ctx context.Context, b *biz.ReturnRequest) (*biz.ReturnRequest, error) {
	req := fromBizReturnRequest(b)
	req.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(req).Error; err != nil {
		return nil, err
	}
	return req.toBiz(), nil
}

// CreateRefundRequest creates a new refund request in the database.
func (r *aftersalesRepo) CreateRefundRequest(ctx context.Context, b *biz.RefundRequest) (*biz.RefundRequest, error) {
	req := fromBizRefundRequest(b)
	if err := r.data.db.WithContext(ctx).Create(req).Error; err != nil {
		return nil, err
	}
	return req.toBiz(), nil
}

// GetRefundRequest retrieves a refund request by ID from the database.
func (r *aftersalesRepo) GetRefundRequest(ctx context.Context, id uint) (*biz.RefundRequest, error) {
	var req RefundRequest
	if err := r.data.db.WithContext(ctx).First(&req, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrRefundRequestNotFound
		}
		return nil, err
	}
	return req.toBiz(), nil
}

// UpdateRefundRequest updates an existing refund request in the database.
func (r *aftersalesRepo) UpdateRefundRequest(ctx context.Context, b *biz.RefundRequest) (*biz.RefundRequest, error) {
	req := fromBizRefundRequest(b)
	req.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(req).Error; err != nil {
		return nil, err
	}
	return req.toBiz(), nil
}
