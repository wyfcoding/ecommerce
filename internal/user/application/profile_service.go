package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/user/domain"
)

// ProfileService 定义了 Profile 相关的服务逻辑。
type ProfileService struct {
	userRepo domain.UserRepository
	logger   *slog.Logger
}

// NewProfileService 创建 Profile 服务实例。
func NewProfileService(userRepo domain.UserRepository, logger *slog.Logger) *ProfileService {
	return &ProfileService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUser 获取用户信息。
func (s *ProfileService) GetUser(ctx context.Context, userID uint64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, uint(userID))
}

// UpdateProfile 更新用户个人资料。
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uint64, nickname, avatar string, gender int8, birthday *time.Time) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.UpdateProfile(nickname, avatar, gender, birthday)

	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to update profile", "user_id", userID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "user profile updated successfully", "user_id", userID)

	return user, nil
}
