package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 业务错误码定义
const (
	// 通用错误码 1xxx
	ErrCodeInternal          = 1000
	ErrCodeInvalidArgument   = 1001
	ErrCodeNotFound          = 1002
	ErrCodeAlreadyExists     = 1003
	ErrCodePermissionDenied  = 1004
	ErrCodeUnauthenticated   = 1005
	ErrCodeResourceExhausted = 1006
	ErrCodeDeadlineExceeded  = 1007

	// 用户相关错误码 2xxx
	ErrCodeUserNotFound      = 2001
	ErrCodeUserAlreadyExists = 2002
	ErrCodeInvalidPassword   = 2003
	ErrCodeInvalidToken      = 2004
	ErrCodeUserDisabled      = 2005
	ErrCodeUserDeleted       = 2006

	// 商品相关错误码 3xxx
	ErrCodeProductNotFound      = 3001
	ErrCodeProductOutOfStock    = 3002
	ErrCodeInvalidProductStatus = 3003
	ErrCodeSKUNotFound          = 3004
	ErrCodeCategoryNotFound     = 3005
	ErrCodeBrandNotFound        = 3006

	// 订单相关错误码 4xxx
	ErrCodeOrderNotFound      = 4001
	ErrCodeInvalidOrderStatus = 4002
	ErrCodeOrderCannotCancel  = 4003
	ErrCodeOrderCannotRefund  = 4004
	ErrCodePaymentFailed      = 4005
	ErrCodeOrderExpired       = 4006
	ErrCodeInvalidOrderAmount = 4007

	// 购物车相关错误码 5xxx
	ErrCodeCartItemNotFound = 5001
	ErrCodeCartEmpty        = 5002
	ErrCodeCartItemExists   = 5003
	ErrCodeInvalidQuantity  = 5004

	// 库存相关错误码 6xxx
	ErrCodeStockNotFound      = 6001
	ErrCodeInsufficientStock  = 6002
	ErrCodeStockLockFailed    = 6003
	ErrCodeStockReleaseFailed = 6004

	// 支付相关错误码 7xxx
	ErrCodePaymentNotFound      = 7001
	ErrCodePaymentExpired       = 7002
	ErrCodePaymentProcessing    = 7003
	ErrCodeRefundFailed         = 7004
	ErrCodeInvalidPaymentMethod = 7005

	// 优惠券相关错误码 8xxx
	ErrCodeCouponNotFound      = 8001
	ErrCodeCouponExpired       = 8002
	ErrCodeCouponUsed          = 8003
	ErrCodeCouponNotApplicable = 8004
	ErrCodeCouponExhausted     = 8005

	// 地址相关错误码 9xxx
	ErrCodeAddressNotFound = 9001
	ErrCodeInvalidAddress  = 9002
)

// BizError 业务错误
type BizError struct {
	Code    int
	Message string
	Details string
}

