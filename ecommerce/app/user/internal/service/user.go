package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "ecommerce/ecommerce/api/user/v1"
	"ecommerce/ecommerce/app/user/internal/biz"
)

// UserService 是 gRPC 服务的实现
type UserService struct {
	v1.UnimplementedUserServer // 必须嵌入，以保证向前兼容

	uc *biz.UserUsecase
}

// NewUserService 创建一个新的 UserService
func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

func (s *UserService) LoginByPassword(ctx context.Context, req *v1.LoginByPasswordRequest) (*v1.LoginByPasswordResponse, error) {
	// 1. 基本参数校验
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username or password cannot be empty")
	}

	// 2. 调用业务逻辑层
	token, expiresAt, err := s.uc.LoginByPassword(ctx, req.Username, req.Password)
	if err != nil {
		// Biz 层返回的业务错误，我们转换为 InvalidArgument
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 3. 组装并返回响应
	return &v1.LoginByPasswordResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// RegisterByPassword 实现了 gRPC 的注册接口
func (s *UserService) RegisterByPassword(ctx context.Context, req *v1.RegisterByPasswordRequest) (*v1.RegisterResponse, error) {
	// 1. 基本参数校验
	if req.Username == "" || len(req.Password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "invalid username or password")
	}

	// 2. 调用业务逻辑层
	userID, err := s.uc.CreateUserByPassword(ctx, req.Username, req.Password)
	if err != nil {
		// 在这里可以根据 biz 层返回的错误类型，转换为不同的 gRPC 状态码
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 3. 组装并返回响应
	return &v1.RegisterResponse{
		UserId: uint64(userID),
	}, nil
}
