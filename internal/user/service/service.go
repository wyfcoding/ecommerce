// package service 实现了 API 层定义的 gRPC 接口。
// 它的主要职责是：
// 1. 接收来自 gRPC 的请求（Request）。
// 2. 对请求进行基本的校验和转换，将 gRPC 的数据模型（Proto a.k.a. DTO）转换为业务逻辑层（biz）的领域模型。
// 3. 调用业务逻辑层（biz）的方法来执行核心业务。
// 4. 将业务逻辑层的返回结果（领域模型或错误）转换为 gRPC 的响应（Response）或 gRPC 状态错误码。
package service

import (
	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/biz"
)

// UserService 是 gRPC 服务的实现。
type UserService struct {
	v1.UnimplementedUserServer

	userUsecase    *biz.UserUsecase
	addressUsecase *biz.AddressUsecase
}

// NewUserService 是 UserService 的构造函数。
func NewUserService(userUC *biz.UserUsecase, addressUC *biz.AddressUsecase) *UserService {
	return &UserService{
		userUsecase:    userUC,
		addressUsecase: addressUC,
	}
}
