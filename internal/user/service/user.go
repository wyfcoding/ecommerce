package service

import (
	"context"
	"errors"
	"strconv"

	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/biz"
	"ecommerce/pkg/jwt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

// bizUserToProto 将 biz.User 领域模型转换为 v1.UserInfo API 模型。
func bizUserToProto(user *biz.User) *v1.UserInfo {
	if user == nil {
		return nil
	}
	res := &v1.UserInfo{
		UserId:   user.UserID,
		Username: user.Username,
	}
	if user.Nickname != nil {
		res.Nickname = *user.Nickname
	}
	if user.Avatar != nil {
		res.Avatar = *user.Avatar
	}
	if user.Gender != nil {
		res.Gender = *user.Gender
	}
	return res
}

// RegisterByPassword 实现了 user.proto 中定义的 RegisterByPassword RPC。
func (s *UserService) RegisterByPassword(ctx context.Context, req *v1.RegisterByPasswordRequest) (*v1.RegisterResponse, error) {
	user, err := s.userUsecase.Register(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrUserAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "用户已存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "注册失败: %v", err)
	}

	return &v1.RegisterResponse{
		UserId: user.UserID,
	}, nil
}

// LoginByPassword 实现了 user.proto 中定义的 LoginByPassword RPC。
func (s *UserService) LoginByPassword(ctx context.Context, req *v1.LoginByPasswordRequest) (*v1.LoginByPasswordResponse, error) {
	token, err := s.userUsecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
		}
		if errors.Is(err, biz.ErrPasswordIncorrect) {
			return nil, status.Errorf(codes.Unauthenticated, "密码错误: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "登录失败: %v", err)
	}

	// 从 JWT token 中解析 expires_at
	claims, err := jwt.ParseToken(token, s.userUsecase.GetJwtSecret())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "解析 JWT 失败: %v", err)
	}

	return &v1.LoginByPasswordResponse{
		Token:    token,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

// GetUserByID 实现了 user.proto 中定义的 GetUserByID RPC。
func (s *UserService) GetUserByID(ctx context.Context, req *v1.GetUserByIDRequest) (*v1.UserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userUsecase.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "获取用户信息失败: %v", err)
	}

	return &v1.UserResponse{
		User: bizUserToProto(user),
	}, nil
}

// UpdateUserInfo 实现了 user.proto 中定义的 UpdateUserInfo RPC。
func (s *UserService) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UserResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 将 gRPC 请求模型转换为 biz 领域模型。
	bizUser := &biz.User{
		UserID: userID,
	}
	if req.HasNickname() {
		nickname := req.GetNickname()
		bizUser.Nickname = &nickname
	}
	if req.HasAvatar() {
		avatar := req.GetAvatar()
		bizUser.Avatar = &avatar
	}
	if req.HasGender() {
		gender := req.GetGender()
		bizUser.Gender = &gender
	}

	updatedUser, err := s.userUsecase.UpdateUserInfo(ctx, bizUser)
	if err != nil {
		if errors.Is(err, biz.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "更新用户信息失败: %v", err)
	}

	return &v1.UserResponse{
		User: bizUserToProto(updatedUser),
	}, nil
}