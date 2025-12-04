// Package errors 定义了项目中使用的标准业务错误码、自定义错误类型以及错误处理辅助函数。
// 旨在提供统一的错误处理机制，便于错误识别、追踪和转换为gRPC状态码。
package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 业务错误码定义
// 错误码按照业务模块进行划分，便于分类和识别。
const (
	// 通用错误码 1xxx: 适用于大多数服务和场景的基础错误。
	ErrCodeInternal          = 1000 // 内部服务器错误
	ErrCodeInvalidArgument   = 1001 // 无效的参数
	ErrCodeNotFound          = 1002 // 资源未找到
	ErrCodeAlreadyExists     = 1003 // 资源已存在
	ErrCodePermissionDenied  = 1004 // 权限不足
	ErrCodeUnauthenticated   = 1005 // 未认证（例如，未登录或Token无效）
	ErrCodeResourceExhausted = 1006 // 资源耗尽（例如，请求速率限制）
	ErrCodeDeadlineExceeded  = 1007 // 操作超时

	// 用户相关错误码 2xxx: 针对用户模块特有的错误。
	ErrCodeUserNotFound      = 2001 // 用户未找到
	ErrCodeUserAlreadyExists = 2002 // 用户已存在
	ErrCodeInvalidPassword   = 2003 // 密码无效
	ErrCodeInvalidToken      = 2004 // Token无效或过期
	ErrCodeUserDisabled      = 2005 // 用户已被禁用
	ErrCodeUserDeleted       = 2006 // 用户已被删除

	// 商品相关错误码 3xxx: 针对商品模块特有的错误。
	ErrCodeProductNotFound      = 3001 // 商品未找到
	ErrCodeProductOutOfStock    = 3002 // 商品库存不足
	ErrCodeInvalidProductStatus = 3003 // 商品状态无效
	ErrCodeSKUNotFound          = 3004 // SKU未找到
	ErrCodeCategoryNotFound     = 3005 // 分类未找到
	ErrCodeBrandNotFound        = 3006 // 品牌未找到

	// 订单相关错误码 4xxx: 针对订单模块特有的错误。
	ErrCodeOrderNotFound      = 4001 // 订单未找到
	ErrCodeInvalidOrderStatus = 4002 // 订单状态无效
	ErrCodeOrderCannotCancel  = 4003 // 订单无法取消
	ErrCodeOrderCannotRefund  = 4004 // 订单无法退款
	ErrCodePaymentFailed      = 4005 // 支付失败
	ErrCodeOrderExpired       = 4006 // 订单已过期
	ErrCodeInvalidOrderAmount = 4007 // 订单金额无效

	// 购物车相关错误码 5xxx: 针对购物车模块特有的错误。
	ErrCodeCartItemNotFound = 5001 // 购物车商品未找到
	ErrCodeCartEmpty        = 5002 // 购物车为空
	ErrCodeCartItemExists   = 5003 // 购物车商品已存在
	ErrCodeInvalidQuantity  = 5004 // 数量无效

	// 库存相关错误码 6xxx: 针对库存模块特有的错误。
	ErrCodeStockNotFound      = 6001 // 库存记录未找到
	ErrCodeInsufficientStock  = 6002 // 库存不足
	ErrCodeStockLockFailed    = 6003 // 库存锁定失败
	ErrCodeStockReleaseFailed = 6004 // 库存释放失败

	// 支付相关错误码 7xxx: 针对支付模块特有的错误。
	ErrCodePaymentNotFound      = 7001 // 支付记录未找到
	ErrCodePaymentExpired       = 7002 // 支付已过期
	ErrCodePaymentProcessing    = 7003 // 支付处理中
	ErrCodeRefundFailed         = 7004 // 退款失败
	ErrCodeInvalidPaymentMethod = 7005 // 无效的支付方式

	// 优惠券相关错误码 8xxx: 针对优惠券模块特有的错误。
	ErrCodeCouponNotFound      = 8001 // 优惠券未找到
	ErrCodeCouponExpired       = 8002 // 优惠券已过期
	ErrCodeCouponUsed          = 8003 // 优惠券已使用
	ErrCodeCouponNotApplicable = 8004 // 优惠券不适用
	ErrCodeCouponExhausted     = 8005 // 优惠券已发完

	// 地址相关错误码 9xxx: 针对用户地址模块特有的错误。
	ErrCodeAddressNotFound = 9001 // 地址未找到
	ErrCodeInvalidAddress  = 9002 // 地址无效
)

// BizError 是自定义的业务错误类型，包含错误码、用户友好的消息和详细信息。
// 它实现了Go的 `error` 接口，可以作为普通错误进行处理。
type BizError struct {
	Code    int    // 业务错误码，用于程序识别和分类错误
	Message string // 用户友好的错误消息，可直接展示给用户
	Details string // 错误的详细信息，用于调试和日志记录
}