// Error 实现error接口
func (e *BizError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建一个新的业务错误
func New(code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建一个带格式化消息的业务错误
func Newf(code int, format string, args ...interface{}) *BizError {
	return &BizError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithDetails 添加错误详情
func (e *BizError) WithDetails(details string) *BizError {
	e.Details = details
	return e
}

// ToGRPCError 转换为gRPC错误
func (e *BizError) ToGRPCError() error {
	var code codes.Code
	switch e.Code {
	case ErrCodeInvalidArgument:
		code = codes.InvalidArgument
	case ErrCodeNotFound, ErrCodeUserNotFound, ErrCodeProductNotFound, ErrCodeOrderNotFound, ErrCodeCartItemNotFound:
		code = codes.NotFound
	case ErrCodeAlreadyExists, ErrCodeUserAlreadyExists:
		code = codes.AlreadyExists
	case ErrCodePermissionDenied:
		code = codes.PermissionDenied
	case ErrCodeUnauthenticated, ErrCodeInvalidPassword, ErrCodeInvalidToken:
		code = codes.Unauthenticated
	default:
		code = codes.Internal
	}
	return status.Error(code, e.Error())
}

// 预定义的常用错误
var (
	ErrInternal         = New(ErrCodeInternal, "internal server error")
	ErrInvalidArgument  = New(ErrCodeInvalidArgument, "invalid argument")
	ErrNotFound         = New(ErrCodeNotFound, "resource not found")
	ErrAlreadyExists    = New(ErrCodeAlreadyExists, "resource already exists")
	ErrPermissionDenied = New(ErrCodePermissionDenied, "permission denied")
	ErrUnauthenticated  = New(ErrCodeUnauthenticated, "unauthenticated")

	// 用户相关
	ErrUserNotFound      = New(ErrCodeUserNotFound, "user not found")
	ErrUserAlreadyExists = New(ErrCodeUserAlreadyExists, "user already exists")
	ErrInvalidPassword   = New(ErrCodeInvalidPassword, "invalid password")
	ErrInvalidToken      = New(ErrCodeInvalidToken, "invalid token")

	// 商品相关
	ErrProductNotFound      = New(ErrCodeProductNotFound, "product not found")
	ErrProductOutOfStock    = New(ErrCodeProductOutOfStock, "product out of stock")
	ErrInvalidProductStatus = New(ErrCodeInvalidProductStatus, "invalid product status")

	// 订单相关
	ErrOrderNotFound      = New(ErrCodeOrderNotFound, "order not found")
	ErrInvalidOrderStatus = New(ErrCodeInvalidOrderStatus, "invalid order status")
	ErrOrderCannotCancel  = New(ErrCodeOrderCannotCancel, "order cannot be cancelled")
	ErrOrderCannotRefund  = New(ErrCodeOrderCannotRefund, "order cannot be refunded")
	ErrPaymentFailed      = New(ErrCodePaymentFailed, "payment failed")

	// 购物车相关
	ErrCartItemNotFound = New(ErrCodeCartItemNotFound, "cart item not found")
	ErrCartEmpty        = New(ErrCodeCartEmpty, "cart is empty")
	ErrCartItemExists   = New(ErrCodeCartItemExists, "cart item already exists")
	ErrInvalidQuantity  = New(ErrCodeInvalidQuantity, "invalid quantity")

	// 库存相关
	ErrStockNotFound      = New(ErrCodeStockNotFound, "stock not found")
	ErrInsufficientStock  = New(ErrCodeInsufficientStock, "insufficient stock")
	ErrStockLockFailed    = New(ErrCodeStockLockFailed, "stock lock failed")
	ErrStockReleaseFailed = New(ErrCodeStockReleaseFailed, "stock release failed")

	// 支付相关
	ErrPaymentNotFound      = New(ErrCodePaymentNotFound, "payment not found")
	ErrPaymentExpired       = New(ErrCodePaymentExpired, "payment expired")
	ErrPaymentProcessing    = New(ErrCodePaymentProcessing, "payment is processing")
	ErrRefundFailed         = New(ErrCodeRefundFailed, "refund failed")
	ErrInvalidPaymentMethod = New(ErrCodeInvalidPaymentMethod, "invalid payment method")

	// 优惠券相关
	ErrCouponNotFound      = New(ErrCodeCouponNotFound, "coupon not found")
	ErrCouponExpired       = New(ErrCodeCouponExpired, "coupon expired")
	ErrCouponUsed          = New(ErrCodeCouponUsed, "coupon already used")
	ErrCouponNotApplicable = New(ErrCodeCouponNotApplicable, "coupon not applicable")
	ErrCouponExhausted     = New(ErrCodeCouponExhausted, "coupon exhausted")

	// 地址相关
	ErrAddressNotFound = New(ErrCodeAddressNotFound, "address not found")
	ErrInvalidAddress  = New(ErrCodeInvalidAddress, "invalid address")
)
