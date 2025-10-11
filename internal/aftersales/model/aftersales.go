package biz

import (
	"context"
	"errors"
	"time"
)

// ErrReturnRequestNotFound 是退货请求未找到时的特定错误。
var ErrReturnRequestNotFound = errors.New("return request not found")

// ErrRefundRequestNotFound 是退款请求未找到时的特定错误。
var ErrRefundRequestNotFound = errors.New("refund request not found")

// ReturnRequest 表示业务层中的退货请求。
type ReturnRequest struct {
	ID        uint
	OrderID   string
	UserID    string
	ProductID string
	Quantity  int32
	Reason    string
	Status    string // 例如：PENDING, APPROVED, REJECTED, RECEIVED, REFUNDED
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RefundRequest 表示业务层中的退款请求。
type RefundRequest struct {
	ID              uint
	ReturnRequestID string
	OrderID         string
	UserID          string
	Amount          float64
	Currency        string
	Status          string // 例如：PENDING, APPROVED, REJECTED, COMPLETED
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AftersalesRepo 定义了售后数据的数据存储接口。
// 业务层依赖于此接口，而不是具体的实现。
type AftersalesRepo interface {
	CreateReturnRequest(ctx context.Context, req *ReturnRequest) (*ReturnRequest, error)
	GetReturnRequest(ctx context.Context, id uint) (*ReturnRequest, error)
	UpdateReturnRequest(ctx context.Context, req *ReturnRequest) (*ReturnRequest, error)

	CreateRefundRequest(ctx context.Context, req *RefundRequest) (*RefundRequest, error)
	GetRefundRequest(ctx context.Context, id uint) (*RefundRequest, error)
	UpdateRefundRequest(ctx context.Context, req *RefundRequest) (*RefundRequest, error)
}

// AftersalesUsecase 是售后操作的用例。
// 它协调业务逻辑。
type AftersalesUsecase struct {
	repo AftersalesRepo
	// 您还可以注入其他依赖项，例如日志记录器或订单服务客户端
}

// NewAftersalesUsecase 创建一个新的 AftersalesUsecase。
func NewAftersalesUsecase(repo AftersalesRepo) *AftersalesUsecase {
	return &AftersalesUsecase{repo: repo}
}

// CreateReturnRequest 创建一个新的退货请求。
func (uc *AftersalesUsecase) CreateReturnRequest(ctx context.Context, orderID, userID, productID string, quantity int32, reason string) (*ReturnRequest, error) {
	// TODO: 添加业务逻辑，例如验证订单，检查产品是否符合退货条件
	req := &ReturnRequest{
		OrderID:   orderID,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Reason:    reason,
		Status:    "PENDING", // 初始状态
	}
	return uc.repo.CreateReturnRequest(ctx, req)
}

// GetReturnRequest 根据 ID 检索退货请求。
func (uc *AftersalesUsecase) GetReturnRequest(ctx context.Context, id uint) (*ReturnRequest, error) {
	return uc.repo.GetReturnRequest(ctx, id)
}

// UpdateReturnRequestStatus 更新退货请求的状态。
func (uc *AftersalesUsecase) UpdateReturnRequestStatus(ctx context.Context, id uint, status string) (*ReturnRequest, error) {
	req, err := uc.repo.GetReturnRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Status = status
	return uc.repo.UpdateReturnRequest(ctx, req)
}

// CreateRefundRequest 创建一个新的退款请求。
func (uc *AftersalesUsecase) CreateRefundRequest(ctx context.Context, returnRequestID, orderID, userID string, amount float64, currency string) (*RefundRequest, error) {
	// TODO: 添加业务逻辑，例如验证退货请求，检查支付状态
	req := &RefundRequest{
		ReturnRequestID: returnRequestID,
		OrderID:         orderID,
		UserID:          userID,
		Amount:          amount,
		Currency:        currency,
		Status:          "PENDING", // 初始状态
	}
	return uc.repo.CreateRefundRequest(ctx, req)
}

// GetRefundRequest 根据 ID 检索退款请求。
func (uc *AftersalesUsecase) GetRefundRequest(ctx context.Context, id uint) (*RefundRequest, error) {
	return uc.repo.GetRefundRequest(ctx, id)
}

// UpdateRefundRequestStatus 更新退款请求的状态。
func (uc *AftersalesUsecase) UpdateRefundRequestStatus(ctx context.Context, id uint, status string) (*RefundRequest, error) {
	req, err := uc.repo.GetRefundRequest(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Status = status
	return uc.repo.UpdateRefundRequest(ctx, req)
}
