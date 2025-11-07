package service

import (
	"context"
	"time"

	v1 "ecommerce/api/user/v1"
	"ecommerce/internal/user/model"
	"ecommerce/internal/user/repository"
	"ecommerce/pkg/hash"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	v1.UnimplementedUserServer
	userRepo    repository.UserRepo
	addressRepo repository.AddressRepo
	jwtSecret   string
	jwtIssuer   string
	jwtExpire   time.Duration
}

func NewUserService(userRepo repository.UserRepo, addressRepo repository.AddressRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *UserService {
	return &UserService{
		userRepo:    userRepo,
		addressRepo: addressRepo,
		jwtSecret:   jwtSecret,
		jwtIssuer:   jwtIssuer,
		jwtExpire:   jwtExpire,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
		Nickname: req.Nickname,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		zap.S().Errorf("failed to create user: %v", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	return &v1.CreateUserResponse{
		UserId: user.ID,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, req *v1.GetUserByIDRequest) (*v1.GetUserByIDResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &v1.GetUserByIDResponse{
		User: &v1.UserInfo{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
		},
	}, nil
}
