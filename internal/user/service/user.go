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
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "Cannot get metadata from context")
	}
	values := md.Get("x-user-id")
	if len(values) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "Missing x-user-id in request header")
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "Invalid x-user-id format")
	}
	return userID, nil
}

// bizUserToProto 将 biz.User 领域模型转换为 v1.UserInfo API 模型。
func bizUserToProto(user *biz.User) *v1.UserInfo {
	if user == nil {
		return nil
	}
	return &v1.UserInfo{
		UserId:   user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		Birthday: timestamppb.New(user.Birthday),
	}
}

// RegisterByPassword 实现了 user.proto 中定义的 RegisterByPassword RPC。
func (s *UserService) RegisterByPassword(ctx context.Context, req *v1.RegisterByPasswordRequest) (*v1.RegisterResponse, error) {
	user, err := s.userUsecase.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		// A more robust way to check for specific errors will be implemented later.
		if errors.Is(err, errors.New("username already exists")) {
			return nil, status.Errorf(codes.AlreadyExists, "Username already exists: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "Failed to register: %v", err)
	}

	return &v1.RegisterResponse{
		UserId: user.ID,
	}, nil
}

// LoginByPassword 实现了 user.proto 中定义的 LoginByPassword RPC。
func (s *UserService) LoginByPassword(ctx context.Context, req *v1.LoginByPasswordRequest) (*v1.LoginByPasswordResponse, error) {
	user, err := s.userUsecase.VerifyPassword(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Incorrect username or password")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpire, jwt.SigningMethodHS256)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token: %v", err)
	}

	claims, err := jwt.ParseToken(token, s.jwtSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to parse token: %v", err)
	}

	return &v1.LoginByPasswordResponse{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}

// VerifyPassword 实现了 user.proto 中定义的 VerifyPassword RPC (供内部服务调用)。
func (s *UserService) VerifyPassword(ctx context.Context, req *v1.VerifyPasswordRequest) (*v1.VerifyPasswordResponse, error) {
    user, err := s.userUsecase.VerifyPassword(ctx, req.Username, req.Password)
    if err != nil {
        return &v1.VerifyPasswordResponse{Success: false}, nil
    }

    return &v1.VerifyPasswordResponse{
        Success: true,
        User:    bizUserToProto(user),
    }, nil
}


// GetUserByID 实现了 user.proto 中定义的 GetUserByID RPC。
func (s *UserService) GetUserByID(ctx context.Context, req *v1.GetUserByIDRequest) (*v1.UserResponse, error) {
	user, err := s.userUsecase.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "User not found: %v", err)
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

	bizUser := &biz.User{
		ID: userID,
	}
	if req.HasNickname() {
		bizUser.Nickname = req.GetNickname()
	}
	if req.HasAvatar() {
		bizUser.Avatar = req.GetAvatar()
	}
	if req.HasGender() {
		bizUser.Gender = req.GetGender()
	}
	if req.HasBirthday() {
		bizUser.Birthday = req.GetBirthday().AsTime()
	}

	updatedUser, err := s.userUsecase.UpdateUser(ctx, bizUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update user info: %v", err)
	}

	return &v1.UserResponse{
		User: bizUserToProto(updatedUser),
	}, nil
}