package service

import (
	"context"
	"errors"
	"strconv"

	v1 "ecommerce/api/auth/v1"
	"ecommerce/internal/auth/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthService is the gRPC service implementation for authentication.
type AuthService struct {
	v1.UnimplementedAuthServiceServer
	uc *biz.AuthUsecase
}

// NewAuthService creates a new AuthService.
func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

// Login implements the Login RPC.
func (s *AuthService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	accessToken, refreshToken, expiresIn, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid username or password")
		}
		return nil, status.Errorf(codes.Internal, "failed to login: %v", err)
	}

	return &v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// ValidateToken implements the ValidateToken RPC.
func (s *AuthService) ValidateToken(ctx context.Context, req *v1.ValidateTokenRequest) (*v1.ValidateTokenResponse, error) {
	if req.AccessToken == "" {
		return nil, status.Error(codes.InvalidArgument, "access_token is required")
	}

	isValid, userID, username, err := s.uc.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return nil, status.Errorf(codes.Internal, "failed to validate token: %v", err)
	}

	return &v1.ValidateTokenResponse{
		IsValid:  isValid,
		UserId:   strconv.FormatUint(userID, 10), // Convert uint64 to string for proto
		Username: username,
	}, nil
}