// Error 方法实现了Go的 `error` 接口，返回错误的字符串表示。
func (e *BizError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New 创建一个新的 `BizError` 实例。
// code: 业务错误码。
// message: 用户友好的错误消息。
func New(code int, message string) *BizError {
	return &BizError{
		Code:    code,
		Message: message,
	}
}

// Newf 创建一个带格式化消息的 `BizError` 实例。
// code: 业务错误码。
// format: 消息格式字符串。
// args: 格式化参数。
func Newf(code int, format string, args ...interface{}) *BizError {
	return &BizError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// WithDetails 为现有 `BizError` 添加详细信息。
// details: 错误的详细描述。
// 返回带有详细信息的新 `BizError` 实例（或修改原实例并返回）。
func (e *BizError) WithDetails(details string) *BizError {
	e.Details = details
	return e
}

// ToGRPCError 将 `BizError` 转换为gRPC的 `status.Error`。
// 这允许业务错误在gRPC服务间以标准化的方式进行传播和处理。
func (e *BizError) ToGRPCError() error {
	var code codes.Code
	// 根据业务错误码映射到gRPC标准错误码。
	// 这只是一个示例映射，实际应用中可能需要更全面的映射规则。
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
		code = codes.Internal // 默认映射为内部错误。
	}
	// 创建并返回gRPC状态错误。
	return status.Error(code, e.Error())
}

// 预定义常用错误实例，方便直接引用。
var (
	// 通用错误
	ErrInternal         = New(ErrCodeInternal, "internal server error")
	ErrInvalidArgument  = New(ErrCodeInvalidArgument, "invalid argument")
	ErrNotFound         = New(ErrCodeNotFound, "resource not found")
	ErrAlreadyExists    = New(ErrCodeAlreadyExists, "resource already exists")
	ErrPermissionDenied = New(ErrCodePermissionDenied, "permission denied")
	ErrUnauthenticated  = New(ErrCodeUnauthenticated, "unauthenticated")

	// 用户相关错误
	ErrUserNotFound      = New(ErrCodeUserNotFound, "user not found")
	ErrUserAlreadyExists = New(ErrCodeUserAlreadyExists, "user already exists")
	ErrInvalidPassword   = New(ErrCodeInvalidPassword, "invalid password")
	ErrInvalidToken      = New(ErrCodeInvalidToken, "invalid token")

	// 商品相关错误
	ErrProductNotFound      = New(ErrCodeProductNotFound, "product not found")
	ErrProductOutOfStock    = New(ErrCodeProductOutOfStock, "product out of stock")
	ErrInvalidProductStatus = New(ErrCodeInvalidProductStatus, "invalid product status")

	// 订单相关错误
	ErrOrderNotFound      = New(ErrCodeOrderNotFound, "order not found")
	ErrInvalidOrderStatus = New(ErrCodeInvalidOrderStatus, "invalid order status")
	ErrOrderCannotCancel  = New(ErrCodeOrderCannotCancel, "order cannot be cancelled")
	ErrOrderCannotRefund  = New(ErrCodeOrderCannotRefund, "order cannot be refunded")
	ErrPaymentFailed      = New(ErrCodePaymentFailed, "payment failed")

	// 购物车相关错误
	ErrCartItemNotFound = New(ErrCodeCartItemNotFound, "cart item not found")
	ErrCartEmpty        = New(ErrCodeCartEmpty, "cart is empty")
	ErrCartItemExists   = New(ErrCodeCartItemExists, "cart item already exists")
	ErrInvalidQuantity  = New(ErrCodeInvalidQuantity, "invalid quantity")

	// 库存相关错误
	ErrStockNotFound      = New(ErrCodeStockNotFound, "stock not found")
	ErrInsufficientStock  = New(ErrCodeInsufficientStock, "insufficient stock")
	ErrStockLockFailed    = New(ErrCodeStockLockFailed, "stock lock failed")
	ErrStockReleaseFailed = New(ErrCodeStockReleaseFailed, "stock release failed")

	// 支付相关错误
	ErrPaymentNotFound      = New(ErrCodePaymentNotFound, "payment not found")
	ErrPaymentExpired       = New(ErrCodePaymentExpired, "payment expired")
	ErrPaymentProcessing    = New(ErrCodePaymentProcessing, "payment is processing")
	ErrRefundFailed         = New(ErrCodeRefundFailed, "refund failed")
	ErrInvalidPaymentMethod = New(ErrCodeInvalidPaymentMethod, "invalid payment method")

	// 优惠券相关错误
	ErrCouponNotFound      = New(ErrCodeCouponNotFound, "coupon not found")
	ErrCouponExpired       = New(ErrCodeCouponExpired, "coupon expired")
	ErrCouponUsed          = New(ErrCodeCouponUsed, "coupon already used")
	ErrCouponNotApplicable = New(ErrCodeCouponNotApplicable, "coupon not applicable")
	ErrCouponExhausted     = New(ErrCodeCouponExhausted, "coupon exhausted")

	// 地址相关错误
	ErrAddressNotFound = New(ErrCodeAddressNotFound, "address not found")
	ErrInvalidAddress  = New(ErrCodeInvalidAddress, "invalid address")
)

// WrapError 使用提供的消息包装一个错误。
// 这允许在错误链中添加上下文信息。
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	// 使用 fmt.Errorf 的 %w 动词包装错误，以便后续可以使用 errors.Is 和 errors.As 进行检查。
	return fmt.Errorf("%s: %w", message, err)
}

// Cause 返回错误链的根本原因。
// 它会递归地解包错误，直到找到一个不再包装其他错误的原始错误。
func Cause(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

// Is 检查错误链中是否包含特定的目标错误。
// 类似于标准库 `errors.Is`。
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 检查错误链中是否包含与目标类型匹配的错误，并将其赋值给目标。
// 类似于标准库 `errors.As`。
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
